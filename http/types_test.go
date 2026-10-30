package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestCustomClient_Send_successWhenBodyIsJSON(t *testing.T) {
	mockServer, mockRequest := _http.MockServer(http.StatusOK, nil, `{"ok":true}`)
	defer mockServer.Close()
	client := _http.NewClient(5 * time.Second)

	_, _ = client.Send(context.Background(), http.MethodPost, mockServer.URL, map[string]string{"symbol": "BTC/USDT"})

	want := _http.MockServerRequest{ContentType: "application/json", Body: `{"symbol":"BTC/USDT"}`}
	got := *mockRequest

	assert.Equal(t, want, got)
}

func TestCustomClient_Send_successWhenResponseHasBody(t *testing.T) {
	mockServer, _ := _http.MockServer(http.StatusOK, nil, `{"ok":true}`)
	defer mockServer.Close()
	client := _http.NewClient(5 * time.Second)

	resp, _ := client.Send(context.Background(), http.MethodGet, mockServer.URL, nil)

	want := `{"ok":true}`
	got := resp.Body

	assert.Equal(t, want, got)
}

func TestCustomClient_Send_successWhenResponseHasStatus(t *testing.T) {
	mockServer, _ := _http.MockServer(http.StatusOK, nil, "")
	defer mockServer.Close()
	client := _http.NewClient(5 * time.Second)

	resp, _ := client.Send(context.Background(), http.MethodGet, mockServer.URL, nil)

	want := "200 OK"
	got := resp.Status

	assert.Equal(t, want, got)
}

func TestCustomClient_Send_successWhenBodyIsNil(t *testing.T) {
	mockServer, mockRequest := _http.MockServer(http.StatusOK, nil, "")
	defer mockServer.Close()
	client := _http.NewClient(5 * time.Second)

	_, _ = client.Send(context.Background(), http.MethodGet, mockServer.URL+"?symbol=BTCUSDT", nil)

	want := _http.MockServerRequest{}
	got := *mockRequest

	assert.Equal(t, want, got)
}

func TestCustomClient_Send_successWhenResponseHasHeaders(t *testing.T) {
	header := http.Header{"X-Mbx-Used-Weight-1m": {"42"}, "Retry-After": {"5"}}
	mockServer, _ := _http.MockServer(http.StatusTooManyRequests, header, "")
	defer mockServer.Close()
	client := _http.NewClient(5 * time.Second)

	resp, _ := client.Send(context.Background(), http.MethodGet, mockServer.URL, nil)

	want := "42"
	got := resp.Header.Get("X-MBX-USED-WEIGHT-1M")

	assert.Equal(t, want, got)
}

func TestCustomClient_Send_successWhenResponseHasRetryAfterHeader(t *testing.T) {
	header := http.Header{"Retry-After": {"5"}}
	mockServer, _ := _http.MockServer(http.StatusTooManyRequests, header, "")
	defer mockServer.Close()
	client := _http.NewClient(5 * time.Second)

	resp, _ := client.Send(context.Background(), http.MethodGet, mockServer.URL, nil)

	want := http.StatusTooManyRequests
	got := resp.StatusCode

	assert.Equal(t, want, got)
}

func TestCustomClient_Send_successWhenOptionSetsHeader(t *testing.T) {
	transport := _http.MockSequenceRoundTripper(_http.MockHTTPResult{StatusCode: http.StatusOK})
	client := _http.NewClient(0)
	client.HTTPClient.Transport = transport

	_, _ = client.Send(context.Background(), http.MethodGet, "http://example.test", nil, _http.WithHeader("Authorization", "Bearer token"))

	want := "Bearer token"
	got := transport.Requests()[0].Header.Get("Authorization")

	assert.Equal(t, want, got)
}

func TestCustomClient_Send_successWhenOptionsAreAppliedInOrder(t *testing.T) {
	transport := _http.MockSequenceRoundTripper(_http.MockHTTPResult{StatusCode: http.StatusOK})
	client := _http.NewClient(0)
	client.HTTPClient.Transport = transport

	_, _ = client.Send(context.Background(), http.MethodGet, "http://example.test", nil, _http.WithHeader("X-Mode", "first"), _http.WithHeader("X-Mode", "second"))

	want := "second"
	got := transport.Requests()[0].Header.Get("X-Mode")

	assert.Equal(t, want, got)
}

func TestCustomClient_Send_successWhenBodyIsJSONAndOptionSetsHeader(t *testing.T) {
	transport := _http.MockSequenceRoundTripper(_http.MockHTTPResult{StatusCode: http.StatusOK})
	client := _http.NewClient(0)
	client.HTTPClient.Transport = transport

	_, _ = client.Send(context.Background(), http.MethodPost, "http://example.test", map[string]string{"a": "b"}, _http.WithHeader("Authorization", "Bearer token"))

	want := "application/json"
	got := transport.Requests()[0].Header.Get("Content-Type")

	assert.Equal(t, want, got)
}

func TestCustomClient_Send_failsWhenURLIsInvalid(t *testing.T) {
	client := _http.NewClient(5 * time.Second)

	want := _http.ErrFailedToCreateRequest
	_, got := client.Send(context.Background(), http.MethodGet, "://invalid-url", nil)

	assert.ErrorIs(t, got, want)
}

func TestCustomClient_Send_failsWhenContextIsCanceled(t *testing.T) {
	mockServer, _ := _http.MockServer(http.StatusOK, nil, "")
	defer mockServer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := _http.NewClient(5 * time.Second)

	want := _http.ErrFailedToExecuteRequest
	wantCause := context.Canceled
	_, got := client.Send(ctx, http.MethodGet, mockServer.URL, nil)

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestSystemRoutes_WithDatabasePing_success(t *testing.T) {
	mockPing := _http.MockDatabasePing(nil)
	mux := http.NewServeMux()
	_http.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestSystemRoutes_WithDatabasePing_failsWhenPingFails(t *testing.T) {
	mockPing := _http.MockDatabasePing(errors.New("database unreachable"))
	mux := http.NewServeMux()
	_http.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusServiceUnavailable
	got := w.Code

	assert.Equal(t, want, got)
}

func TestSystemRoutes_WithDatabasePing_failsWhenPingFailsWithSensitiveError(t *testing.T) {
	underlyingErrText := "pgx: connection refused on host db-internal:5432"
	mockPing := _http.MockDatabasePing(errors.New(underlyingErrText))
	mux := http.NewServeMux()
	_http.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	want := `{"code":503,"message":"` + _http.ErrMsgDatabaseUnavailable + `"}` + "\n"
	got := w.Body.String()

	assert.Equal(t, want, got)
}
