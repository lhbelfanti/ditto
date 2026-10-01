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

	err := b.InitMessageConsumerWithFunction(2, MockRecordingProcessor(bodies, nil))
	b.inFlight.Wait()

	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), <-bodies)
	assert.Equal(t, 2, ch.qosCount)
	assert.Equal(t, uint64(1), <-acknowledger.acks)
	assert.Empty(t, acknowledger.nacks)
}

func TestInitMessageConsumerWithFunction_successWhenSubscriptionStartsAfterQoS(t *testing.T) {
	messages := make(chan amqp091.Delivery)
	close(messages)
	ch := &mockChannel{messages: messages}
	b := &RabbitMQBroker{channel: ch, queue: amqp091.Queue{Name: "work"}, consumer: true}

	err := b.InitMessageConsumerWithFunction(3, MockProcessor(nil))

	assert.NoError(t, err)
	assert.Equal(t, 1, ch.consumeCalls)
	assert.Equal(t, 3, ch.qosAtConsume)
}

func TestInitMessageConsumerWithFunction_failsWhenSubscriptionFails(t *testing.T) {
	underlying := errors.New("subscription failed")
	ch := &mockChannel{consumeErr: underlying}
	b := &RabbitMQBroker{channel: ch, queue: amqp091.Queue{Name: "work"}, consumer: true}

	err := b.InitMessageConsumerWithFunction(2, MockProcessor(nil))

	assert.ErrorIs(t, err, ErrFailedToConsumeQueue)
	assert.ErrorIs(t, err, underlying)
	assert.Equal(t, 1, ch.consumeCalls)
	assert.Equal(t, 2, ch.qosAtConsume)
}

func TestInitMessageConsumerWithFunction_successWhenProcessorFails(t *testing.T) {
	messages := make(chan amqp091.Delivery, 1)
	acknowledger := MockAcknowledger(1, 1)
	messages <- amqp091.Delivery{DeliveryTag: 2, Acknowledger: acknowledger}
	close(messages)
	b := &RabbitMQBroker{channel: &mockChannel{}, messages: messages}

	err := b.InitMessageConsumerWithFunction(1, MockProcessor(errors.New("processing failed")))
	b.inFlight.Wait()

	assert.NoError(t, err)
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
	var err error
	go func() {
		err = b.InitMessageConsumerWithFunction(1, processor)
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

	assert.NoError(t, err)
	assert.Len(t, acknowledger.acks, 2)
}

func TestInitMessageConsumerWithFunction_failsWhenQoSUnavailable(t *testing.T) {
	messages := make(chan amqp091.Delivery, 1)
	messages <- amqp091.Delivery{}
	underlying := errors.New("qos unavailable")
	ch := &mockChannel{qosErr: underlying}
	b := &RabbitMQBroker{channel: ch, messages: messages}
	bodies := make(chan []byte, 1)

	err := b.InitMessageConsumerWithFunction(1, MockRecordingProcessor(bodies, nil))

	assert.ErrorIs(t, err, ErrFailedToSetQoS)
	assert.ErrorIs(t, err, underlying)
	assert.Empty(t, bodies)
	assert.Equal(t, 1, ch.qosCount)
}

func TestInitMessageConsumerWithFunction_failsWhenConcurrencyIsInvalid(t *testing.T) {
	messages := make(chan amqp091.Delivery)
	ch := &mockChannel{}
	b := &RabbitMQBroker{channel: ch, messages: messages}

	err := b.InitMessageConsumerWithFunction(0, nil)

	assert.ErrorIs(t, err, ErrInvalidConcurrency)
	assert.Zero(t, ch.qosCount)
}

func TestInitMessageConsumerWithFunction_failsWhenBrokerIsProducer(t *testing.T) {
	b := &RabbitMQBroker{}
	done := make(chan struct{})
	var err error
	go func() {
		err = b.InitMessageConsumerWithFunction(1, MockProcessor(nil))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("InitMessageConsumerWithFunction blocked on a nil messages channel")
	}
	assert.ErrorIs(t, err, ErrNotAConsumer)
}
