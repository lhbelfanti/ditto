package broker_test

import (
	"context"
	"errors"
	"testing"

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

func TestMockBroker_InitMessageConsumerWithFunction_success(t *testing.T) {
	mockBroker := &broker.MockBroker{}

	got := mockBroker.InitMessageConsumerWithFunction(1, nil)

	assert.NoError(t, got)
}
