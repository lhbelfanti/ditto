package broker

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDial_failsWhenURLIsInvalid(t *testing.T) {
	_, err := dial(context.Background(), "not-a-valid-url")

	assert.ErrorIs(t, err, ErrFailedToConnect)
}

func TestDial_failsWhenContextIsAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// A real-looking but unroutable address, so amqp091.Dial doesn't fail synchronously before
	// the select ever runs — ctx must be what resolves this call.
	_, err := dial(ctx, "amqp://guest:guest@10.255.255.1:5672/")

	assert.ErrorIs(t, err, ErrFailedToConnect)
}

func TestCloseLateDial_successWhenDialFailed(t *testing.T) {
	done := make(chan dialResult, 1)
	done <- dialResult{err: errors.New("dial failed")}

	closeLateDial(done)

	want := 0
	got := len(done)

	assert.Equal(t, want, got)
}
