package broker

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type (
	mockBroker struct {
		enqueueErr error
	}

	mockChannel struct {
		qosErr       error
		qosCount     int
		queueErr     error
		publishErr   error
		consumeErr   error
		consumeCalls int
		qosAtConsume int
		messages     <-chan amqp091.Delivery
		published    amqp091.Publishing
	}

	mockAcknowledger struct {
		acks  chan uint64
		nacks chan uint64
	}
)

func MockEnqueue(err error) MessageBroker {
	return &mockBroker{enqueueErr: err}
}

func (m *mockBroker) EnqueueMessage(_ context.Context, _ string) error {
	return m.enqueueErr
}

func (m *mockBroker) InitMessageConsumerWithFunction(_ int, _ ProcessorFunction) {}

func (m *mockBroker) CloseConnection() {}

func (m *mockChannel) Qos(count, _ int, _ bool) error {
	m.qosCount = count
	return m.qosErr
}

func (m *mockChannel) Consume(_, _ string, _, _, _, _ bool, _ amqp091.Table) (<-chan amqp091.Delivery, error) {
	m.consumeCalls++
	m.qosAtConsume = m.qosCount
	return m.messages, m.consumeErr
}

func (m *mockChannel) QueueDeclare(name string, _, _, _, _ bool, _ amqp091.Table) (amqp091.Queue, error) {
	if m.queueErr != nil {
		return amqp091.Queue{}, m.queueErr
	}
	return amqp091.Queue{Name: name}, nil
}

func (m *mockChannel) PublishWithContext(_ context.Context, _, _ string, _, _ bool, msg amqp091.Publishing) error {
	m.published = msg
	return m.publishErr
}

func (m *mockChannel) Close() error { return nil }

func MockAcknowledger(ackSize, nackSize int) *mockAcknowledger {
	return &mockAcknowledger{
		acks:  make(chan uint64, ackSize),
		nacks: make(chan uint64, nackSize),
	}
}

func (m *mockAcknowledger) Ack(tag uint64, _ bool) error {
	m.acks <- tag
	return nil
}

func (m *mockAcknowledger) Nack(tag uint64, _, _ bool) error {
	m.nacks <- tag
	return nil
}

func (m *mockAcknowledger) Reject(tag uint64, _ bool) error {
	m.nacks <- tag
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
