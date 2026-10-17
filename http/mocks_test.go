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

	want := _http.FailedToExecuteRequest
	_, got := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	assert.ErrorIs(t, got, want)
}
