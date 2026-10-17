package database

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeError_Error_success(t *testing.T) {
	safe := &SafeError{sentinel: ErrDatabaseUnavailable, cause: errors.New("dial tcp user:secret@host: refused")}

	want := ErrDatabaseUnavailable.Error()
	got := safe.Error()

	assert.Equal(t, want, got)
}

func TestSafeError_Unwrap_success(t *testing.T) {
	cause := errors.New("dial failed")
	safe := &SafeError{sentinel: ErrCantInitDatabase, cause: cause}

	want := []error{ErrCantInitDatabase, cause}
	got := safe.Unwrap()

	assert.Equal(t, want, got)
}
