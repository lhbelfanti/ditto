package http

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunServe_success(t *testing.T) {
	mockErr := errors.New("boom")
	mockServe := MockServe(mockErr)
	serveErr := make(chan error, 1)

	runServe(mockServe, serveErr)

	want := mockErr
	got := <-serveErr

	assert.ErrorIs(t, got, want)
}
