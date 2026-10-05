package broker

import (
	"context"
	"time"
)

// drainTimeout bounds how long CloseConnection waits for in-flight message-processing goroutines
// (spawned by InitMessageConsumerWithFunction) to finish before closing the connection regardless.
const drainTimeout time.Duration = 10 * time.Second

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
