package setup_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/log"
	"github.com/lhbelfanti/ditto/v2/setup"
)

func TestInit_success(t *testing.T) {
	want := "test"
	got := setup.Init(want, nil)

	assert.Equal(t, want, got)
}

func TestInit_failsWhenErrorIsPassed(t *testing.T) {
	output, restore := log.Capture()
	t.Cleanup(restore)
	initErr := errors.New("database: boom")
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = setup.Init("test", initErr)
	}()

	want := []any{initErr, []map[string]any{{"level": "error", "message": "failed to initialize a dependency", "error": "database: boom"}}}
	got := []any{recovered, output.Entries()}

	assert.Equal(t, want, got)
}

func TestMust_success(t *testing.T) {
	assert.NotPanics(t, func() {
		setup.Must(nil)
	})
}

func TestMust_failsWhenErrorIsPassed(t *testing.T) {
	output, restore := log.Capture()
	t.Cleanup(restore)
	initErr := errors.New("database: boom")
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		setup.Must(initErr)
	}()

	want := []any{initErr, []map[string]any{{"level": "error", "message": "failed to initialize a dependency", "error": "database: boom"}}}
	got := []any{recovered, output.Entries()}

	assert.Equal(t, want, got)
}
