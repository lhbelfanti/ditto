package broker

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"github.com/lhbelfanti/ditto/v2/log"
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

// InitMessageConsumerWithFunction initializes the message consumer and starts processing
// messages, bounding both RabbitMQ's own delivery rate (via channel QoS) and the number of
// concurrently running processorFunc goroutines to concurrentMessages. Calling this on a broker
// built by NewProducer (no messages channel) logs and returns instead of blocking forever on a
// nil channel.
func (b *RabbitMQBroker) InitMessageConsumerWithFunction(concurrentMessages int, processorFunc ProcessorFunction) {
	if b.messages == nil {
		log.Error(context.Background(), ErrNotAConsumer.Error())
		return
	}
	if concurrentMessages <= 0 {
		log.Error(context.Background(), ErrInvalidConcurrency.Error())
		return
	}

	if err := b.channel.Qos(concurrentMessages, 0, false); err != nil {
		log.Err(context.Background(), err, ErrFailedToSetQoS.Error())
		return
	}

	semaphore := make(chan struct{}, concurrentMessages)
	stop := b.stopChannel()

	for {
		var msg amqp091.Delivery
		select {
		case <-stop:
			return
		case received, ok := <-b.messages:
			if !ok {
				return
			}
			msg = received
		}

		select {
		case semaphore <- struct{}{}:
		case <-stop:
			return
		}

		b.dispatchMu.Lock()
		if b.closing {
			b.dispatchMu.Unlock()
			<-semaphore
			return
		}
		b.inFlight.Add(1)
		b.dispatchMu.Unlock()

		go func(d amqp091.Delivery) {
			defer b.inFlight.Done()
			defer func() { <-semaphore }()

			ctx := context.Background()
			if err := processorFunc(ctx, d.Body); err != nil {
				_ = d.Nack(false, false)
				return
			}
			_ = d.Ack(false)
		}(msg)
	}
}
