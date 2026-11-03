package broker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/broker"
)

func TestMockEnqueue_success(t *testing.T) {
	mockBroker := broker.MockEnqueue(nil)

	got := mockBroker.EnqueueMessage(context.Background(), "{}")

	assert.NoError(t, got)
}

func TestMockEnqueue_failsWhenEnqueueErrorIsConfigured(t *testing.T) {
	want := errors.New("fail")
	mockBroker := broker.MockEnqueue(want)

	got := mockBroker.EnqueueMessage(context.Background(), "{}")

	assert.ErrorIs(t, got, want)
}

func TestMockBroker_CloseConnection_success(t *testing.T) {
	mockBroker := &broker.MockBroker{}

	assert.NotPanics(t, mockBroker.CloseConnection)
}

func TestMockBroker_InitMessageConsumer_success(t *testing.T) {
	mockBroker := &broker.MockBroker{}

	got := mockBroker.InitMessageConsumer(1, nil)

	assert.NoError(t, got)
}

func TestMockChannel_Qos_failsWhenCountIsUnexpected(t *testing.T) {
	mockChannel := &broker.MockChannel{ExpectedQos: 2}

	got := mockChannel.Qos(3, 0, false)

	assert.Error(t, got)
}

func TestMockChannel_Consume_failsWhenExpectedQosIsNotInEffect(t *testing.T) {
	mockChannel := &broker.MockChannel{ExpectedQos: 2}

	_, got := mockChannel.Consume("work", "", false, false, false, false, nil)

	assert.Error(t, got)
}

func TestMockChannel_PublishWithContext_failsWhenMessageIsUnexpected(t *testing.T) {
	mockChannel := &broker.MockChannel{ExpectedPublishing: &amqp091.Publishing{Body: []byte("expected")}}

	got := mockChannel.PublishWithContext(context.Background(), "", "work", false, false, amqp091.Publishing{Body: []byte("other")})

	assert.Error(t, got)
}
