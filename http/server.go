package http

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type (
	// Serve runs an HTTP server until it is shut down or fails unexpectedly.
	Serve func() error

	// Shutdown gracefully stops an HTTP server before the deadline carried by ctx elapses.
	Shutdown func(ctx context.Context) error
)

// Listen serves handler on addr until ctx is cancelled, then shuts the server down gracefully
// within shutdownTimeout.
func Listen(ctx context.Context, addr string, handler http.Handler, shutdownTimeout time.Duration) error {
	server := &http.Server{Addr: addr, Handler: handler}
	return GracefulShutdown(ctx, server.ListenAndServe, server.Shutdown, shutdownTimeout)
}

// GracefulShutdown runs serve until ctx is cancelled, then calls shutdown bounded by timeout so
// in-flight requests get a chance to finish before the caller's process exits. It returns nil on
// a clean shutdown — http.ErrServerClosed from serve is never treated as a failure — and
// shutdown's own error otherwise.
func GracefulShutdown(ctx context.Context, serve Serve, shutdown Shutdown, timeout time.Duration) error {
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- serve()
	}()

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return shutdown(shutdownCtx)
	}
}
