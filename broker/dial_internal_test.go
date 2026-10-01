package broker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Internal (white-box) tests for dial: it is unexported, and exercising it against a real
// RabbitMQ broker is out of scope for a unit test — these only cover the error-wrapping and
// ctx-cancellation branches, both reachable without any network dependency.

func TestDial_errorWhenURLIsInvalid(t *testing.T) {
	_, err := dial(context.Background(), "not-a-valid-url")

	assert.ErrorIs(t, err, ErrFailedToConnect)
}

func TestDial_errorWhenContextIsAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// A real-looking but unroutable address, so amqp091.Dial doesn't fail synchronously before
	// the select ever runs — ctx must be what resolves this call.
	_, err := dial(ctx, "amqp://guest:guest@10.255.255.1:5672/")

	assert.ErrorIs(t, err, ErrFailedToConnect)
}

func TestInitMessageConsumerWithFunction_returnsImmediatelyWhenNotAConsumer(t *testing.T) {
	b := &RabbitMQBroker{} // built as if by NewProducer: messages is nil

	done := make(chan struct{})
	go func() {
		b.InitMessageConsumerWithFunction(1, func(context.Context, []byte) error { return nil })
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("InitMessageConsumerWithFunction blocked on a nil messages channel instead of returning")
	}
}
