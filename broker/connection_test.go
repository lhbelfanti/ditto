package broker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/broker"
)

func TestMakeConnection(t *testing.T) {
	t.Run("MockEnqueue(nil) returns no error on EnqueueMessage", func(t *testing.T) {
		m := broker.MockEnqueue(nil)
		err := m.EnqueueMessage(context.Background(), "{}")
		assert.NoError(t, err)
	})

	t.Run("MockEnqueue(err) returns error on EnqueueMessage", func(t *testing.T) {
		expectedErr := errors.New("fail")
		m := broker.MockEnqueue(expectedErr)
		err := m.EnqueueMessage(context.Background(), "{}")
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("MockEnqueue implements CloseConnection without panic", func(t *testing.T) {
		m := broker.MockEnqueue(nil)
		assert.NotPanics(t, func() {
			m.CloseConnection()
		})
	})

	t.Run("MockEnqueue implements InitMessageConsumerWithFunction without panic", func(t *testing.T) {
		m := broker.MockEnqueue(nil)
		assert.NotPanics(t, func() {
			m.InitMessageConsumerWithFunction(1, nil)
		})
	})
}
