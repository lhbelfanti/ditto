package env_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/env"
)

func TestGet_returnsValueWhenSet(t *testing.T) {
	t.Setenv("TEST_KEY", "hello")
	got := env.Get("TEST_KEY", "fallback")
	assert.Equal(t, "hello", got)
}

func TestGet_returnsFallbackWhenUnset(t *testing.T) {
	got := env.Get("TEST_KEY_UNSET_XYZ", "fallback")
	assert.Equal(t, "fallback", got)
}

func TestGet_returnsFallbackWhenEmpty(t *testing.T) {
	t.Setenv("TEST_KEY_EMPTY", "")
	got := env.Get("TEST_KEY_EMPTY", "fallback")
	assert.Equal(t, "fallback", got)
}

func TestGetOrPanic_returnsValueWhenSet(t *testing.T) {
	t.Setenv("TEST_KEY_PANIC", "world")
	got := env.GetOrPanic("TEST_KEY_PANIC")
	assert.Equal(t, "world", got)
}

func TestGetOrPanic_panicsWhenUnset(t *testing.T) {
	assert.Panics(t, func() {
		env.GetOrPanic("TEST_KEY_UNSET_PANIC_XYZ")
	})
}

func TestGetOrPanic_panicsWhenEmpty(t *testing.T) {
	t.Setenv("TEST_KEY_PANIC_EMPTY", "")
	assert.Panics(t, func() {
		env.GetOrPanic("TEST_KEY_PANIC_EMPTY")
	})
}
