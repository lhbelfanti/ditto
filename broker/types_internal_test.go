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
	acknowledger := MockAcknowledger(1, 1)
	b := &RabbitMQBroker{processor: MockProcessor(nil), semaphore: make(chan struct{}, 1)}
	b.semaphore <- struct{}{}
	b.inFlight.Add(1)

	b.processDelivery(amqp091.Delivery{DeliveryTag: 7, Acknowledger: acknowledger})

	want := uint64(7)
	got := <-acknowledger.Acks

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_processDelivery_successWhenSemaphoreIsReleased(t *testing.T) {
	acknowledger := MockAcknowledger(1, 1)
	b := &RabbitMQBroker{processor: MockProcessor(nil), semaphore: make(chan struct{}, 1)}
	b.semaphore <- struct{}{}
	b.inFlight.Add(1)

	b.processDelivery(amqp091.Delivery{DeliveryTag: 1, Acknowledger: acknowledger})

	want := 0
	got := len(b.semaphore)

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_processDelivery_failsWhenProcessorFails(t *testing.T) {
	acknowledger := MockAcknowledger(1, 1)
	b := &RabbitMQBroker{processor: MockProcessor(errors.New("processing failed")), semaphore: make(chan struct{}, 1)}
	b.semaphore <- struct{}{}
	b.inFlight.Add(1)

	b.processDelivery(amqp091.Delivery{DeliveryTag: 9, Acknowledger: acknowledger})

	want := uint64(9)
	got := <-acknowledger.Nacks

	assert.Equal(t, want, got)
}
