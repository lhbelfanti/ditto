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
	done := make(chan struct{})
	serve := dittohttp.MockServeUntilDone(done, http.ErrServerClosed)
	shutdownCalls := 0
	shutdown := func(context.Context) error {
		shutdownCalls++
		close(done)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := dittohttp.GracefulShutdown(ctx, serve, shutdown, time.Second)

	assert.Nil(t, got)
	assert.Equal(t, 1, shutdownCalls)
}

func TestGracefulShutdown_returnsShutdownError(t *testing.T) {
	done := make(chan struct{})
	serve := dittohttp.MockServeUntilDone(done, http.ErrServerClosed)
	want := errors.New("shutdown: deadline exceeded")
	shutdown := func(context.Context) error {
		close(done)
		return want
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := dittohttp.GracefulShutdown(ctx, serve, shutdown, time.Second)

	assert.ErrorIs(t, got, want)
}

func TestGracefulShutdown_boundsShutdownWithTimeout(t *testing.T) {
	done := make(chan struct{})
	serve := dittohttp.MockServeUntilDone(done, http.ErrServerClosed)

	var gotDeadline time.Time
	var gotOK bool
	shutdown := func(shutdownCtx context.Context) error {
		gotDeadline, gotOK = shutdownCtx.Deadline()
		close(done)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	timeout := 5 * time.Second
	before := time.Now()
	_ = dittohttp.GracefulShutdown(ctx, serve, shutdown, timeout)
	after := time.Now()

	assert.True(t, gotOK)
	assert.True(t, !gotDeadline.Before(before.Add(timeout)))
	assert.True(t, !gotDeadline.After(after.Add(timeout)))
}
