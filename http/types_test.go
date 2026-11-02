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

func TestRequest_Call_successWhenBodyIsJSON(t *testing.T) {
	client := _http.MockClientExpecting(_http.MockExpectedRequest{
		URL:    "http://example.test",
		Header: http.Header{"Content-Type": {"application/json"}},
		Body:   `{"symbol":"BTC/USDT"}`,
	})
	request := client.NewRequest(http.MethodPost, "http://example.test")
	request.WithBody(map[string]string{"symbol": "BTC/USDT"})

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_Call_successWhenResponseHasBody(t *testing.T) {
	mockServer := _http.MockServer(http.StatusOK, nil, `{"ok":true}`)
	defer mockServer.Close()
	client := _http.NewClient(5 * time.Second)

	resp, _ := client.NewRequest(http.MethodGet, mockServer.URL).Call(context.Background())

	want := `{"ok":true}`
	got := resp.Body

	assert.Equal(t, want, got)
}

func TestRequest_Call_successWhenResponseHasStatus(t *testing.T) {
	mockServer := _http.MockServer(http.StatusOK, nil, "")
	defer mockServer.Close()
	client := _http.NewClient(5 * time.Second)

	resp, _ := client.NewRequest(http.MethodGet, mockServer.URL).Call(context.Background())

	want := "200 OK"
	got := resp.Status

	assert.Equal(t, want, got)
}

func TestRequest_Call_successWhenBodyIsNil(t *testing.T) {
	client := _http.MockClientExpecting(_http.MockExpectedRequest{URL: "http://example.test?symbol=BTCUSDT", Header: http.Header{}})

	_, got := client.NewRequest(http.MethodGet, "http://example.test?symbol=BTCUSDT").Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_Call_successWhenResponseHasHeaders(t *testing.T) {
	header := http.Header{"X-Mbx-Used-Weight-1m": {"42"}, "Retry-After": {"5"}}
	mockServer := _http.MockServer(http.StatusTooManyRequests, header, "")
	defer mockServer.Close()
	client := _http.NewClient(5 * time.Second)

	resp, _ := client.NewRequest(http.MethodGet, mockServer.URL).Call(context.Background())

	want := "42"
	got := resp.Header.Get("X-MBX-USED-WEIGHT-1M")

	assert.Equal(t, want, got)
}

func TestRequest_Call_successWhenResponseHasRetryAfterHeader(t *testing.T) {
	header := http.Header{"Retry-After": {"5"}}
	mockServer := _http.MockServer(http.StatusTooManyRequests, header, "")
	defer mockServer.Close()
	client := _http.NewClient(5 * time.Second)

	resp, _ := client.NewRequest(http.MethodGet, mockServer.URL).Call(context.Background())

	want := http.StatusTooManyRequests
	got := resp.StatusCode

	assert.Equal(t, want, got)
}

func TestRequest_WithHeader_success(t *testing.T) {
	client := _http.MockClientExpecting(_http.MockExpectedRequest{URL: "http://example.test", Header: http.Header{"Authorization": {"Bearer token"}}})
	request := client.NewRequest(http.MethodGet, "http://example.test")
	request.WithHeader("Authorization", "Bearer token")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithHeader_successWhenTheHeaderIsSetTwice(t *testing.T) {
	client := _http.MockClientExpecting(_http.MockExpectedRequest{URL: "http://example.test", Header: http.Header{"X-Mode": {"second"}}})
	request := client.NewRequest(http.MethodGet, "http://example.test")
	request.WithHeader("X-Mode", "first")
	request.WithHeader("X-Mode", "second")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithQueryParam_success(t *testing.T) {
	client := _http.MockClientExpecting(_http.MockExpectedRequest{URL: "http://example.test/candles?pair=BTC%2FUSDT&timeframe=1h", Header: http.Header{}})
	request := client.NewRequest(http.MethodGet, "http://example.test/candles")
	request.WithQueryParam("pair", "BTC/USDT")
	request.WithQueryParam("timeframe", "1h")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithQueryParam_successWhenTheURLAlreadyHasAQuery(t *testing.T) {
	client := _http.MockClientExpecting(_http.MockExpectedRequest{URL: "http://example.test/candles?pair=BTC&pair=ETH", Header: http.Header{}})
	request := client.NewRequest(http.MethodGet, "http://example.test/candles?pair=BTC")
	request.WithQueryParam("pair", "ETH")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithBody_success(t *testing.T) {
	client := _http.MockClientExpecting(_http.MockExpectedRequest{
		URL:    "http://example.test",
		Header: http.Header{"Authorization": {"Bearer token"}, "Content-Type": {"application/json"}},
		Body:   `{"a":"b"}`,
	})
	request := client.NewRequest(http.MethodPost, "http://example.test")
	request.WithBody(map[string]string{"a": "b"})
	request.WithHeader("Authorization", "Bearer token")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithBody_successWhenTheBodyIsBytes(t *testing.T) {
	client := _http.MockClientExpecting(_http.MockExpectedRequest{
		URL:    "http://example.test",
		Header: http.Header{"Content-Type": {"application/json"}},
		Body:   `{"raw":true}`,
	})
	request := client.NewRequest(http.MethodPost, "http://example.test")
	request.WithBody([]byte(`{"raw":true}`))

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_Call_failsWhenTheBodyCannotBeMarshaled(t *testing.T) {
	client, _ := _http.MockSequenceClient()
	request := client.NewRequest(http.MethodPost, "http://example.test")
	request.WithBody(make(chan int))

	want := _http.ErrFailedToMarshalBody
	_, got := request.Call(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestRequest_Call_failsWhenURLIsInvalid(t *testing.T) {
	client := _http.NewClient(5 * time.Second)

	want := _http.ErrFailedToCreateRequest
	_, got := client.NewRequest(http.MethodGet, "://invalid-url").Call(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestRequest_Call_failsWhenContextIsCanceled(t *testing.T) {
	mockServer := _http.MockServer(http.StatusOK, nil, "")
	defer mockServer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := _http.NewClient(5 * time.Second)

	want := _http.ErrFailedToExecuteRequest
	wantCause := context.Canceled
	_, got := client.NewRequest(http.MethodGet, mockServer.URL).Call(ctx)

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
