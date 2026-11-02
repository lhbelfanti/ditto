package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"time"
)

type (
	// MockHTTPResult describes one response from MockSequenceRoundTripper.
	MockHTTPResult struct {
		StatusCode int
		Body       string
		Header     http.Header
		Err        error
	}

	// MockSequenceTransport replays a fixed sequence of MockHTTPResult values, one per request.
	MockSequenceTransport struct {
		mu       sync.Mutex
		results  []MockHTTPResult
		requests []*http.Request
		next     int
	}

	// MockExpectedRequest is the request a MockExpectingTransport expects: its URL, headers and body.
	MockExpectedRequest struct {
		URL    string
		Header http.Header
		Body   string
	}

	// MockExpectingTransport answers only the request it expects.
	MockExpectingTransport struct {
		expected MockExpectedRequest
	}
)

// MockSequenceRoundTripper returns a transport that answers each request with the next result, and fails once they run out.
func MockSequenceRoundTripper(results ...MockHTTPResult) *MockSequenceTransport {
	return &MockSequenceTransport{results: results}
}

// MockSequenceClient returns a client whose requests are answered by results in sequence, and the transport that records them.
func MockSequenceClient(results ...MockHTTPResult) (*CustomClient, *MockSequenceTransport) {
	transport := MockSequenceRoundTripper(results...)
	client := NewClient(0)
	client.HTTPClient.Transport = transport

	return client, transport
}

// RoundTrip answers with the next result of the sequence, or fails when the sequence is exhausted.
func (m *MockSequenceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests = append(m.requests, req)

	if m.next >= len(m.results) {
		return nil, fmt.Errorf("mock HTTP response sequence exhausted after %d requests", m.next)
	}

	result := m.results[m.next]
	m.next++
	if result.Err != nil {
		return nil, result.Err
	}

	return &http.Response{
		Status:     fmt.Sprintf("%d %s", result.StatusCode, http.StatusText(result.StatusCode)),
		StatusCode: result.StatusCode,
		Body:       io.NopCloser(strings.NewReader(result.Body)),
		Header:     result.Header,
	}, nil
}

// Requests returns the requests the transport received, in order, including the one that found the sequence exhausted.
func (m *MockSequenceTransport) Requests() []*http.Request {
	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]*http.Request(nil), m.requests...)
}

// MockDatabasePing builds a DatabasePing that always returns err.
func MockDatabasePing(err error) DatabasePing {
	return func(context.Context) error { return err }
}

// MockServe builds a Serve that returns err immediately.
func MockServe(err error) Serve {
	return func() error { return err }
}

// MockServeBlocking builds a Serve that never returns, for exercising a caller's ctx-cancellation
// path without racing a real server lifecycle.
func MockServeBlocking() Serve {
	return func() error {
		select {}
	}
}

// MockShutdown builds a Shutdown that always returns err.
func MockShutdown(err error) Shutdown {
	return func(context.Context) error { return err }
}

// MockShutdownExpectingTimeout builds a Shutdown that returns nil when its context ends within a
// second of timeout from now, and an error otherwise.
func MockShutdownExpectingTimeout(timeout time.Duration) Shutdown {
	return func(ctx context.Context) error {
		deadline, hasDeadline := ctx.Deadline()
		remaining := time.Until(deadline)
		if !hasDeadline || remaining > timeout || remaining < timeout-time.Second {
			return fmt.Errorf("unexpected shutdown deadline: %s", deadline)
		}

		return nil
	}
}

// MockServer starts a test server that answers with status, header and body. The caller closes it.
func MockServer(status int, header http.Header, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for key, values := range header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

// MockClientExpecting returns a client whose requests succeed when they are the expected one, and fail otherwise.
func MockClientExpecting(expected MockExpectedRequest) *CustomClient {
	client := NewClient(0)
	client.HTTPClient.Transport = &MockExpectingTransport{expected: expected}

	return client
}

// RoundTrip answers 200 when req is the expected request, and fails otherwise.
func (m *MockExpectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body := []byte{}
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
	}

	received := MockExpectedRequest{
		URL:    req.URL.String(),
		Header: req.Header,
		Body:   string(body),
	}
	if !reflect.DeepEqual(m.expected, received) {
		return nil, fmt.Errorf("unexpected request: %+v", received)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
		Header:     http.Header{},
	}, nil
}
