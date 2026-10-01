package broker

import (
	"context"
	"fmt"
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

// InitMessageConsumerWithFunction subscribes and starts processing messages, bounding both
// RabbitMQ's own delivery rate (via channel QoS) and the number of concurrently running
// processorFunc goroutines to concurrentMessages. It blocks for the consumer's lifetime — call it
// in its own goroutine — returning nil only once stopped deliberately via CloseConnection, or a
// non-nil error otherwise: immediately if setup (QoS, subscribing) fails, or
// ErrConsumerChannelClosed if the delivery channel closes on its own (e.g. a dropped connection)
// before CloseConnection ever ran — a disconnect never reads as a clean shutdown. It never logs
// itself; a caller that needs this failure visible decides how, the same as database.Check.
// Calling this on a broker built by NewProducer (no messages channel) returns ErrNotAConsumer
// instead of blocking forever on a nil channel.
func (b *RabbitMQBroker) InitMessageConsumerWithFunction(concurrentMessages int, processorFunc ProcessorFunction) error {
	if !b.consumer && b.messages == nil {
		return ErrNotAConsumer
	}
	if concurrentMessages <= 0 {
		return ErrInvalidConcurrency
	}

	err := b.channel.Qos(concurrentMessages, 0, false)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToSetQoS, err)
	}
	if b.messages == nil {
		b.messages, err = b.channel.Consume(
			b.queue.Name, // queue
			"",           // consumer
			false,        // auto-ack
			false,        // exclusive
			false,        // no-local
			false,        // no-wait
			nil,          // args
		)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToConsumeQueue, err)
		}
	}

	semaphore := make(chan struct{}, concurrentMessages)
	stop := b.stopChannel()
	process := makeProcessDelivery(processorFunc, semaphore, &b.inFlight)

	for {
		var msg amqp091.Delivery
		select {
		case <-stop:
			return nil
		case received, ok := <-b.messages:
			if !ok {
				select {
				case <-stop:
					return nil
				default:
					return ErrConsumerChannelClosed
				}
			}
			msg = received
		}

		select {
		case semaphore <- struct{}{}:
		case <-stop:
			return nil
		}

		b.dispatchMu.Lock()
		if b.closing {
			b.dispatchMu.Unlock()
			<-semaphore
			return nil
		}
		b.inFlight.Add(1)
		b.dispatchMu.Unlock()

		go process(msg)
	}
}

func makeProcessDelivery(processor ProcessorFunction, semaphore chan struct{}, inFlight *sync.WaitGroup) func(amqp091.Delivery) {
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
