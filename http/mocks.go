package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/stretchr/testify/mock"
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
		mu      sync.Mutex
		results []MockHTTPResult
		next    int
	}

	// MockServerRequest records the last request a MockServer received.
	MockServerRequest struct {
		ContentType string
		Body        string
	}

	// MockHTTPClient mocks HTTP client.
	MockHTTPClient struct {
		mock.Mock
	}
)

// MockSequenceRoundTripper returns a transport that answers each request with the next result, and fails once they run out.
func MockSequenceRoundTripper(results ...MockHTTPResult) http.RoundTripper {
	return &MockSequenceTransport{results: results}
}

// RoundTrip answers with the next result of the sequence, or fails when the sequence is exhausted.
func (m *MockSequenceTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

// NewRequest returns the response and error configured with On.
func (m *MockHTTPClient) NewRequest(ctx context.Context, method, url string, body any) (Response, error) {
	args := m.Called(ctx, method, url, body)
	return args.Get(0).(Response), args.Error(1)
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

// MockCountingShutdown builds a Shutdown that returns err and increments *calls on every call.
func MockCountingShutdown(err error, calls *int) Shutdown {
	return func(context.Context) error {
		*calls++
		return err
	}
}

// MockCapturingShutdown builds a Shutdown that returns err and records the ctx of its most
// recent call into *captured.
func MockCapturingShutdown(err error, captured *context.Context) Shutdown {
	return func(ctx context.Context) error {
		*captured = ctx
		return err
	}
}

// MockServer starts a test server that records the request it receives and answers with status, header and body. The caller closes it.
func MockServer(status int, header http.Header, body string) (*httptest.Server, *MockServerRequest) {
	received := &MockServerRequest{}
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
