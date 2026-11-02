package broker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeclareQueue_success(t *testing.T) {
	mockChannel := &MockChannel{}

	queue, _ := declareQueue(mockChannel, "work")

	want := "work"
	got := queue.Name

	assert.Equal(t, want, got)
}

func TestDeclareQueue_failsWhenChannelRejectsDeclaration(t *testing.T) {
	mockErr := errors.New("unavailable")
	mockChannel := &MockChannel{QueueErr: mockErr}

	want := ErrFailedToDeclareQueue
	wantCause := mockErr
	_, got := declareQueue(mockChannel, "work")

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}
