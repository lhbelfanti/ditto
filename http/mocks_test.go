package http_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestCaptureSequenceRoundTripper_success(t *testing.T) {
	client := _http.NewClient(0)
	client.HTTPClient.Transport = _http.CaptureSequenceRoundTripper(
		_http.MockHTTPResult{StatusCode: http.StatusTooManyRequests, Body: "retry"},
		_http.MockHTTPResult{StatusCode: http.StatusOK, Body: "done"},
	)
	_, _ = client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	second, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	want := "done"
	got := second.Body

	assert.Equal(t, want, got)
}

func TestCaptureSequenceRoundTripper_successWhenResultHasHeaders(t *testing.T) {
	client := _http.NewClient(0)
	client.HTTPClient.Transport = _http.CaptureSequenceRoundTripper(
		_http.MockHTTPResult{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"1"}}},
	)

	first, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	want := "1"
	got := first.Header.Get("Retry-After")

	assert.Equal(t, want, got)
}

func TestCaptureSequenceRoundTripper_failsWhenSequenceIsExhausted(t *testing.T) {
	client := _http.NewClient(0)
	client.HTTPClient.Transport = _http.CaptureSequenceRoundTripper()

	want := _http.ErrFailedToExecuteRequest
	_, got := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestCaptureTransport_Requests_success(t *testing.T) {
	transport := _http.CaptureSequenceRoundTripper(_http.MockHTTPResult{StatusCode: http.StatusOK}, _http.MockHTTPResult{StatusCode: http.StatusOK})
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

func TestCaptureTransport_Requests_successWhenSequenceIsExhausted(t *testing.T) {
	transport := _http.CaptureSequenceRoundTripper()
	client := _http.NewClient(0)
	client.HTTPClient.Transport = transport
	_, _ = client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	want := 1
	got := len(transport.Requests())

	assert.Equal(t, want, got)
}

func TestCaptureSequenceClient_success(t *testing.T) {
	client, _ := _http.CaptureSequenceClient(_http.MockHTTPResult{StatusCode: http.StatusOK, Body: "done"})

	response, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	want := "done"
	got := response.Body

	assert.Equal(t, want, got)
}

func TestCaptureSequenceClient_successWhenReadingTheRecordedRequests(t *testing.T) {
	client, transport := _http.CaptureSequenceClient(_http.MockHTTPResult{StatusCode: http.StatusOK})
	request := client.NewRequest(http.MethodGet, "http://example.test/path")
	request.WithHeader("Authorization", "Bearer token")
	_, _ = request.Call(context.Background())

	want := "Bearer token"
	got := transport.Requests()[0].Header.Get("Authorization")

	assert.Equal(t, want, got)
}

func TestCaptureSequenceClient_failsWhenSequenceIsExhausted(t *testing.T) {
	client, _ := _http.CaptureSequenceClient()

	want := _http.ErrFailedToExecuteRequest
	_, got := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.ErrorIs(t, got, want)
}
