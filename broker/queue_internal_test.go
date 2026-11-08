package broker

import (
	"errors"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestDeclareQueue_success(t *testing.T) {
	mockChannel := &MockChannel{}

	want := amqp091.Queue{Name: "work"}
	got, _ := declareQueue(mockChannel, "work")

	assert.Equal(t, want, got)
}

func TestDeclareQueue_failsWhenChannelRejectsDeclaration(t *testing.T) {
	mockErr := errors.New("unavailable")
	mockChannel := &MockChannel{QueueErr: mockErr}

	want := ErrFailedToDeclareQueue
	_, got := declareQueue(mockChannel, "work")

	assert.ErrorIs(t, got, want)
}
