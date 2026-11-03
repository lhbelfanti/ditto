package broker

import (
	"errors"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestRabbitMQBroker_stopChannel_success(t *testing.T) {
	b := &RabbitMQBroker{}

	first := b.stopChannel()
	second := b.stopChannel()

	want := first
	got := second

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_processDelivery_success(t *testing.T) {
	captureAcks := CaptureAcks(1, 1)
	b := &RabbitMQBroker{processor: MockProcessor(nil), semaphore: make(chan struct{}, 1)}
	b.semaphore <- struct{}{}
	b.inFlight.Add(1)

	b.processDelivery(amqp091.Delivery{DeliveryTag: 7, Acknowledger: captureAcks})

	want := uint64(7)
	got := <-captureAcks.Acks

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_processDelivery_successWhenBodyIsProcessed(t *testing.T) {
	mockAck := MockAcknowledger(1)
	bodies := make(chan []byte, 1)
	b := &RabbitMQBroker{processor: MockRecordingProcessor(bodies, nil), semaphore: make(chan struct{}, 1)}
	b.semaphore <- struct{}{}
	b.inFlight.Add(1)

	b.processDelivery(amqp091.Delivery{DeliveryTag: 1, Acknowledger: mockAck, Body: []byte("hello")})

	want := []byte("hello")
	got := <-bodies

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_processDelivery_successWhenSemaphoreIsReleased(t *testing.T) {
	mockAck := MockAcknowledger(1)
	b := &RabbitMQBroker{processor: MockProcessor(nil), semaphore: make(chan struct{}, 1)}
	b.semaphore <- struct{}{}
	b.inFlight.Add(1)

	b.processDelivery(amqp091.Delivery{DeliveryTag: 1, Acknowledger: mockAck})

	want := 0
	got := len(b.semaphore)

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_processDelivery_failsWhenProcessorFails(t *testing.T) {
	captureAcks := CaptureAcks(1, 1)
	b := &RabbitMQBroker{processor: MockProcessor(errors.New("processing failed")), semaphore: make(chan struct{}, 1)}
	b.semaphore <- struct{}{}
	b.inFlight.Add(1)

	b.processDelivery(amqp091.Delivery{DeliveryTag: 9, Acknowledger: captureAcks})

	want := uint64(9)
	got := <-captureAcks.Nacks

	assert.Equal(t, want, got)
}
