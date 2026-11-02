package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

type (
	// MockHTTPResult describes one response from MockSequenceRoundTripper.
	MockHTTPResult struct {
		StatusCode int
		Body       string
		Header     http.Header
		Err        error
	}

	// CaptureTransport replays a fixed sequence of MockHTTPResult values, one per request.
	CaptureTransport struct {
		mu       sync.Mutex
		results  []MockHTTPResult
		requests []*http.Request
		next     int
	}

	// CapturedRequest records the last request a MockServer received.
	CapturedRequest struct {
		ContentType string
		Body        string
	}
)

// CaptureSequenceRoundTripper returns a transport that answers each request with the next result, and fails once they run out.
func CaptureSequenceRoundTripper(results ...MockHTTPResult) *CaptureTransport {
	return &CaptureTransport{results: results}
}

// CaptureSequenceClient returns a client whose requests are answered by results in sequence, and the transport that records them.
func CaptureSequenceClient(results ...MockHTTPResult) (*CustomClient, *CaptureTransport) {
	transport := CaptureSequenceRoundTripper(results...)
	client := NewClient(0)
	client.HTTPClient.Transport = transport

	return client, transport
}

// RoundTrip answers with the next result of the sequence, or fails when the sequence is exhausted.
func (m *CaptureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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
func (m *CaptureTransport) Requests() []*http.Request {
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

// CaptureShutdown builds a Shutdown that returns err and records the ctx of its most
// recent call into *captured.
func CaptureShutdown(err error, captured *context.Context) Shutdown {
	return func(ctx context.Context) error {
		*captured = ctx
		return err
	}
}

// CaptureServer starts a test server that records the request it receives and answers with status, header and body. The caller closes it.
func CaptureServer(status int, header http.Header, body string) (*httptest.Server, *CapturedRequest) {
	received := &CapturedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.ContentType = r.Header.Get("Content-Type")
		requestBody, _ := io.ReadAll(r.Body)
		received.Body = string(requestBody)
		for key, values := range header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	return server, received
}
