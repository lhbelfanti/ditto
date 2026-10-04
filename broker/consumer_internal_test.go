package broker

import (
	"errors"
	"sync"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestMakeProcessDelivery_success(t *testing.T) {
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}
	inFlight := &sync.WaitGroup{}
	inFlight.Add(1)
	acknowledger := MockAcknowledger(1, 1)
	processDelivery := makeProcessDelivery(MockProcessor(nil), semaphore, inFlight)

	processDelivery(amqp091.Delivery{DeliveryTag: 7, Acknowledger: acknowledger})
	inFlight.Wait()

	want := uint64(7)
	got := <-acknowledger.Acks

	assert.Equal(t, want, got)
}

func TestMakeProcessDelivery_successWhenSemaphoreIsReleased(t *testing.T) {
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}
	inFlight := &sync.WaitGroup{}
	inFlight.Add(1)
	acknowledger := MockAcknowledger(1, 1)
	processDelivery := makeProcessDelivery(MockProcessor(nil), semaphore, inFlight)

	processDelivery(amqp091.Delivery{DeliveryTag: 1, Acknowledger: acknowledger})
	inFlight.Wait()

	want := 0
	got := len(semaphore)

	assert.Equal(t, want, got)
}

func TestMakeProcessDelivery_failsWhenProcessorFails(t *testing.T) {
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}
	inFlight := &sync.WaitGroup{}
	inFlight.Add(1)
	acknowledger := MockAcknowledger(1, 1)
	processDelivery := makeProcessDelivery(MockProcessor(errors.New("processing failed")), semaphore, inFlight)

	processDelivery(amqp091.Delivery{DeliveryTag: 9, Acknowledger: acknowledger})
	inFlight.Wait()

	want := uint64(9)
	got := <-acknowledger.Nacks

	assert.Equal(t, want, got)
}
