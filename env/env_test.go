package env_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/env"
)

func TestGet_success(t *testing.T) {
	t.Setenv("TEST_KEY", "hello")

	want := "hello"
	got := env.Get("TEST_KEY", "fallback")

	assert.Equal(t, want, got)
}

func TestGet_successWhenKeyIsUnset(t *testing.T) {
	want := "fallback"
	got := env.Get("TEST_KEY_UNSET_XYZ", "fallback")

	assert.Equal(t, want, got)
}

func TestGet_successWhenValueIsEmpty(t *testing.T) {
	t.Setenv("TEST_KEY_EMPTY", "")

	want := "fallback"
	got := env.Get("TEST_KEY_EMPTY", "fallback")

	assert.Equal(t, want, got)
}

func TestGetOrPanic_success(t *testing.T) {
	t.Setenv("TEST_KEY_PANIC", "world")

	want := "world"
	got := env.GetOrPanic("TEST_KEY_PANIC")

	assert.Equal(t, want, got)
}

func TestGetOrPanic_failsWhenKeyIsUnset(t *testing.T) {
	assert.Panics(t, func() {
		env.GetOrPanic("TEST_KEY_UNSET_PANIC_XYZ")
	})
}

func TestGetOrPanic_failsWhenValueIsEmpty(t *testing.T) {
	t.Setenv("TEST_KEY_PANIC_EMPTY", "")

	assert.Panics(t, func() {
		env.GetOrPanic("TEST_KEY_PANIC_EMPTY")
	})
}
