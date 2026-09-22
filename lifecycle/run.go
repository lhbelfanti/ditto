package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type (
	// Serve runs the HTTP server until it is shut down or fails unexpectedly.
	Serve func() error

	// Shutdown gracefully stops the HTTP server before the deadline carried by ctx elapses.
	Shutdown func(ctx context.Context) error

	// CloseDatabase releases the database connection pool.
	CloseDatabase func()

	// Run drives the bounded, ordered lifecycle until ctx is cancelled or serving fails.
	Run func(ctx context.Context) error
)

// MakeRun creates a Run function. shutdownTimeout bounds HTTP shutdown and closeDatabaseTimeout
// separately bounds the non-context-aware database pool close that follows it.
func MakeRun(serve Serve, shutdown Shutdown, closeDatabase CloseDatabase, shutdownTimeout, closeDatabaseTimeout time.Duration) Run {
	return func(ctx context.Context) error {
		serveErr := make(chan error, 1)
		go func() {
			serveErr <- serve()
		}()

		select {
		case err := <-serveErr:
			closeErr := boundedCloseDatabase(closeDatabase, closeDatabaseTimeout)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("%w: %w", ErrServeFailed, err)
			}
			return closeErr
		case <-ctx.Done():
			shutdownErr := boundedShutdown(shutdown, shutdownTimeout)
			closeErr := boundedCloseDatabase(closeDatabase, closeDatabaseTimeout)
			if shutdownErr != nil {
				return shutdownErr
			}
			return closeErr
		}
	}
}

func boundedShutdown(shutdown Shutdown, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := shutdown(ctx)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("%w: %w", ErrShutdownTimedOut, err)
	default:
		return fmt.Errorf("%w: %w", ErrShutdownFailed, err)
	}
}

func boundedCloseDatabase(closeDatabase CloseDatabase, timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		closeDatabase()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return ErrDatabaseCloseTimeout
	}
}
