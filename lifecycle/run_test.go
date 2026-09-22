package lifecycle_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/lifecycle"
)

const testTimeout = 20 * time.Millisecond

func TestMakeRun_successWhenSignalCancelsContext(t *testing.T) {
	served := make(chan struct{})
	defer close(served)
	serve := lifecycle.MockServeUntilSignal(served, http.ErrServerClosed)
	shutdown := lifecycle.MockShutdown(nil)
	var calls int
	closeDatabase := lifecycle.MockCloseDatabase(&calls)

	run := lifecycle.MakeRun(serve, shutdown, closeDatabase, testTimeout, testTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := run(ctx)

	assert.Nil(t, got)
	assert.Equal(t, 1, calls)
}

func TestMakeRun_successWhenServeReturnsErrServerClosed(t *testing.T) {
	serve := lifecycle.MockServe(http.ErrServerClosed)
	shutdown := lifecycle.MockShutdown(nil)
	var calls int
	closeDatabase := lifecycle.MockCloseDatabase(&calls)

	run := lifecycle.MakeRun(serve, shutdown, closeDatabase, testTimeout, testTimeout)

	got := run(context.Background())

	assert.Nil(t, got)
	assert.Equal(t, 1, calls)
}

func TestMakeRun_failsWhenServeReturnsUnexpectedError(t *testing.T) {
	sourceErr := errors.New("bind: address already in use")
	serve := lifecycle.MockServe(sourceErr)
	shutdown := lifecycle.MockShutdown(nil)
	var calls int
	closeDatabase := lifecycle.MockCloseDatabase(&calls)

	run := lifecycle.MakeRun(serve, shutdown, closeDatabase, testTimeout, testTimeout)

	want := lifecycle.ErrServeFailed
	got := run(context.Background())

	assert.ErrorIs(t, got, want)
	assert.Equal(t, 1, calls)
}

func TestMakeRun_failsWhenShutdownFails(t *testing.T) {
	served := make(chan struct{})
	defer close(served)
	serve := lifecycle.MockServeUntilSignal(served, http.ErrServerClosed)
	sourceErr := errors.New("shutdown source error")
	shutdown := lifecycle.MockShutdown(sourceErr)
	var calls int
	closeDatabase := lifecycle.MockCloseDatabase(&calls)

	run := lifecycle.MakeRun(serve, shutdown, closeDatabase, testTimeout, testTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	want := lifecycle.ErrShutdownFailed
	got := run(ctx)

	assert.ErrorIs(t, got, want)
	assert.Equal(t, 1, calls)
}

func TestMakeRun_failsWhenShutdownTimesOut(t *testing.T) {
	served := make(chan struct{})
	defer close(served)
	serve := lifecycle.MockServeUntilSignal(served, http.ErrServerClosed)
	shutdown := lifecycle.MockShutdownRespectingContext()
	var calls int
	closeDatabase := lifecycle.MockCloseDatabase(&calls)

	run := lifecycle.MakeRun(serve, shutdown, closeDatabase, testTimeout, testTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	want := lifecycle.ErrShutdownTimedOut
	got := run(ctx)

	assert.ErrorIs(t, got, want)
	assert.Equal(t, 1, calls)
}

func TestMakeRun_failsWhenDatabaseCloseTimesOut(t *testing.T) {
	served := make(chan struct{})
	defer close(served)
	serve := lifecycle.MockServeUntilSignal(served, http.ErrServerClosed)
	shutdown := lifecycle.MockShutdown(nil)
	closeSignal := make(chan struct{})
	defer close(closeSignal)
	closeDatabase := lifecycle.MockBlockingCloseDatabase(closeSignal)

	run := lifecycle.MakeRun(serve, shutdown, closeDatabase, testTimeout, testTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	want := lifecycle.ErrDatabaseCloseTimeout
	got := run(ctx)

	assert.ErrorIs(t, got, want)
}

func TestMakeRun_successWhenShutdownRunsBeforeDatabaseClose(t *testing.T) {
	served := make(chan struct{})
	defer close(served)
	serve := lifecycle.MockServeUntilSignal(served, http.ErrServerClosed)
	var order []string
	shutdown := lifecycle.MockRecordingShutdown(&order, nil)
	closeDatabase := lifecycle.MockRecordingCloseDatabase(&order)

	run := lifecycle.MakeRun(serve, shutdown, closeDatabase, testTimeout, testTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	want := []string{"shutdown", "close"}
	_ = run(ctx)
	got := order

	assert.Equal(t, want, got)
}
