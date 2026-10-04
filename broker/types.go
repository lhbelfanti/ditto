package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type (
	// MessageBroker defines the contract for interacting with the message broker.
	MessageBroker interface {
		EnqueueMessage(ctx context.Context, body string) error
		InitMessageConsumerWithFunction(concurrentMessages int, processorFunc ProcessorFunction) error
		CloseConnection()
	}

	// ProcessorFunction is a function that processes a message body.
	ProcessorFunction func(ctx context.Context, body []byte) error

	channel interface {
		Qos(prefetchCount, prefetchSize int, global bool) error
		Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error)
		QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error)
		PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error
		Close() error
	}

	// RabbitMQBroker is an implementation of MessageBroker using RabbitMQ.
	RabbitMQBroker struct {
		conn         *amqp091.Connection
		channel      channel
		queue        amqp091.Queue
		messages     <-chan amqp091.Delivery
		consumer     bool
		stopOnce     sync.Once
		stop         chan struct{}
		shutdownOnce sync.Once
		dispatchMu   sync.Mutex
		closing      bool

		// inFlight tracks message-processing goroutines spawned by InitMessageConsumerWithFunction,
		// so CloseConnection can give them a bounded chance to finish before tearing the connection
		// down. Zero-value on a producer-only broker, where it is never touched.
		inFlight sync.WaitGroup
	}

	dialResult struct {
		conn *amqp091.Connection
		err  error
	}
)

func (b *RabbitMQBroker) stopChannel() <-chan struct{} {
	b.stopOnce.Do(func() { b.stop = make(chan struct{}) })
	return b.stop
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
		return fmt.Errorf("%w: %w", ErrFailedToPublishMessage, err)
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
