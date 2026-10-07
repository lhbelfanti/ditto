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
	assert.Panics(t, func() {
		_ = setup.Init("test", errors.New("initialization failed"))
	})
}

func TestMust_success(t *testing.T) {
	assert.NotPanics(t, func() {
		setup.Must(nil)
	})
}

func TestMust_failsWhenErrorIsPassed(t *testing.T) {
	assert.Panics(t, func() {
		setup.Must(errors.New("initialization failed"))
	})
}

func TestInit_successWhenErrorIsLoggedBeforePanicking(t *testing.T) {
	output := log.MockLogOutput(t)

	assert.Panics(t, func() {
		_ = setup.Init("test", errors.New("database: boom"))
	})

	assert.Contains(t, output.String(), "database: boom")
}

func TestMust_successWhenErrorIsLoggedBeforePanicking(t *testing.T) {
	output := log.MockLogOutput(t)

	assert.Panics(t, func() {
		setup.Must(errors.New("database: boom"))
	})

	assert.Contains(t, output.String(), "database: boom")
}
