package broker

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

// MessageBroker defines the contract for interacting with the message broker.
type MessageBroker interface {
	EnqueueMessage(ctx context.Context, body string) error
	InitMessageConsumerWithFunction(concurrentMessages int, processorFunc ProcessorFunction)
	CloseConnection()
}

// ProcessorFunction is a function that processes a message body.
type ProcessorFunction func(ctx context.Context, body []byte) error

// RabbitMQBroker is an implementation of MessageBroker using RabbitMQ.
type RabbitMQBroker struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	queue    amqp091.Queue
	messages <-chan amqp091.Delivery
}
