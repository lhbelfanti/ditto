package broker

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// NewConsumer creates a new RabbitMQBroker configured for consuming messages.
func NewConsumer(ctx context.Context, url, queueName string) (*RabbitMQBroker, error) {
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

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("%w: %v", ErrFailedToConsumeQueue, err)
	}

	return &RabbitMQBroker{
		conn:     conn,
		channel:  ch,
		queue:    q,
		messages: msgs,
	}, nil
}

// InitMessageConsumerWithFunction initializes the message consumer and starts processing messages.
func (b *RabbitMQBroker) InitMessageConsumerWithFunction(concurrentMessages int, processorFunc ProcessorFunction) {
	// concurrentMessages is currently handled by the number of goroutines spawned.
	// In a more complex implementation, we would use a semaphore or prefetch count.
	for msg := range b.messages {
		go func(d amqp091.Delivery) {
			ctx := context.Background()
			if err := processorFunc(ctx, d.Body); err != nil {
				d.Nack(false, false)
				return
			}
			d.Ack(false)
		}(msg)
	}
}
