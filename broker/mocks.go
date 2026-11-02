package broker

import (
	"context"
	"os"
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

type (
	// MockBroker is a MessageBroker whose EnqueueMessage returns EnqueueErr and whose other methods do nothing.
	MockBroker struct {
		EnqueueErr error
	}

	// MockChannel is an in-memory channel that returns the configured errors and messages, and records
	// its calls in Capture when it is set.
	MockChannel struct {
		QosErr     error
		QueueErr   error
		PublishErr error
		ConsumeErr error
		Messages   <-chan amqp091.Delivery
		Capture    *ChannelCapture
	}

	// ChannelCapture holds what a MockChannel was asked to do: the QoS, the subscriptions and the published message.
	ChannelCapture struct {
		QosCount     int
		ConsumeCalls int
		QosAtConsume int
		Published    amqp091.Publishing
	}

	// AckCapture is an amqp091.Acknowledger that records the delivery tags it acknowledges and rejects.
	AckCapture struct {
		Acks  chan uint64
		Nacks chan uint64
	}

	// Acks holds the delivery tags a AckCapture acknowledged and rejected.
	Acks struct {
		Acks  []uint64
		Nacks []uint64
	}
)

// MockEnqueue returns a MessageBroker whose EnqueueMessage always returns err.
func MockEnqueue(err error) MessageBroker {
	return &MockBroker{EnqueueErr: err}
}

// MockRabbitMQBroker builds a RabbitMQBroker over channel without dialing RabbitMQ. A nil channel builds a producer-only broker.
func MockRabbitMQBroker(channel *MockChannel, queueName string, consumer bool, messages <-chan amqp091.Delivery) *RabbitMQBroker {
	b := &RabbitMQBroker{
		queue:    amqp091.Queue{Name: queueName},
		consumer: consumer,
		messages: messages,
	}
	if channel != nil {
		b.channel = channel
	}

	return b
}

// EnqueueMessage returns the configured EnqueueErr without publishing anything.
func (m *MockBroker) EnqueueMessage(_ context.Context, _ string) error {
	return m.EnqueueErr
}

// InitMessageConsumer returns nil without consuming anything.
func (m *MockBroker) InitMessageConsumer(_ int, _ Processor) error { return nil }

// CloseConnection does nothing.
func (m *MockBroker) CloseConnection() {}

// capture returns Capture, or a throwaway one when it is not set.
func (m *MockChannel) capture() *ChannelCapture {
	if m.Capture == nil {
		return &ChannelCapture{}
	}

	return m.Capture
}

// Qos records count as QosCount and returns QosErr.
func (m *MockChannel) Qos(count, _ int, _ bool) error {
	m.capture().QosCount = count
	return m.QosErr
}

// Consume counts the call, records the QoS in effect as QosAtConsume, and returns Messages and ConsumeErr.
func (m *MockChannel) Consume(_, _ string, _, _, _, _ bool, _ amqp091.Table) (<-chan amqp091.Delivery, error) {
	capture := m.capture()
	capture.ConsumeCalls++
	capture.QosAtConsume = capture.QosCount
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
	m.capture().Published = msg
	return m.PublishErr
}

// Close does nothing and returns nil.
func (m *MockChannel) Close() error { return nil }

// MockAcknowledger returns an AckCapture for a delivery whose acknowledgements the test does not read.
func MockAcknowledger() *AckCapture {
	return CaptureAcks(1, 1)
}

// CaptureAcks returns a AckCapture with room for ackSize acknowledgements and nackSize rejections.
func CaptureAcks(ackSize, nackSize int) *AckCapture {
	return &AckCapture{
		Acks:  make(chan uint64, ackSize),
		Nacks: make(chan uint64, nackSize),
	}
}

// Ack records tag in Acks.
func (m *AckCapture) Ack(tag uint64, _ bool) error {
	m.Acks <- tag
	return nil
}

// Nack records tag in Nacks.
func (m *AckCapture) Nack(tag uint64, _, _ bool) error {
	m.Nacks <- tag
	return nil
}

// Reject records tag in Nacks.
func (m *AckCapture) Reject(tag uint64, _ bool) error {
	m.Nacks <- tag
	return nil
}

// Recorded drains the tags acknowledged and rejected so far.
func (m *AckCapture) Recorded() Acks {
	recorded := Acks{}
	for len(m.Acks) > 0 {
		recorded.Acks = append(recorded.Acks, <-m.Acks)
	}

	for len(m.Nacks) > 0 {
		recorded.Nacks = append(recorded.Nacks, <-m.Nacks)
	}

	return recorded
}

// MockProcessor returns a Processor that always returns err.
func MockProcessor(err error) Processor {
	return func(context.Context, []byte) error { return err }
}

// MockRecordingProcessor returns a Processor that sends every body it receives to bodies and returns err.
func MockRecordingProcessor(bodies chan<- []byte, err error) Processor {
	return func(_ context.Context, body []byte) error {
		bodies <- body
		return err
	}
}

// MockBlockingProcessor returns a Processor that signals started, waits for release to close, and returns err.
func MockBlockingProcessor(started chan<- struct{}, release <-chan struct{}, err error) Processor {
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

// MockConnection returns a connection to the RabbitMQ server in DITTO_TEST_RABBITMQ_URL, closed when the test ends, and skips the test when the variable is unset.
func MockConnection(t *testing.T) *amqp091.Connection {
	t.Helper()
	url := os.Getenv("DITTO_TEST_RABBITMQ_URL")
	if url == "" {
		t.Skip("DITTO_TEST_RABBITMQ_URL not set")
	}

	conn, err := amqp091.Dial(url)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = conn.Close() })

	return conn
}
