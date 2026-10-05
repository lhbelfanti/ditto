package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	dittohttp "github.com/lhbelfanti/ditto/v2/http"
)

func TestCustomClient_NewRequest_successWhenBodyIsJSON(t *testing.T) {
	server, request := dittohttp.MockServer(http.StatusOK, nil, `{"ok":true}`)
	defer server.Close()
	client := dittohttp.NewClient(5 * time.Second)

	_, _ = client.NewRequest(context.Background(), http.MethodPost, server.URL, map[string]string{"symbol": "BTC/USDT"})

	want := `{"symbol":"BTC/USDT"}`
	got := request.Body

	assert.Equal(t, want, got)
}

func TestCustomClient_NewRequest_successWhenBodyIsJSONSetsContentType(t *testing.T) {
	server, request := dittohttp.MockServer(http.StatusOK, nil, "")
	defer server.Close()
	client := dittohttp.NewClient(5 * time.Second)

	_, _ = client.NewRequest(context.Background(), http.MethodPost, server.URL, map[string]string{"symbol": "BTC/USDT"})

	want := "application/json"
	got := request.ContentType

	assert.Equal(t, want, got)
}

func TestCustomClient_NewRequest_successWhenResponseHasBody(t *testing.T) {
	server, _ := dittohttp.MockServer(http.StatusOK, nil, `{"ok":true}`)
	defer server.Close()
	client := dittohttp.NewClient(5 * time.Second)

	resp, _ := client.NewRequest(context.Background(), http.MethodGet, server.URL, nil)

	want := `{"ok":true}`
	got := resp.Body

	assert.Equal(t, want, got)
}

func TestCustomClient_NewRequest_successWhenResponseHasStatus(t *testing.T) {
	server, _ := dittohttp.MockServer(http.StatusOK, nil, "")
	defer server.Close()
	client := dittohttp.NewClient(5 * time.Second)

	resp, _ := client.NewRequest(context.Background(), http.MethodGet, server.URL, nil)

	want := "200 OK"
	got := resp.Status

	assert.Equal(t, want, got)
}

func TestCustomClient_NewRequest_successWhenBodyIsNil(t *testing.T) {
	server, request := dittohttp.MockServer(http.StatusOK, nil, "")
	defer server.Close()
	client := dittohttp.NewClient(5 * time.Second)

	_, _ = client.NewRequest(context.Background(), http.MethodGet, server.URL+"?symbol=BTCUSDT", nil)

	want := ""
	got := request.Body

	assert.Equal(t, want, got)
}

func TestCustomClient_NewRequest_successWhenBodyIsNilOmitsContentType(t *testing.T) {
	server, request := dittohttp.MockServer(http.StatusOK, nil, "")
	defer server.Close()
	client := dittohttp.NewClient(5 * time.Second)

	_, _ = client.NewRequest(context.Background(), http.MethodGet, server.URL, nil)

	want := ""
	got := request.ContentType

	assert.Equal(t, want, got)
}

func TestCustomClient_NewRequest_successWhenResponseHasHeaders(t *testing.T) {
	header := http.Header{"X-Mbx-Used-Weight-1m": {"42"}, "Retry-After": {"5"}}
	server, _ := dittohttp.MockServer(http.StatusTooManyRequests, header, "")
	defer server.Close()
	client := dittohttp.NewClient(5 * time.Second)

	resp, _ := client.NewRequest(context.Background(), http.MethodGet, server.URL, nil)

	want := "42"
	got := resp.Header.Get("X-MBX-USED-WEIGHT-1M")

	assert.Equal(t, want, got)
}

func TestCustomClient_NewRequest_successWhenResponseHasRetryAfterHeader(t *testing.T) {
	header := http.Header{"Retry-After": {"5"}}
	server, _ := dittohttp.MockServer(http.StatusTooManyRequests, header, "")
	defer server.Close()
	client := dittohttp.NewClient(5 * time.Second)

	resp, _ := client.NewRequest(context.Background(), http.MethodGet, server.URL, nil)

	want := http.StatusTooManyRequests
	got := resp.StatusCode

	assert.Equal(t, want, got)
}

func TestCustomClient_NewRequest_failsWhenURLIsInvalid(t *testing.T) {
	client := dittohttp.NewClient(5 * time.Second)

	want := dittohttp.FailedToCreateRequest
	_, got := client.NewRequest(context.Background(), http.MethodGet, "://invalid-url", nil)

	assert.ErrorIs(t, got, want)
}

func TestCustomClient_NewRequest_failsWhenContextIsCanceled(t *testing.T) {
	server, _ := dittohttp.MockServer(http.StatusOK, nil, "")
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := dittohttp.NewClient(5 * time.Second)

	want := dittohttp.FailedToExecuteRequest
	_, got := client.NewRequest(ctx, http.MethodGet, server.URL, nil)

	assert.ErrorIs(t, got, want)
}

func TestCustomClient_NewRequest_failsWhenContextIsCanceledKeepsCause(t *testing.T) {
	server, _ := dittohttp.MockServer(http.StatusOK, nil, "")
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := dittohttp.NewClient(5 * time.Second)

	want := context.Canceled
	_, got := client.NewRequest(ctx, http.MethodGet, server.URL, nil)

	assert.ErrorIs(t, got, want)
}

func TestSystemRoutes_WithDatabasePing_success(t *testing.T) {
	mockPing := dittohttp.MockDatabasePing(nil)
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestSystemRoutes_WithDatabasePing_failsWhenPingFails(t *testing.T) {
	mockPing := dittohttp.MockDatabasePing(errors.New("database unreachable"))
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusServiceUnavailable
	got := w.Code

	assert.Equal(t, want, got)
}

func TestSystemRoutes_WithDatabasePing_failsWhenPingFailsWithSensitiveError(t *testing.T) {
	underlyingErrText := "pgx: connection refused on host db-internal:5432"
	mockPing := dittohttp.MockDatabasePing(errors.New(underlyingErrText))
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	got := w.Body.String()

	assert.NotContains(t, got, underlyingErrText)
}
