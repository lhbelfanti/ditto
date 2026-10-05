package http_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	dittohttp "github.com/lhbelfanti/ditto/v2/http"
)

func TestGracefulShutdown_successWhenServeClosesCleanly(t *testing.T) {
	serve := dittohttp.MockServe(http.ErrServerClosed)
	shutdown := dittohttp.MockShutdown(nil)

	got := dittohttp.GracefulShutdown(context.Background(), serve, shutdown, time.Second)

	assert.NoError(t, got)
}

func TestGracefulShutdown_failsWhenServeFailsUnexpectedly(t *testing.T) {
	want := errors.New("bind: address already in use")
	serve := dittohttp.MockServe(want)
	shutdown := dittohttp.MockShutdown(nil)

	got := dittohttp.GracefulShutdown(context.Background(), serve, shutdown, time.Second)

	assert.ErrorIs(t, got, want)
}

func TestGracefulShutdown_successWhenContextIsCancelled(t *testing.T) {
	serve := dittohttp.MockServeBlocking()
	calls := 0
	shutdown := dittohttp.MockCountingShutdown(nil, &calls)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_ = dittohttp.GracefulShutdown(ctx, serve, shutdown, time.Second)

	want := 1
	got := calls

	assert.Equal(t, want, got)
}

func TestGracefulShutdown_successWhenContextIsCancelledReturnsNil(t *testing.T) {
	serve := dittohttp.MockServeBlocking()
	calls := 0
	shutdown := dittohttp.MockCountingShutdown(nil, &calls)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := dittohttp.GracefulShutdown(ctx, serve, shutdown, time.Second)

	assert.NoError(t, got)
}

func TestGracefulShutdown_failsWhenShutdownFails(t *testing.T) {
	serve := dittohttp.MockServeBlocking()
	want := errors.New("shutdown: deadline exceeded")
	shutdown := dittohttp.MockShutdown(want)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := dittohttp.GracefulShutdown(ctx, serve, shutdown, time.Second)

	assert.ErrorIs(t, got, want)
}

func TestGracefulShutdown_successWhenShutdownIsBoundedByTimeout(t *testing.T) {
	serve := dittohttp.MockServeBlocking()
	var captured context.Context
	shutdown := dittohttp.MockCapturingShutdown(nil, &captured)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	timeout := 5 * time.Second
	start := time.Now()

	_ = dittohttp.GracefulShutdown(ctx, serve, shutdown, timeout)

	want := start.Add(timeout)
	got, _ := captured.Deadline()

	assert.WithinDuration(t, want, got, time.Second)
}
