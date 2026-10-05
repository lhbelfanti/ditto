package broker

import (
	"context"
	"sync"

	"github.com/rabbitmq/amqp091-go"
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

func makeProcessDelivery(processor Processor, semaphore chan struct{}, inFlight *sync.WaitGroup) func(amqp091.Delivery) {
	return func(d amqp091.Delivery) {
		defer inFlight.Done()
		defer func() { <-semaphore }()

		ctx := context.Background()
		err := processor(ctx, d.Body)
		if err != nil {
			_ = d.Nack(false, false)
			return
		}
		_ = d.Ack(false)
	}
}
