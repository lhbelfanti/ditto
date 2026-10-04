package http

import (
	"context"
	"net/http"
)

type (
	// Client is an abstraction of the CustomClient methods
	Client interface {
		NewRequest(ctx context.Context, method, url string, body interface{}) (Response, error)
	}

	// CustomClient represent a custom http.CustomClient
	CustomClient struct {
		HTTPClient *http.Client
	}

	// Response represent the necessary data of the request response
	Response struct {
		Body       string
		Status     string
		StatusCode int
		Header     http.Header
	}

	// Serve runs an HTTP server until it is shut down or fails unexpectedly.
	Serve func() error

	// Shutdown gracefully stops an HTTP server before the deadline carried by ctx elapses.
	Shutdown func(ctx context.Context) error

	// MigrationRunner is a function that executes pending database migrations.
	MigrationRunner func(ctx context.Context) error

	// DatabasePing checks whether the database dependency is reachable.
	DatabasePing func(ctx context.Context) error

	// SystemRoutes mounts ditto's standard system endpoints on a mux, one opt-in route at a time.
	SystemRoutes struct {
		mux *http.ServeMux
	}
)
