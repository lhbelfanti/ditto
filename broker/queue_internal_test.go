package broker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeclareQueue_success(t *testing.T) {
	captureChannel := &CaptureChannel{}

	queue, _ := declareQueue(captureChannel, "work")

	want := "work"
	got := queue.Name

	assert.Equal(t, want, got)
}

func TestDeclareQueue_failsWhenChannelRejectsDeclaration(t *testing.T) {
	mockErr := errors.New("unavailable")
	captureChannel := &CaptureChannel{QueueErr: mockErr}

	want := ErrFailedToDeclareQueue
	wantCause := mockErr
	_, got := declareQueue(captureChannel, "work")

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}
