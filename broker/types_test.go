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
	b := broker.MockRabbitMQBroker(&broker.CaptureChannel{}, "work", true, broker.MockClosedMessages())

	want := broker.ErrConsumerChannelClosed
	got := b.InitMessageConsumer(1, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenMessageIsAcknowledged(t *testing.T) {
	captureAck := broker.CaptureAcks(1, 1)
	mockMessages := broker.MockClosedMessages(amqp091.Delivery{DeliveryTag: 1, Acknowledger: captureAck, Body: []byte("hello")})
	b := broker.MockRabbitMQBroker(&broker.CaptureChannel{}, "work", true, mockMessages)

	_ = b.InitMessageConsumer(2, broker.MockProcessor(nil))
	b.CloseConnection()

	want := uint64(1)
	got := <-captureAck.Acks

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenMessageBodyIsProcessed(t *testing.T) {
	captureAck := broker.CaptureAcks(1, 1)
	bodies := make(chan []byte, 1)
	mockMessages := broker.MockClosedMessages(amqp091.Delivery{DeliveryTag: 1, Acknowledger: captureAck, Body: []byte("hello")})
	b := broker.MockRabbitMQBroker(&broker.CaptureChannel{}, "work", true, mockMessages)

	_ = b.InitMessageConsumer(1, broker.MockRecordingProcessor(bodies, nil))
	b.CloseConnection()

	want := []byte("hello")
	got := <-bodies

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenConcurrencyIsSetAsQoS(t *testing.T) {
	captureChannel := &broker.CaptureChannel{}
	b := broker.MockRabbitMQBroker(captureChannel, "work", true, broker.MockClosedMessages())

	_ = b.InitMessageConsumer(2, broker.MockProcessor(nil))

	want := 2
	got := captureChannel.QosCount

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenSubscriptionStartsAfterQoS(t *testing.T) {
	captureChannel := &broker.CaptureChannel{Messages: broker.MockClosedMessages()}
	b := broker.MockRabbitMQBroker(captureChannel, "work", true, nil)

	_ = b.InitMessageConsumer(3, broker.MockProcessor(nil))

	want := 3
	got := captureChannel.QosAtConsume

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenQueueIsSubscribedOnce(t *testing.T) {
	captureChannel := &broker.CaptureChannel{Messages: broker.MockClosedMessages()}
	b := broker.MockRabbitMQBroker(captureChannel, "work", true, nil)

	_ = b.InitMessageConsumer(1, broker.MockProcessor(nil))

	want := 1
	got := captureChannel.ConsumeCalls

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenSubscriptionFails(t *testing.T) {
	mockErr := errors.New("subscription failed")
	captureChannel := &broker.CaptureChannel{ConsumeErr: mockErr}
	b := broker.MockRabbitMQBroker(captureChannel, "work", true, nil)

	want := broker.ErrFailedToConsumeQueue
	wantCause := mockErr
	got := b.InitMessageConsumer(2, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenProcessorFails(t *testing.T) {
	captureAck := broker.CaptureAcks(1, 1)
	mockMessages := broker.MockClosedMessages(amqp091.Delivery{DeliveryTag: 2, Acknowledger: captureAck})
	b := broker.MockRabbitMQBroker(&broker.CaptureChannel{}, "work", true, mockMessages)

	_ = b.InitMessageConsumer(1, broker.MockProcessor(errors.New("processing failed")))
	b.CloseConnection()

	want := broker.Acks{Nacks: []uint64{2}}
	got := captureAck.Recorded()

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenTwoMessagesAndLimitIsOne(t *testing.T) {
	captureAck := broker.CaptureAcks(2, 2)
	mockMessages := broker.MockClosedMessages(
		amqp091.Delivery{DeliveryTag: 1, Acknowledger: captureAck},
		amqp091.Delivery{DeliveryTag: 2, Acknowledger: captureAck},
	)
	b := broker.MockRabbitMQBroker(&broker.CaptureChannel{}, "work", true, mockMessages)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	done := make(chan struct{})
	mockProcessor := broker.MockBlockingProcessor(started, release, nil)
	go func() {
		_ = b.InitMessageConsumer(1, mockProcessor)
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
	got := len(captureAck.Acks)

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenStoppedViaCloseConnection(t *testing.T) {
	b := broker.MockRabbitMQBroker(&broker.CaptureChannel{}, "work", true, make(chan amqp091.Delivery))
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
	mockErr := errors.New("qos unavailable")
	captureChannel := &broker.CaptureChannel{QosErr: mockErr}
	b := broker.MockRabbitMQBroker(captureChannel, "work", true, broker.MockClosedMessages(amqp091.Delivery{}))

	want := broker.ErrFailedToSetQoS
	wantCause := mockErr
	got := b.InitMessageConsumer(1, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenConcurrencyIsInvalid(t *testing.T) {
	b := broker.MockRabbitMQBroker(&broker.CaptureChannel{}, "work", true, make(chan amqp091.Delivery))

	want := broker.ErrInvalidConcurrency
	got := b.InitMessageConsumer(0, nil)

	assert.ErrorIs(t, got, want)
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

	want := broker.ErrNotAConsumer

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_EnqueueMessage_success(t *testing.T) {
	captureChannel := &broker.CaptureChannel{}
	b := broker.MockRabbitMQBroker(captureChannel, "work", false, nil)

	_ = b.EnqueueMessage(context.Background(), `{"job":1}`)

	want := []byte(`{"job":1}`)
	got := captureChannel.Published.Body

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_EnqueueMessage_successWhenContentTypeIsJSON(t *testing.T) {
	captureChannel := &broker.CaptureChannel{}
	b := broker.MockRabbitMQBroker(captureChannel, "work", false, nil)

	_ = b.EnqueueMessage(context.Background(), `{"job":1}`)

	want := "application/json"
	got := captureChannel.Published.ContentType

	assert.Equal(t, want, got)
}

func TestRabbitMQBroker_EnqueueMessage_failsWhenPublishFails(t *testing.T) {
	mockErr := errors.New("unavailable")
	captureChannel := &broker.CaptureChannel{PublishErr: mockErr}
	b := broker.MockRabbitMQBroker(captureChannel, "work", false, nil)

	want := broker.ErrFailedToPublishMessage
	wantCause := mockErr
	got := b.EnqueueMessage(context.Background(), "{}")

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestRabbitMQBroker_CloseConnection_successWhenProcessingIsInFlight(t *testing.T) {
	captureAck := broker.CaptureAcks(1, 1)
	messages := make(chan amqp091.Delivery, 1)
	messages <- amqp091.Delivery{DeliveryTag: 1, Acknowledger: captureAck}
	b := broker.MockRabbitMQBroker(&broker.CaptureChannel{}, "work", true, messages)
	started := make(chan struct{})
	release := make(chan struct{})
	consumerDone := make(chan struct{})
	mockProcessor := broker.MockBlockingProcessor(started, release, nil)
	go func() {
		_ = b.InitMessageConsumer(1, mockProcessor)
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
	got := <-captureAck.Acks

	assert.Equal(t, want, got)
}
