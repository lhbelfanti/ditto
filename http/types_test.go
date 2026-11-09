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
	"github.com/lhbelfanti/ditto/v2/http/response"
)

func TestRequest_Call_successWhenBodyIsJSON(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{
		URL:    "http://example.test",
		Header: http.Header{"Content-Type": {"application/json"}},
		Body:   `{"symbol":"BTC/USDT"}`,
	}
	client := _http.MockClientExpecting(mockExpectedRequest)
	request := client.NewRequest(http.MethodPost, "http://example.test")
	request.WithBody(map[string]string{"symbol": "BTC/USDT"})

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_Call_successWhenResponseHasBody(t *testing.T) {
	mockHTTPResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: `{"ok":true}`}
	client, _ := _http.MockSequenceClient(mockHTTPResult)

	want := _http.Response{Body: `{"ok":true}`, Status: "200 OK", StatusCode: http.StatusOK}
	got, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.Equal(t, want, got)
}

func TestRequest_Call_successWhenResponseHasStatus(t *testing.T) {
	mockHTTPResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	client, _ := _http.MockSequenceClient(mockHTTPResult)

	want := _http.Response{Status: "200 OK", StatusCode: http.StatusOK}
	got, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.Equal(t, want, got)
}

func TestRequest_Call_successWhenBodyIsNil(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test?symbol=BTCUSDT", Header: http.Header{}}
	client := _http.MockClientExpecting(mockExpectedRequest)

	_, got := client.NewRequest(http.MethodGet, "http://example.test?symbol=BTCUSDT").Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_Call_successWhenResponseHasHeaders(t *testing.T) {
	mockHeader := http.Header{"X-Mbx-Used-Weight-1m": {"42"}, "Retry-After": {"5"}}
	mockHTTPResult := _http.MockHTTPResult{StatusCode: http.StatusTooManyRequests, Header: mockHeader}
	client, _ := _http.MockSequenceClient(mockHTTPResult)

	want := _http.Response{Status: "429 Too Many Requests", StatusCode: http.StatusTooManyRequests, Header: http.Header{"X-Mbx-Used-Weight-1m": {"42"}, "Retry-After": {"5"}}}
	got, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.Equal(t, want, got)
}

func TestRequest_WithHeader_success(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test", Header: http.Header{"Authorization": {"Bearer token"}}}
	client := _http.MockClientExpecting(mockExpectedRequest)
	request := client.NewRequest(http.MethodGet, "http://example.test")
	request.WithHeader("Authorization", "Bearer token")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithHeader_successWhenTheHeaderIsSetTwice(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test", Header: http.Header{"X-Mode": {"second"}}}
	client := _http.MockClientExpecting(mockExpectedRequest)
	request := client.NewRequest(http.MethodGet, "http://example.test")
	request.WithHeader("X-Mode", "first")
	request.WithHeader("X-Mode", "second")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithQueryParam_success(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test/candles?pair=BTC%2FUSDT&timeframe=1h", Header: http.Header{}}
	client := _http.MockClientExpecting(mockExpectedRequest)
	request := client.NewRequest(http.MethodGet, "http://example.test/candles")
	request.WithQueryParam("pair", "BTC/USDT")
	request.WithQueryParam("timeframe", "1h")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithQueryParam_successWhenTheURLAlreadyHasAQuery(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test/candles?pair=BTC&pair=ETH", Header: http.Header{}}
	client := _http.MockClientExpecting(mockExpectedRequest)
	request := client.NewRequest(http.MethodGet, "http://example.test/candles?pair=BTC")
	request.WithQueryParam("pair", "ETH")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithBearer_success(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test", Header: http.Header{"Authorization": {"Bearer token"}}}
	client := _http.MockClientExpecting(mockExpectedRequest)
	request := client.NewRequest(http.MethodGet, "http://example.test")
	request.WithBearer("token")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithCookies_success(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test", Header: http.Header{"Cookie": {"state=abc; nonce=xyz"}}}
	client := _http.MockClientExpecting(mockExpectedRequest)
	mockResponse := _http.Response{Header: http.Header{"Set-Cookie": {"state=abc; Path=/; HttpOnly", "nonce=xyz; Path=/"}}}
	request := client.NewRequest(http.MethodGet, "http://example.test")
	request.WithCookies(mockResponse)

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithCookies_successWhenTheResponseSetNoCookie(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test", Header: http.Header{}}
	client := _http.MockClientExpecting(mockExpectedRequest)
	request := client.NewRequest(http.MethodGet, "http://example.test")
	request.WithCookies(_http.Response{Header: http.Header{}})

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestResponse_Cookie_success(t *testing.T) {
	response := _http.Response{Header: http.Header{"Set-Cookie": {"state=; Max-Age=0", "session=token; Path=/; HttpOnly"}}}

	want := "token"
	got := response.Cookie("session")

	assert.Equal(t, want, got)
}

func TestResponse_Cookie_successWhenTheCookieIsNotSet(t *testing.T) {
	response := _http.Response{Header: http.Header{"Set-Cookie": {"state=; Max-Age=0"}}}

	want := ""
	got := response.Cookie("session")

	assert.Equal(t, want, got)
}

func TestRequest_WithBody_success(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{
		URL:    "http://example.test",
		Header: http.Header{"Authorization": {"Bearer token"}, "Content-Type": {"application/json"}},
		Body:   `{"a":"b"}`,
	}
	client := _http.MockClientExpecting(mockExpectedRequest)
	request := client.NewRequest(http.MethodPost, "http://example.test")
	request.WithBody(map[string]string{"a": "b"})
	request.WithHeader("Authorization", "Bearer token")

	_, got := request.Call(context.Background())

	assert.NoError(t, got)
}

func TestRequest_WithBody_successWhenTheBodyIsBytes(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{
		URL:    "http://example.test",
		Header: http.Header{"Content-Type": {"application/json"}},
		Body:   `{"raw":true}`,
	}
	client := _http.MockClientExpecting(mockExpectedRequest)
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
	_, got := client.NewRequest(http.MethodGet, mockServer.URL).Call(ctx)

	assert.ErrorIs(t, got, want)
}

func TestSystemRoutes_WithDatabasePing_success(t *testing.T) {
	mockPing := _http.MockDatabasePing(nil)
	mux := http.NewServeMux()
	_http.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := response.Recorded[any]{Status: http.StatusOK, Message: "pong"}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestSystemRoutes_WithDatabasePing_failsWhenPingFails(t *testing.T) {
	mockPing := _http.MockDatabasePing(errors.New("database unreachable"))
	mux := http.NewServeMux()
	_http.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := response.Recorded[any]{Status: http.StatusServiceUnavailable, Message: _http.ErrMsgDatabaseUnavailable}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestSystemRoutes_WithDatabasePing_failsWhenPingFailsWithSensitiveError(t *testing.T) {
	mockErrText := "pgx: connection refused on host db-internal:5432"
	mockPing := _http.MockDatabasePing(errors.New(mockErrText))
	mux := http.NewServeMux()
	_http.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	want := response.Recorded[any]{Status: http.StatusServiceUnavailable, Message: _http.ErrMsgDatabaseUnavailable}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}
