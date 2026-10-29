package http_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestMockSequenceRoundTripper_success(t *testing.T) {
	client := _http.NewClient(0)
	client.HTTPClient.Transport = _http.MockSequenceRoundTripper(
		_http.MockHTTPResult{StatusCode: http.StatusTooManyRequests, Body: "retry"},
		_http.MockHTTPResult{StatusCode: http.StatusOK, Body: "done"},
	)
	_, _ = client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	second, _ := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	want := "done"
	got := second.Body

	assert.Equal(t, want, got)
}

func TestMockSequenceRoundTripper_successWhenResultHasHeaders(t *testing.T) {
	client := _http.NewClient(0)
	client.HTTPClient.Transport = _http.MockSequenceRoundTripper(
		_http.MockHTTPResult{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"1"}}},
	)

	first, _ := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	want := "1"
	got := first.Header.Get("Retry-After")

	assert.Equal(t, want, got)
}

func TestMockSequenceRoundTripper_failsWhenSequenceIsExhausted(t *testing.T) {
	client := _http.NewClient(0)
	client.HTTPClient.Transport = _http.MockSequenceRoundTripper()

	want := _http.ErrFailedToExecuteRequest
	_, got := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	assert.ErrorIs(t, got, want)
}

func TestMockSequenceTransport_Requests_success(t *testing.T) {
	transport := _http.MockSequenceRoundTripper(_http.MockHTTPResult{StatusCode: http.StatusOK}, _http.MockHTTPResult{StatusCode: http.StatusOK})
	client := _http.NewClient(0)
	client.HTTPClient.Transport = transport
	_, _ = client.NewRequest(context.Background(), http.MethodGet, "http://example.test/first", nil)
	_, _ = client.NewRequest(context.Background(), http.MethodGet, "http://example.test/second", nil)

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
	_, _ = client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	want := 1
	got := len(transport.Requests())

	assert.Equal(t, want, got)
}

func TestMockHTTPClient_NewRequest_successWhenCalledWithOptions(t *testing.T) {
	mockClient := &_http.MockHTTPClient{}
	mockClient.On("NewRequest", context.Background(), http.MethodGet, "http://example.test", nil).Return(_http.Response{Body: "ok"}, nil)

	resp, _ := mockClient.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil, _http.WithHeader("Authorization", "Bearer token"))

	want := "ok"
	got := resp.Body

	assert.Equal(t, want, got)
}

func TestMockSequenceClient_success(t *testing.T) {
	client, _ := _http.MockSequenceClient(_http.MockHTTPResult{StatusCode: http.StatusOK, Body: "done"})

	response, _ := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	want := "done"
	got := response.Body

	assert.Equal(t, want, got)
}

func TestMockSequenceClient_successWhenReadingTheRecordedRequests(t *testing.T) {
	client, transport := _http.MockSequenceClient(_http.MockHTTPResult{StatusCode: http.StatusOK})
	_, _ = client.NewRequest(context.Background(), http.MethodGet, "http://example.test/path", nil, _http.WithHeader("Authorization", "Bearer token"))

	want := "Bearer token"
	got := transport.Requests()[0].Header.Get("Authorization")

	assert.Equal(t, want, got)
}

func TestMockSequenceClient_failsWhenSequenceIsExhausted(t *testing.T) {
	client, _ := _http.MockSequenceClient()

	want := _http.ErrFailedToExecuteRequest
	_, got := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	assert.ErrorIs(t, got, want)
}
