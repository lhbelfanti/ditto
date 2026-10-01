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

func TestGracefulShutdown_returnsNilWhenServeClosesCleanly(t *testing.T) {
	serve := dittohttp.MockServe(http.ErrServerClosed)
	shutdown := dittohttp.MockShutdown(nil)

	got := dittohttp.GracefulShutdown(context.Background(), serve, shutdown, time.Second)

	assert.Nil(t, got)
}

func TestGracefulShutdown_returnsErrorWhenServeFailsUnexpectedly(t *testing.T) {
	want := errors.New("bind: address already in use")
	serve := dittohttp.MockServe(want)
	shutdown := dittohttp.MockShutdown(nil)

	got := dittohttp.GracefulShutdown(context.Background(), serve, shutdown, time.Second)

	assert.ErrorIs(t, got, want)
}

func TestGracefulShutdown_callsShutdownWhenContextIsCancelled(t *testing.T) {
	serve := dittohttp.MockServeBlocking()
	calls := 0
	shutdown := dittohttp.MockCountingShutdown(nil, &calls)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := dittohttp.GracefulShutdown(ctx, serve, shutdown, time.Second)

	assert.Nil(t, got)
	assert.Equal(t, 1, calls)
}

func TestGracefulShutdown_returnsShutdownError(t *testing.T) {
	serve := dittohttp.MockServeBlocking()
	want := errors.New("shutdown: deadline exceeded")
	shutdown := dittohttp.MockShutdown(want)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := dittohttp.GracefulShutdown(ctx, serve, shutdown, time.Second)

	assert.ErrorIs(t, got, want)
}

func TestGracefulShutdown_boundsShutdownWithTimeout(t *testing.T) {
	serve := dittohttp.MockServeBlocking()
	var captured context.Context
	shutdown := dittohttp.MockCapturingShutdown(nil, &captured)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	timeout := 5 * time.Second
	before := time.Now()
	_ = dittohttp.GracefulShutdown(ctx, serve, shutdown, timeout)
	after := time.Now()

	deadline, ok := captured.Deadline()
	assert.True(t, ok)
	assert.True(t, !deadline.Before(before.Add(timeout)))
	assert.True(t, !deadline.After(after.Add(timeout)))
}
