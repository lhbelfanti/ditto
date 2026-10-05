package broker

import (
	"context"
)

// NewConsumer creates a new RabbitMQBroker configured for consuming messages.
func NewConsumer(ctx context.Context, url, queueName string) (*RabbitMQBroker, error) {
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
		conn:     conn,
		channel:  ch,
		queue:    q,
		consumer: true,
	}, nil
}
