package http_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestMockSequenceRoundTripper_success(t *testing.T) {
	client := _http.NewClient(0)
	mockFirstHTTPResult := _http.MockHTTPResult{StatusCode: http.StatusTooManyRequests, Body: "retry"}
	mockSecondHTTPResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: "done"}
	client.HTTPClient.Transport = _http.MockSequenceRoundTripper(
		mockFirstHTTPResult,
		mockSecondHTTPResult,
	)
	_, _ = client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	want := _http.Response{Body: "done", Status: "200 OK", StatusCode: http.StatusOK}
	got, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.Equal(t, want, got)
}

func TestMockSequenceRoundTripper_successWhenResultHasHeaders(t *testing.T) {
	client := _http.NewClient(0)
	mockHTTPResult := _http.MockHTTPResult{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"1"}}}
	client.HTTPClient.Transport = _http.MockSequenceRoundTripper(
		mockHTTPResult,
	)

	want := _http.Response{Status: "429 Too Many Requests", StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"1"}}}
	got, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.Equal(t, want, got)
}

func TestMockSequenceRoundTripper_failsWhenSequenceIsExhausted(t *testing.T) {
	client := _http.NewClient(0)
	client.HTTPClient.Transport = _http.MockSequenceRoundTripper()

	want := _http.ErrFailedToExecuteRequest
	_, got := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMockSequenceTransport_Requests_success(t *testing.T) {
	mockFirstHTTPResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	mockSecondHTTPResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	transport := _http.MockSequenceRoundTripper(mockFirstHTTPResult, mockSecondHTTPResult)
	client := _http.NewClient(0)
	client.HTTPClient.Transport = transport
	_, _ = client.NewRequest(http.MethodGet, "http://example.test/first").Call(context.Background())
	_, _ = client.NewRequest(http.MethodGet, "http://example.test/second").Call(context.Background())

	urls := []string{}
	for _, req := range transport.Requests() {
		urls = append(urls, req.URL.String())
	}

	want := []string{"http://example.test/first", "http://example.test/second"}
	got := urls

	assert.Equal(t, want, got)
}

func TestMockSequenceTransport_Requests_successWhenSequenceIsExhausted(t *testing.T) {
	transport := _http.MockSequenceRoundTripper()
	client := _http.NewClient(0)
	client.HTTPClient.Transport = transport
	_, _ = client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	want := 1
	got := len(transport.Requests())

	assert.Equal(t, want, got)
}

func TestMockSequenceClient_success(t *testing.T) {
	mockHTTPResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: "done"}
	client, _ := _http.MockSequenceClient(mockHTTPResult)

	want := _http.Response{Body: "done", Status: "200 OK", StatusCode: http.StatusOK}
	got, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.Equal(t, want, got)
}

func TestMockSequenceClient_successWhenReadingTheRecordedRequests(t *testing.T) {
	mockHTTPResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	client, transport := _http.MockSequenceClient(mockHTTPResult)
	request := client.NewRequest(http.MethodGet, "http://example.test/path")
	request.WithHeader("Authorization", "Bearer token")
	_, _ = request.Call(context.Background())

	headers := []http.Header{}
	for _, req := range transport.Requests() {
		headers = append(headers, req.Header)
	}

	want := []http.Header{{"Authorization": []string{"Bearer token"}}}
	got := headers

	assert.Equal(t, want, got)
}

func TestMockSequenceClient_failsWhenSequenceIsExhausted(t *testing.T) {
	client, _ := _http.MockSequenceClient()

	want := _http.ErrFailedToExecuteRequest
	_, got := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMockClientExpecting_success(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test", Header: http.Header{}}
	client := _http.MockClientExpecting(mockExpectedRequest)

	_, got := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.NoError(t, got)
}

func TestMockClientExpecting_failsWhenRequestIsUnexpected(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test/other", Header: http.Header{}}
	client := _http.MockClientExpecting(mockExpectedRequest)

	want := _http.ErrFailedToExecuteRequest
	_, got := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMockClientAnswering_success(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test", Header: http.Header{}}
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusUnauthorized, Body: `{"code":401}`}
	client := _http.MockClientAnswering(mockExpectedRequest, mockResult)

	want := _http.Response{Body: `{"code":401}`, Status: "401 Unauthorized", StatusCode: http.StatusUnauthorized}
	got, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.Equal(t, want, got)
}

func TestMockClientAnswering_failsWhenRequestIsUnexpected(t *testing.T) {
	mockExpectedRequest := _http.MockExpectedRequest{URL: "http://example.test/other", Header: http.Header{}}
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	client := _http.MockClientAnswering(mockExpectedRequest, mockResult)

	want := _http.ErrFailedToExecuteRequest
	_, got := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMockShutdownExpectingTimeout_failsWhenContextHasNoDeadline(t *testing.T) {
	mockShutdown := _http.MockShutdownExpectingTimeout(time.Second)

	got := mockShutdown(context.Background())

	assert.Error(t, got)
}

func TestMockShutdownExpectingTimeout_failsWhenTimeoutIsUnexpected(t *testing.T) {
	mockShutdown := _http.MockShutdownExpectingTimeout(time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got := mockShutdown(ctx)

	assert.Error(t, got)
}
