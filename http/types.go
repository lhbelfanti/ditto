package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/log"
)

type (
	// Client is an abstraction of the CustomClient methods
	Client interface {
		NewRequest(ctx context.Context, method, url string, body any, opts ...RequestOption) (Response, error)
	}

	// RequestOption changes an outgoing request before it is sent.
	RequestOption func(req *http.Request)

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

	// DatabasePing checks whether the database dependency is reachable.
	DatabasePing func(ctx context.Context) error

	// SystemRoutes mounts ditto's standard system endpoints on a mux, one opt-in route at a time.
	SystemRoutes struct {
		mux *http.ServeMux
	}
)

// WithDatabasePing mounts GET /database/ping/v1.
func (s *SystemRoutes) WithDatabasePing(dbPing DatabasePing) *SystemRoutes {
	s.mux.HandleFunc("GET /database/ping/v1", databasePingHandlerV1(dbPing))
	return s
}

// NewRequest sends a request with an optional JSON body and returns the response, with the body read and the headers exposed. Each option is applied to the request before it is sent.
func (c *CustomClient) NewRequest(ctx context.Context, method, url string, body any, opts ...RequestOption) (Response, error) {
	var reqBody io.Reader
	var hasJSONBody bool
	if body != nil {
		var jsonData []byte
		var err error
		switch v := body.(type) {
		case []byte:
			jsonData = v
		default:
			jsonData, err = json.Marshal(body)
			if err != nil {
				return Response{}, fmt.Errorf("%w: %w", ErrFailedToMarshalBody, err)
			}
		}

		reqBody = bytes.NewBuffer(jsonData)
		hasJSONBody = true
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return Response{}, fmt.Errorf("%w: %w", ErrFailedToCreateRequest, err)
	}

	if hasJSONBody {
		req.Header.Set("Content-Type", "application/json")
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("%w: %w", ErrFailedToExecuteRequest, err)
	}

	defer func(body io.ReadCloser) {
		closeErr := body.Close()
		if closeErr != nil {
			log.Warn(log.With(ctx, log.Param("cause", closeErr)), "failed to close the response body")
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("%w: %w", ErrFailedToReadResponse, err)
	}

	return Response{
		Body:       string(respBody),
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
	}, nil
}
