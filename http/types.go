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
		NewRequest(ctx context.Context, method, url string, body any) (Response, error)
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

// NewRequest sends a request with an optional JSON body and returns the response, with the body read and the headers exposed.
func (c *CustomClient) NewRequest(ctx context.Context, method, url string, body any) (Response, error) {
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
				return Response{}, fmt.Errorf("%w: %w", FailedToMarshalBody, err)
			}
		}

		reqBody = bytes.NewBuffer(jsonData)
		hasJSONBody = true
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return Response{}, fmt.Errorf("%w: %w", FailedToCreateRequest, err)
	}

	// A nil body means the caller is making a body-less request (e.g. a GET with only
	// query parameters, like an exchange klines endpoint) — Content-Type would be
	// misleading there, so it is only set when a JSON body is actually being sent.
	if hasJSONBody {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("%w: %w", FailedToExecuteRequest, err)
	}

	defer func(body io.ReadCloser) {
		err = body.Close()
		if err != nil {
			log.Warn(ctx, "failed to close the response body: "+err.Error())
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("%w: %w", FailedToReadResponse, err)
	}

	return Response{
		Body:       string(respBody),
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
	}, nil
}
