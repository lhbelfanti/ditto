package broker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeclareQueue_success(t *testing.T) {
	ch := &MockChannel{}

	q, err := declareQueue(ch, "work")

	assert.NoError(t, err)
	assert.Equal(t, "work", q.Name)
}

func TestDeclareQueue_failsWhenChannelRejectsDeclaration(t *testing.T) {
	underlying := errors.New("unavailable")
	ch := &MockChannel{QueueErr: underlying}

	_, err := declareQueue(ch, "work")

	assert.ErrorIs(t, err, ErrFailedToDeclareQueue)
	assert.ErrorIs(t, err, underlying)
}
