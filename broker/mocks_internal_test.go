package broker

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type mockChannel struct {
	qosErr     error
	qosCount   int
	queueErr   error
	publishErr error
	consumeErr error
	messages   <-chan amqp091.Delivery
	published  amqp091.Publishing
}

func (m *mockChannel) Qos(count, _ int, _ bool) error {
	m.qosCount = count
	return m.qosErr
}

func (m *mockChannel) Consume(_, _ string, _, _, _, _ bool, _ amqp091.Table) (<-chan amqp091.Delivery, error) {
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

type mockAcknowledger struct {
	acks  chan uint64
	nacks chan uint64
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
