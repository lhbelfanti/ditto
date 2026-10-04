package broker

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type (
	MockBroker struct {
		EnqueueErr error
	}

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

	MockAckRecorder struct {
		Acks  chan uint64
		Nacks chan uint64
	}
)

func MockEnqueue(err error) MessageBroker {
	return &MockBroker{EnqueueErr: err}
}

func MockRabbitMQBroker(channel *MockChannel, queueName string, consumer bool, messages <-chan amqp091.Delivery) *RabbitMQBroker {
	b := &RabbitMQBroker{queue: amqp091.Queue{Name: queueName}, consumer: consumer, messages: messages}
	if channel != nil {
		b.channel = channel
	}
	return b
}

func (m *MockBroker) EnqueueMessage(_ context.Context, _ string) error {
	return m.EnqueueErr
}

func (m *MockBroker) InitMessageConsumerWithFunction(_ int, _ ProcessorFunction) error { return nil }

func (m *MockBroker) CloseConnection() {}

func (m *MockChannel) Qos(count, _ int, _ bool) error {
	m.QosCount = count
	return m.QosErr
}

func (m *MockChannel) Consume(_, _ string, _, _, _, _ bool, _ amqp091.Table) (<-chan amqp091.Delivery, error) {
	m.ConsumeCalls++
	m.QosAtConsume = m.QosCount
	return m.Messages, m.ConsumeErr
}

func (m *MockChannel) QueueDeclare(name string, _, _, _, _ bool, _ amqp091.Table) (amqp091.Queue, error) {
	if m.QueueErr != nil {
		return amqp091.Queue{}, m.QueueErr
	}
	return amqp091.Queue{Name: name}, nil
}

func (m *MockChannel) PublishWithContext(_ context.Context, _, _ string, _, _ bool, msg amqp091.Publishing) error {
	m.Published = msg
	return m.PublishErr
}

func (m *MockChannel) Close() error { return nil }

func MockAcknowledger(ackSize, nackSize int) *MockAckRecorder {
	return &MockAckRecorder{
		Acks:  make(chan uint64, ackSize),
		Nacks: make(chan uint64, nackSize),
	}
}

func (m *MockAckRecorder) Ack(tag uint64, _ bool) error {
	m.Acks <- tag
	return nil
}

func (m *MockAckRecorder) Nack(tag uint64, _, _ bool) error {
	m.Nacks <- tag
	return nil
}

func (m *MockAckRecorder) Reject(tag uint64, _ bool) error {
	m.Nacks <- tag
	return nil
}

func MockProcessor(err error) ProcessorFunction {
	return func(context.Context, []byte) error { return err }
}

func MockRecordingProcessor(bodies chan<- []byte, err error) ProcessorFunction {
	return func(_ context.Context, body []byte) error {
		bodies <- body
		return err
	}
}

func MockBlockingProcessor(started chan<- struct{}, release <-chan struct{}, err error) ProcessorFunction {
	return func(context.Context, []byte) error {
		started <- struct{}{}
		<-release
		return err
	}
}

func MockClosedMessages(deliveries ...amqp091.Delivery) <-chan amqp091.Delivery {
	messages := make(chan amqp091.Delivery, len(deliveries))
	for _, d := range deliveries {
		messages <- d
	}
	close(messages)
	return messages
}
