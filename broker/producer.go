package broker

import (
	"context"
	"fmt"
)

// NewProducer creates a new RabbitMQBroker configured for producing messages.
func NewProducer(ctx context.Context, url, queueName string) (*RabbitMQBroker, error) {
	conn, err := dial(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCantCreateProducer, err)
	}

	ch, err := openChannel(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("%w: %w", ErrCantCreateProducer, err)
	}

	q, err := declareQueue(ch, queueName)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("%w: %w", ErrCantCreateProducer, err)
	}

	return &RabbitMQBroker{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}
