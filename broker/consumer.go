package broker

import (
	"context"
	"fmt"
)

// NewConsumer creates a new RabbitMQBroker configured for consuming messages.
func NewConsumer(ctx context.Context, url, queueName string) (*RabbitMQBroker, error) {
	conn, err := dial(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCantCreateConsumer, err)
	}

	ch, err := openChannel(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("%w: %w", ErrCantCreateConsumer, err)
	}

	q, err := declareQueue(ch, queueName)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("%w: %w", ErrCantCreateConsumer, err)
	}

	return &RabbitMQBroker{
		conn:     conn,
		channel:  ch,
		queue:    q,
		consumer: true,
	}, nil
}
