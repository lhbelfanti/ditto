package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// drainTimeout bounds how long CloseConnection waits for in-flight message-processing goroutines
// (spawned by InitMessageConsumerWithFunction) to finish before closing the connection regardless.
const drainTimeout = 10 * time.Second

// NewProducer creates a new RabbitMQBroker configured for producing messages.
func NewProducer(ctx context.Context, url, queueName string) (*RabbitMQBroker, error) {
	conn, err := dial(ctx, url)
	if err != nil {
		return nil, err
	}

	ch, err := openChannel(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	q, err := declareQueue(ch, queueName)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQBroker{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

// EnqueueMessage publishes a message to the broker.
func (b *RabbitMQBroker) EnqueueMessage(ctx context.Context, body string) error {
	err := b.channel.PublishWithContext(ctx,
		"",           // exchange
		b.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		},
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToPublishMessage, err)
	}
	return nil
}

// CloseConnection waits up to drainTimeout for any in-flight message-processing goroutines to
// finish, then closes the broker connection regardless. A producer-only broker (no goroutines
// ever spawned) closes immediately, since inFlight is already at zero.
func (b *RabbitMQBroker) CloseConnection() {
	b.stopChannel()
	b.shutdownOnce.Do(func() {
		b.dispatchMu.Lock()
		b.closing = true
		close(b.stop)
		b.dispatchMu.Unlock()
	})

	done := make(chan struct{})
	go func() {
		b.inFlight.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(drainTimeout):
	}

	if b.conn != nil {
		b.conn.Close()
	}
}
