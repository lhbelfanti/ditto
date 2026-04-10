package broker

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// NewProducer creates a new RabbitMQBroker configured for producing messages.
func NewProducer(ctx context.Context, url, queueName string) (*RabbitMQBroker, error) {
	conn, err := dial(url)
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

// CloseConnection closes the broker connection.
func (b *RabbitMQBroker) CloseConnection() {
	if b.conn != nil {
		b.conn.Close()
	}
}
