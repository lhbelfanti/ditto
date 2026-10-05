package broker

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type (
	// MockBroker is a MessageBroker whose EnqueueMessage returns EnqueueErr and whose other methods do nothing.
	MockBroker struct {
		EnqueueErr error
	}

	// MockChannel is an in-memory channel that records the QoS, subscription and published message, and returns the configured errors.
	MockChannel struct {
		QosErr       error
		QosCount     int
		QueueErr     error
		PublishErr   error
		ConsumeErr   error
		ConsumeCalls int
		QosAtConsume int
		Messages     <-chan amqp091.Delivery
		Published    amqp091.Publishing
	}

	// MockAckRecorder is an amqp091.Acknowledger that records the delivery tags it acknowledges and rejects.
	MockAckRecorder struct {
		Acks  chan uint64
		Nacks chan uint64
	}
)

// MockEnqueue returns a MessageBroker whose EnqueueMessage always returns err.
func MockEnqueue(err error) MessageBroker {
	return &MockBroker{EnqueueErr: err}
}

// MockRabbitMQBroker builds a RabbitMQBroker over channel without dialing RabbitMQ. A nil channel builds a producer-only broker.
func MockRabbitMQBroker(channel *MockChannel, queueName string, consumer bool, messages <-chan amqp091.Delivery) *RabbitMQBroker {
	b := &RabbitMQBroker{queue: amqp091.Queue{Name: queueName}, consumer: consumer, messages: messages}
	if channel != nil {
		b.channel = channel
	}
	return b
}

// EnqueueMessage returns the configured EnqueueErr without publishing anything.
func (m *MockBroker) EnqueueMessage(_ context.Context, _ string) error {
	return m.EnqueueErr
}

// InitMessageConsumerWithFunction returns nil without consuming anything.
func (m *MockBroker) InitMessageConsumerWithFunction(_ int, _ ProcessorFunction) error { return nil }

// CloseConnection does nothing.
func (m *MockBroker) CloseConnection() {}

// Qos records count as QosCount and returns QosErr.
func (m *MockChannel) Qos(count, _ int, _ bool) error {
	m.QosCount = count
	return m.QosErr
}

// Consume counts the call, records the QoS in effect as QosAtConsume, and returns Messages and ConsumeErr.
func (m *MockChannel) Consume(_, _ string, _, _, _, _ bool, _ amqp091.Table) (<-chan amqp091.Delivery, error) {
	m.ConsumeCalls++
	m.QosAtConsume = m.QosCount
	return m.Messages, m.ConsumeErr
}

// QueueDeclare returns a queue named name, or QueueErr when it is set.
func (m *MockChannel) QueueDeclare(name string, _, _, _, _ bool, _ amqp091.Table) (amqp091.Queue, error) {
	if m.QueueErr != nil {
		return amqp091.Queue{}, m.QueueErr
	}
	return amqp091.Queue{Name: name}, nil
}

// PublishWithContext records msg as Published and returns PublishErr.
func (m *MockChannel) PublishWithContext(_ context.Context, _, _ string, _, _ bool, msg amqp091.Publishing) error {
	m.Published = msg
	return m.PublishErr
}

// Close does nothing and returns nil.
func (m *MockChannel) Close() error { return nil }

// MockAcknowledger returns a MockAckRecorder with room for ackSize acknowledgements and nackSize rejections.
func MockAcknowledger(ackSize, nackSize int) *MockAckRecorder {
	return &MockAckRecorder{
		Acks:  make(chan uint64, ackSize),
		Nacks: make(chan uint64, nackSize),
	}
}

// Ack records tag in Acks.
func (m *MockAckRecorder) Ack(tag uint64, _ bool) error {
	m.Acks <- tag
	return nil
}

// Nack records tag in Nacks.
func (m *MockAckRecorder) Nack(tag uint64, _, _ bool) error {
	m.Nacks <- tag
	return nil
}

// Reject records tag in Nacks.
func (m *MockAckRecorder) Reject(tag uint64, _ bool) error {
	m.Nacks <- tag
	return nil
}

// MockProcessor returns a ProcessorFunction that always returns err.
func MockProcessor(err error) ProcessorFunction {
	return func(context.Context, []byte) error { return err }
}

// MockRecordingProcessor returns a ProcessorFunction that sends every body it receives to bodies and returns err.
func MockRecordingProcessor(bodies chan<- []byte, err error) ProcessorFunction {
	return func(_ context.Context, body []byte) error {
		bodies <- body
		return err
	}
}

// MockBlockingProcessor returns a ProcessorFunction that signals started, waits for release to close, and returns err.
func MockBlockingProcessor(started chan<- struct{}, release <-chan struct{}, err error) ProcessorFunction {
	return func(context.Context, []byte) error {
		started <- struct{}{}
		<-release
		return err
	}
}

// MockClosedMessages returns a delivery channel already holding deliveries and already closed, as if the connection dropped.
func MockClosedMessages(deliveries ...amqp091.Delivery) <-chan amqp091.Delivery {
	messages := make(chan amqp091.Delivery, len(deliveries))
	for _, d := range deliveries {
		messages <- d
	}
	close(messages)
	return messages
}
