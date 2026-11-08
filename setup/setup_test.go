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
	recovered := setup.CapturePanic(func() { setup.Init("test", initErr) })

	want := []any{initErr, []map[string]any{{"level": "error", "message": "failed to initialize a dependency", "error": "database: boom"}}}
	got := []any{recovered, output.Entries()}

	assert.Equal(t, want, got)
}

func TestMust_success(t *testing.T) {
	got := setup.CapturePanic(func() { setup.Must(nil) })

	assert.Nil(t, got)
}

func TestMust_failsWhenErrorIsPassed(t *testing.T) {
	output, restore := log.Capture()
	t.Cleanup(restore)
	initErr := errors.New("database: boom")
	recovered := setup.CapturePanic(func() { setup.Must(initErr) })

	want := []any{initErr, []map[string]any{{"level": "error", "message": "failed to initialize a dependency", "error": "database: boom"}}}
	got := []any{recovered, output.Entries()}

	assert.Equal(t, want, got)
}
