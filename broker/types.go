package broker

import (
	"context"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

type (
	// MessageBroker defines the contract for interacting with the message broker.
	MessageBroker interface {
		EnqueueMessage(ctx context.Context, body string) error
		InitMessageConsumerWithFunction(concurrentMessages int, processorFunc ProcessorFunction)
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
)

func (b *RabbitMQBroker) stopChannel() <-chan struct{} {
	b.stopOnce.Do(func() { b.stop = make(chan struct{}) })
	return b.stop
}
