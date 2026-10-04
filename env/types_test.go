package env_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/env"
)

func TestKeyError_Error_success(t *testing.T) {
	keyErr := &env.KeyError{Key: "APP_PORT", Err: errors.New("is missing")}

	want := "APP_PORT: is missing"
	got := keyErr.Error()

	assert.Equal(t, want, got)
}

func TestKeyError_Unwrap_success(t *testing.T) {
	want := errors.New("is missing")
	keyErr := &env.KeyError{Key: "APP_PORT", Err: want}

	got := keyErr.Unwrap()

	assert.Equal(t, want, got)
}
