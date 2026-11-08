package broker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/broker"
)

func TestRabbitMQBroker_InitMessageConsumer_failsWhenChannelClosesUnexpectedly(t *testing.T) {
	mockMessages := broker.MockClosedMessages()
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, mockMessages)

	want := broker.ErrConsumerChannelClosed
	got := b.InitMessageConsumer(1, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenConcurrencyIsSetAsQoS(t *testing.T) {
	mockChannel := &broker.MockChannel{ExpectedQos: 2}
	mockMessages := broker.MockClosedMessages()
	b := broker.MockRabbitMQBroker(mockChannel, "work", true, mockMessages)

	want := broker.ErrConsumerChannelClosed
	got := b.InitMessageConsumer(2, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenSubscriptionStartsAfterQoS(t *testing.T) {
	mockChannel := &broker.MockChannel{Messages: broker.MockClosedMessages(), ExpectedQos: 3}
	b := broker.MockRabbitMQBroker(mockChannel, "work", true, nil)

	want := broker.ErrConsumerChannelClosed
	got := b.InitMessageConsumer(3, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenSubscriptionFails(t *testing.T) {
	mockErr := errors.New("subscription failed")
	mockChannel := &broker.MockChannel{ConsumeErr: mockErr}
	b := broker.MockRabbitMQBroker(mockChannel, "work", true, nil)

	want := broker.ErrFailedToConsumeQueue
	got := b.InitMessageConsumer(2, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenTwoMessagesAndLimitIsOne(t *testing.T) {
	mockAck := broker.MockAcknowledger(2)
	mockFirstDelivery := amqp091.Delivery{DeliveryTag: 1, Acknowledger: mockAck}
	mockSecondDelivery := amqp091.Delivery{DeliveryTag: 2, Acknowledger: mockAck}
	mockMessages := broker.MockClosedMessages(
		mockFirstDelivery,
		mockSecondDelivery,
	)
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, mockMessages)
	t.Cleanup(b.CloseConnection)

	want := broker.ErrConsumerChannelClosed
	got := b.InitMessageConsumer(1, broker.MockSerialProcessor(t))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_successWhenStoppedViaCloseConnection(t *testing.T) {
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, make(chan amqp091.Delivery))
	b.CloseConnection()

	got := b.InitMessageConsumer(1, broker.MockProcessor(nil))

	assert.NoError(t, got)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenQoSUnavailable(t *testing.T) {
	mockErr := errors.New("qos unavailable")
	mockChannel := &broker.MockChannel{QosErr: mockErr}
	mockMessages := broker.MockClosedMessages(amqp091.Delivery{})
	b := broker.MockRabbitMQBroker(mockChannel, "work", true, mockMessages)

	want := broker.ErrFailedToSetQoS
	got := b.InitMessageConsumer(1, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenConcurrencyIsInvalid(t *testing.T) {
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, make(chan amqp091.Delivery))

	want := broker.ErrInvalidConcurrency
	got := b.InitMessageConsumer(0, nil)

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_InitMessageConsumer_failsWhenBrokerIsProducer(t *testing.T) {
	b := broker.MockRabbitMQBroker(nil, "", false, nil)

	want := broker.ErrNotAConsumer
	got := b.InitMessageConsumer(1, broker.MockProcessor(nil))

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_EnqueueMessage_success(t *testing.T) {
	mockChannel := &broker.MockChannel{ExpectedPublishing: &amqp091.Publishing{ContentType: "application/json", Body: []byte(`{"job":1}`)}}
	b := broker.MockRabbitMQBroker(mockChannel, "work", false, nil)

	got := b.EnqueueMessage(context.Background(), `{"job":1}`)

	assert.NoError(t, got)
}

func TestRabbitMQBroker_EnqueueMessage_failsWhenPublishFails(t *testing.T) {
	mockErr := errors.New("unavailable")
	mockChannel := &broker.MockChannel{PublishErr: mockErr}
	b := broker.MockRabbitMQBroker(mockChannel, "work", false, nil)

	want := broker.ErrFailedToPublishMessage
	got := b.EnqueueMessage(context.Background(), "{}")

	assert.ErrorIs(t, got, want)
}

func TestRabbitMQBroker_CloseConnection_successWhenProcessingIsInFlight(t *testing.T) {
	captureAck := broker.CaptureAcks(1, 1)
	messages := make(chan amqp091.Delivery, 1)
	messages <- amqp091.Delivery{DeliveryTag: 1, Acknowledger: captureAck}
	b := broker.MockRabbitMQBroker(&broker.MockChannel{}, "work", true, messages)
	started := make(chan struct{})
	go b.InitMessageConsumer(1, broker.MockSlowProcessor(started))
	<-started

	b.CloseConnection()

	want := 1
	got := len(captureAck.Acks)

	assert.Equal(t, want, got)
}
