package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type (
	// Client is an abstraction of the CustomClient methods
	Client interface {
		NewRequest(method, url string) *Request
	}

	// CustomClient represent a custom http.CustomClient
	CustomClient struct {
		HTTPClient *http.Client
	}

	// Request is an outgoing request: it collects its headers, query params and body until Call sends it.
	Request struct {
		httpClient *http.Client
		method     string
		url        string
		header     http.Header
		query      url.Values
		body       any
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
func (s *SystemRoutes) WithDatabasePing(databasePing DatabasePing) *SystemRoutes {
	s.mux.HandleFunc("GET /database/ping/v1", databasePingHandlerV1(databasePing))
	return s
}

// NewRequest creates a Request with method and endpoint, sent through this client when Call runs.
func (c *CustomClient) NewRequest(method, endpoint string) *Request {
	return &Request{
		httpClient: c.HTTPClient,
		method:     method,
		url:        endpoint,
		header:     http.Header{},
		query:      url.Values{},
	}
}

// WithHeader sets a header on the request, replacing any value it already has.
func (r *Request) WithHeader(key, value string) {
	r.header.Set(key, value)
}

// WithBearer sets the Authorization header to the bearer token.
func (r *Request) WithBearer(token string) {
	r.header.Set("Authorization", "Bearer "+token)
}

// WithCookies sets the Cookie header to the cookies response set, the way a browser sends them back.
func (r *Request) WithCookies(response Response) {
	cookies := (&http.Response{Header: response.Header}).Cookies()
	if len(cookies) == 0 {
		return
	}

	pairs := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		pairs = append(pairs, cookie.Name+"="+cookie.Value)
	}

	r.header.Set("Cookie", strings.Join(pairs, "; "))
}

// WithQueryParam adds a query param to the request, after the ones the url already has.
func (r *Request) WithQueryParam(key, value string) {
	r.query.Add(key, value)
}

// WithBody sets the body of the request: a []byte is sent as it is, any other value as its JSON.
func (r *Request) WithBody(body any) {
	r.body = body
}

// Call sends the request and returns the response, with the body read and the headers exposed.
func (r *Request) Call(ctx context.Context) (Response, error) {
	var reqBody io.Reader
	var hasJSONBody bool
	if r.body != nil {
		var jsonData []byte
		var err error
		switch v := r.body.(type) {
		case []byte:
			jsonData = v
		default:
			jsonData, err = json.Marshal(r.body)
			if err != nil {
				return Response{}, fmt.Errorf("%w: %w", ErrFailedToMarshalBody, err)
			}
		}

		reqBody = bytes.NewBuffer(jsonData)
		hasJSONBody = true
	}

	req, err := http.NewRequestWithContext(ctx, r.method, r.url, reqBody)
	if err != nil {
		return Response{}, fmt.Errorf("%w: %w", ErrFailedToCreateRequest, err)
	}

	if hasJSONBody {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, values := range r.header {
		req.Header[key] = values
	}

	if len(r.query) > 0 {
		query := req.URL.Query()
		for key, values := range r.query {
			for _, value := range values {
				query.Add(key, value)
			}
		}

		req.URL.RawQuery = query.Encode()
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("%w: %w", ErrFailedToExecuteRequest, err)
	}

	defer closeBody(ctx, resp.Body)

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

// Cookie returns the value of the cookie name that the response set, or "" when it set none.
func (r Response) Cookie(name string) string {
	for _, cookie := range (&http.Response{Header: r.Header}).Cookies() {
		if cookie.Name == name {
			return cookie.Value
		}
	}

	return ""
}
