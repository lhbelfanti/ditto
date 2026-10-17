package http_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestGracefulShutdown_successWhenServeClosesCleanly(t *testing.T) {
	mockServe := _http.MockServe(http.ErrServerClosed)
	mockShutdown := _http.MockShutdown(nil)

	got := _http.GracefulShutdown(context.Background(), mockServe, mockShutdown, time.Second)

	assert.NoError(t, got)
}

func TestGracefulShutdown_failsWhenServeFailsUnexpectedly(t *testing.T) {
	want := errors.New("bind: address already in use")
	mockServe := _http.MockServe(want)
	mockShutdown := _http.MockShutdown(nil)

	got := _http.GracefulShutdown(context.Background(), mockServe, mockShutdown, time.Second)

	assert.ErrorIs(t, got, want)
}

func TestGracefulShutdown_successWhenContextIsCancelled(t *testing.T) {
	mockServe := _http.MockServeBlocking()
	calls := 0
	mockShutdown := _http.MockCountingShutdown(nil, &calls)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_ = _http.GracefulShutdown(ctx, mockServe, mockShutdown, time.Second)

	want := 1
	got := calls

	assert.Equal(t, want, got)
}

func TestGracefulShutdown_successWhenContextIsCancelledReturnsNil(t *testing.T) {
	mockServe := _http.MockServeBlocking()
	calls := 0
	mockShutdown := _http.MockCountingShutdown(nil, &calls)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := _http.GracefulShutdown(ctx, mockServe, mockShutdown, time.Second)

	assert.NoError(t, got)
}

func TestGracefulShutdown_failsWhenShutdownFails(t *testing.T) {
	mockServe := _http.MockServeBlocking()
	want := errors.New("shutdown: deadline exceeded")
	mockShutdown := _http.MockShutdown(want)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := _http.GracefulShutdown(ctx, mockServe, mockShutdown, time.Second)

	assert.ErrorIs(t, got, want)
}

func TestGracefulShutdown_successWhenShutdownIsBoundedByTimeout(t *testing.T) {
	mockServe := _http.MockServeBlocking()
	var captured context.Context
	mockShutdown := _http.MockCapturingShutdown(nil, &captured)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	timeout := 5 * time.Second
	start := time.Now()

	_ = _http.GracefulShutdown(ctx, mockServe, mockShutdown, timeout)

	want := start.Add(timeout)
	got, _ := captured.Deadline()

	assert.WithinDuration(t, want, got, time.Second)
}
