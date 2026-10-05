package http_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	dittohttp "github.com/lhbelfanti/ditto/v2/http"
)

func TestMockSequenceRoundTripper_success(t *testing.T) {
	client := dittohttp.NewClient(0)
	client.HTTPClient.Transport = dittohttp.MockSequenceRoundTripper(
		dittohttp.MockHTTPResult{StatusCode: http.StatusTooManyRequests, Body: "retry"},
		dittohttp.MockHTTPResult{StatusCode: http.StatusOK, Body: "done"},
	)
	_, _ = client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	want := "done"
	second, _ := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)
	got := second.Body

	assert.Equal(t, want, got)
}

func TestMockSequenceRoundTripper_successWhenResultHasHeaders(t *testing.T) {
	client := dittohttp.NewClient(0)
	client.HTTPClient.Transport = dittohttp.MockSequenceRoundTripper(
		dittohttp.MockHTTPResult{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"1"}}},
	)

	want := "1"
	first, _ := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)
	got := first.Header.Get("Retry-After")

	assert.Equal(t, want, got)
}

func TestMockSequenceRoundTripper_failsWhenSequenceIsExhausted(t *testing.T) {
	client := dittohttp.NewClient(0)
	client.HTTPClient.Transport = dittohttp.MockSequenceRoundTripper()

	want := dittohttp.FailedToExecuteRequest
	_, got := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	assert.ErrorIs(t, got, want)
}
