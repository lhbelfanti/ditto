package broker

import (
	"errors"
	"sync"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func setUpProcessDelivery(err error) (func(amqp091.Delivery), *mockAcknowledger, chan struct{}, *sync.WaitGroup) {
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}
	inFlight := &sync.WaitGroup{}
	inFlight.Add(1)
	acknowledger := MockAcknowledger(1, 1)
	processDelivery := makeProcessDelivery(MockProcessor(err), semaphore, inFlight)
	return processDelivery, acknowledger, semaphore, inFlight
}

func TestMakeProcessDelivery_success(t *testing.T) {
	processDelivery, acknowledger, _, inFlight := setUpProcessDelivery(nil)

	processDelivery(amqp091.Delivery{DeliveryTag: 7, Acknowledger: acknowledger})
	inFlight.Wait()

	want := uint64(7)
	got := <-acknowledger.acks

	assert.Equal(t, want, got)
}

func TestMakeProcessDelivery_successWhenSemaphoreIsReleased(t *testing.T) {
	processDelivery, acknowledger, semaphore, inFlight := setUpProcessDelivery(nil)

	processDelivery(amqp091.Delivery{DeliveryTag: 1, Acknowledger: acknowledger})
	inFlight.Wait()

	want := 0
	got := len(semaphore)

	assert.Equal(t, want, got)
}

func TestMakeProcessDelivery_failsWhenProcessorFails(t *testing.T) {
	processDelivery, acknowledger, _, inFlight := setUpProcessDelivery(errors.New("processing failed"))

	processDelivery(amqp091.Delivery{DeliveryTag: 9, Acknowledger: acknowledger})
	inFlight.Wait()

	want := uint64(9)
	got := <-acknowledger.nacks

	assert.Equal(t, want, got)
}
