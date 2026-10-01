package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
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

	mockSequenceRoundTripper struct {
		mu      sync.Mutex
		results []MockHTTPResult
		next    int
	}

	// MockHTTPClient mocks HTTP client.
	MockHTTPClient struct {
		mock.Mock
	}
)

func MockSequenceRoundTripper(results ...MockHTTPResult) http.RoundTripper {
	return &mockSequenceRoundTripper{results: results}
}

func (m *mockSequenceRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
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

func (m *MockHTTPClient) NewRequest(ctx context.Context, method, url string, body interface{}) (Response, error) {
	args := m.Called(ctx, method, url, body)
	return args.Get(0).(Response), args.Error(1)
}

// MockDatabasePing builds a DatabasePing that always returns err.
func MockDatabasePing(err error) DatabasePing {
	return func(context.Context) error { return err }
}

// MockMigrationRunner builds a MigrationRunner that always returns err.
func MockMigrationRunner(err error) MigrationRunner {
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
