package broker

import (
	"errors"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestInitMessageConsumerWithFunction_success(t *testing.T) {
	messages := make(chan amqp091.Delivery, 1)
	acknowledger := MockAcknowledger(1, 1)
	bodies := make(chan []byte, 1)
	messages <- amqp091.Delivery{DeliveryTag: 1, Acknowledger: acknowledger, Body: []byte("hello")}
	close(messages)
	ch := &mockChannel{}
	b := &RabbitMQBroker{channel: ch, messages: messages}

	b.InitMessageConsumerWithFunction(2, MockRecordingProcessor(bodies, nil))
	b.inFlight.Wait()

	assert.Equal(t, []byte("hello"), <-bodies)
	assert.Equal(t, 2, ch.qosCount)
	assert.Equal(t, uint64(1), <-acknowledger.acks)
	assert.Empty(t, acknowledger.nacks)
}

func TestInitMessageConsumerWithFunction_successWhenProcessorFails(t *testing.T) {
	messages := make(chan amqp091.Delivery, 1)
	acknowledger := MockAcknowledger(1, 1)
	messages <- amqp091.Delivery{DeliveryTag: 2, Acknowledger: acknowledger}
	close(messages)
	b := &RabbitMQBroker{channel: &mockChannel{}, messages: messages}

	b.InitMessageConsumerWithFunction(1, MockProcessor(errors.New("processing failed")))
	b.inFlight.Wait()

	assert.Equal(t, uint64(2), <-acknowledger.nacks)
	assert.Empty(t, acknowledger.acks)
}

func TestInitMessageConsumerWithFunction_successWhenTwoMessagesAndLimitIsOne(t *testing.T) {
	messages := make(chan amqp091.Delivery, 2)
	acknowledger := MockAcknowledger(2, 2)
	messages <- amqp091.Delivery{DeliveryTag: 1, Acknowledger: acknowledger}
	messages <- amqp091.Delivery{DeliveryTag: 2, Acknowledger: acknowledger}
	close(messages)
	b := &RabbitMQBroker{channel: &mockChannel{}, messages: messages}
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	done := make(chan struct{})
	processor := MockBlockingProcessor(started, release, nil)
	go func() {
		b.InitMessageConsumerWithFunction(1, processor)
		close(done)
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("first processor never started")
	}
	select {
	case <-started:
		t.Fatal("a second processor started while the first was still running")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("consumer did not finish")
	}
	b.inFlight.Wait()

	assert.Len(t, acknowledger.acks, 2)
}

func TestInitMessageConsumerWithFunction_failsWhenQoSUnavailable(t *testing.T) {
	messages := make(chan amqp091.Delivery, 1)
	messages <- amqp091.Delivery{}
	ch := &mockChannel{qosErr: errors.New("qos unavailable")}
	b := &RabbitMQBroker{channel: ch, messages: messages}
	bodies := make(chan []byte, 1)

	b.InitMessageConsumerWithFunction(1, MockRecordingProcessor(bodies, nil))

	assert.Empty(t, bodies)
	assert.Equal(t, 1, ch.qosCount)
}

func TestInitMessageConsumerWithFunction_failsWhenConcurrencyIsInvalid(t *testing.T) {
	messages := make(chan amqp091.Delivery)
	ch := &mockChannel{}
	b := &RabbitMQBroker{channel: ch, messages: messages}

	b.InitMessageConsumerWithFunction(0, nil)

	assert.Zero(t, ch.qosCount)
}

func TestInitMessageConsumerWithFunction_failsWhenBrokerIsProducer(t *testing.T) {
	b := &RabbitMQBroker{}
	done := make(chan struct{})
	go func() {
		b.InitMessageConsumerWithFunction(1, MockProcessor(nil))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("InitMessageConsumerWithFunction blocked on a nil messages channel")
	}
}
