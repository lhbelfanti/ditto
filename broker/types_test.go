package broker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/broker"
)

func TestRabbitMQBroker_InitMessageConsumer_failsWhenChannelClosesUnexpectedly(t *testing.T) {
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, broker.MockClosedMessages())

	want := broker.ErrConsumerChannelClosed
	got := b.InitMessageConsumer(1, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenMessageIsAcknowledged(t *testing.T) {
	ack := broker.MockAcknowledger(1, 1)
	messages := broker.MockClosedMessages(amqp091.Delivery{DeliveryTag: 1, Acknowledger: ack, Body: []byte("hello")})
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, messages)

	_ = b.InitMessageConsumer(2, broker.MockProcessor(nil))
	b.CloseConnection()

	want := uint64(1)
	got := <-ack.Acks

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenMessageBodyIsProcessed(t *testing.T) {
	ack := broker.MockAcknowledger(1, 1)
	bodies := make(chan []byte, 1)
	messages := broker.MockClosedMessages(amqp091.Delivery{DeliveryTag: 1, Acknowledger: ack, Body: []byte("hello")})
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, messages)

	_ = b.InitMessageConsumer(1, broker.MockRecordingProcessor(bodies, nil))
	b.CloseConnection()

	want := []byte("hello")
	got := <-bodies

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenConcurrencyIsSetAsQoS(t *testing.T) {
	ch := &broker.MockChannel{}
	b := broker.MockRabbitMQBroker(ch, "work", true, broker.MockClosedMessages())

	_ = b.InitMessageConsumer(2, broker.MockProcessor(nil))

	want := 2
	got := ch.QosCount

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenSubscriptionStartsAfterQoS(t *testing.T) {
	ch := &broker.MockChannel{Messages: broker.MockClosedMessages()}
	b := broker.MockRabbitMQBroker(ch, "work", true, nil)

	_ = b.InitMessageConsumer(3, broker.MockProcessor(nil))

	want := 3
	got := ch.QosAtConsume

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenQueueIsSubscribedOnce(t *testing.T) {
	ch := &broker.MockChannel{Messages: broker.MockClosedMessages()}
	b := broker.MockRabbitMQBroker(ch, "work", true, nil)

	_ = b.InitMessageConsumer(1, broker.MockProcessor(nil))

	want := 1
	got := ch.ConsumeCalls

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenSubscriptionFails(t *testing.T) {
	ch := &broker.MockChannel{ConsumeErr: errors.New("subscription failed")}
	b := broker.MockRabbitMQBroker(ch, "work", true, nil)

	want := broker.ErrFailedToConsumeQueue
	got := b.InitMessageConsumer(2, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenSubscriptionFailsKeepsCause(t *testing.T) {
	want := errors.New("subscription failed")
	ch := &broker.MockChannel{ConsumeErr: want}
	b := broker.MockRabbitMQBroker(ch, "work", true, nil)

	got := b.InitMessageConsumer(2, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenProcessorFails(t *testing.T) {
	ack := broker.MockAcknowledger(1, 1)
	messages := broker.MockClosedMessages(amqp091.Delivery{DeliveryTag: 2, Acknowledger: ack})
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, messages)

	_ = b.InitMessageConsumer(1, broker.MockProcessor(errors.New("processing failed")))
	b.CloseConnection()

	want := uint64(2)
	got := <-ack.Nacks

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenProcessorFailsDoesNotAcknowledge(t *testing.T) {
	ack := broker.MockAcknowledger(1, 1)
	messages := broker.MockClosedMessages(amqp091.Delivery{DeliveryTag: 2, Acknowledger: ack})
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, messages)

	_ = b.InitMessageConsumer(1, broker.MockProcessor(errors.New("processing failed")))
	b.CloseConnection()

	want := 0
	got := len(ack.Acks)

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenTwoMessagesAndLimitIsOne(t *testing.T) {
	ack := broker.MockAcknowledger(2, 2)
	messages := broker.MockClosedMessages(
		amqp091.Delivery{DeliveryTag: 1, Acknowledger: ack},
		amqp091.Delivery{DeliveryTag: 2, Acknowledger: ack},
	)
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, messages)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	done := make(chan struct{})
	processor := broker.MockBlockingProcessor(started, release, nil)
	go func() {
		_ = b.InitMessageConsumer(1, processor)
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
	b.CloseConnection()

	want := 2
	got := len(ack.Acks)

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenStoppedViaCloseConnection(t *testing.T) {
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, make(chan amqp091.Delivery))
	done := make(chan struct{})
	var got error
	go func() {
		got = b.InitMessageConsumer(1, broker.MockProcessor(nil))
		close(done)
	}()

	b.CloseConnection()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("consumer did not stop after CloseConnection")
	}

	assert.NoError(t, got)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenQoSUnavailable(t *testing.T) {
	ch := &broker.MockChannel{QosErr: errors.New("qos unavailable")}
	b := broker.MockRabbitMQBroker(ch, "work", true, broker.MockClosedMessages(amqp091.Delivery{}))

	want := broker.ErrFailedToSetQoS
	got := b.InitMessageConsumer(1, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenQoSUnavailableKeepsCause(t *testing.T) {
	want := errors.New("qos unavailable")
	ch := &broker.MockChannel{QosErr: want}
	b := broker.MockRabbitMQBroker(ch, "work", true, broker.MockClosedMessages(amqp091.Delivery{}))

	got := b.InitMessageConsumer(1, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenQoSUnavailableDoesNotProcess(t *testing.T) {
	ch := &broker.MockChannel{QosErr: errors.New("qos unavailable")}
	bodies := make(chan []byte, 1)
	b := broker.MockRabbitMQBroker(ch, "work", true, broker.MockClosedMessages(amqp091.Delivery{}))

	_ = b.InitMessageConsumer(1, broker.MockRecordingProcessor(bodies, nil))

	want := 0
	got := len(bodies)

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenConcurrencyIsInvalid(t *testing.T) {
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, make(chan amqp091.Delivery))

	want := broker.ErrInvalidConcurrency
	got := b.InitMessageConsumer(0, nil)

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenConcurrencyIsInvalidSkipsQoS(t *testing.T) {
	ch := &broker.MockChannel{}
	b := broker.MockRabbitMQBroker(ch, "work", true, make(chan amqp091.Delivery))

	_ = b.InitMessageConsumer(0, nil)

	want := 0
	got := ch.QosCount

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenBrokerIsProducer(t *testing.T) {
	b := broker.MockRabbitMQBroker(nil, "", false, nil)
	done := make(chan struct{})
	var got error
	go func() {
		got = b.InitMessageConsumer(1, broker.MockProcessor(nil))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("InitMessageConsumer blocked on a nil messages channel")
	}

	assert.ErrorIs(t, got, broker.ErrNotAConsumer)
}

func TestRabbitMQBroker_EnqueueMessage_success(t *testing.T) {
	ch := &broker.MockChannel{}
	b := broker.MockRabbitMQBroker(ch, "work", false, nil)

	_ = b.EnqueueMessage(context.Background(), `{"job":1}`)

	want := []byte(`{"job":1}`)
	got := ch.Published.Body

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_EnqueueMessage_successWhenContentTypeIsJSON(t *testing.T) {
	ch := &broker.MockChannel{}
	b := broker.MockRabbitMQBroker(ch, "work", false, nil)

	_ = b.EnqueueMessage(context.Background(), `{"job":1}`)

	want := "application/json"
	got := ch.Published.ContentType

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_EnqueueMessage_failsWhenPublishFails(t *testing.T) {
	ch := &broker.MockChannel{PublishErr: errors.New("unavailable")}
	b := broker.MockRabbitMQBroker(ch, "work", false, nil)

	want := broker.ErrFailedToPublishMessage
	got := b.EnqueueMessage(context.Background(), "{}")

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_EnqueueMessage_failsWhenPublishFailsKeepsCause(t *testing.T) {
	want := errors.New("unavailable")
	ch := &broker.MockChannel{PublishErr: want}
	b := broker.MockRabbitMQBroker(ch, "work", false, nil)

	got := b.EnqueueMessage(context.Background(), "{}")

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_CloseConnection_successWhenProcessingIsInFlight(t *testing.T) {
	ack := broker.MockAcknowledger(1, 1)
	messages := make(chan amqp091.Delivery, 1)
	messages <- amqp091.Delivery{DeliveryTag: 1, Acknowledger: ack}
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, messages)
	started := make(chan struct{})
	release := make(chan struct{})
	consumerDone := make(chan struct{})
	processor := broker.MockBlockingProcessor(started, release, nil)
	go func() {
		_ = b.InitMessageConsumer(1, processor)
		close(consumerDone)
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("processor never started")
	}

	closed := make(chan struct{})
	go func() {
		b.CloseConnection()
		close(closed)
	}()
	select {
	case <-closed:
		t.Fatal("CloseConnection returned before in-flight processing finished")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("CloseConnection did not finish after processing completed")
	}
	select {
	case <-consumerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("consumer remained blocked after shutdown")
	}

	want := uint64(1)
	got := <-ack.Acks

	assert.Equal(t, want, got)
}
