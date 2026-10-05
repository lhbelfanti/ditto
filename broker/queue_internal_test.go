package broker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeclareQueue_success(t *testing.T) {
	ch := &MockChannel{}

	queue, _ := declareQueue(ch, "work")

	want := "work"
	got := queue.Name

	assert.Equal(t, want, got)
}

func TestDeclareQueue_failsWhenChannelRejectsDeclaration(t *testing.T) {
	ch := &MockChannel{QueueErr: errors.New("unavailable")}

	want := ErrFailedToDeclareQueue
	_, got := declareQueue(ch, "work")

	assert.ErrorIs(t, got, want)
}

func TestDeclareQueue_failsWhenChannelRejectsDeclarationKeepsCause(t *testing.T) {
	want := errors.New("unavailable")
	ch := &MockChannel{QueueErr: want}

	_, got := declareQueue(ch, "work")

	assert.ErrorIs(t, got, want)
}
