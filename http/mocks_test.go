package http_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	dittohttp "github.com/lhbelfanti/ditto/v2/http"
)

func TestMockSequenceRoundTripper_successWhenResponsesAreSequential(t *testing.T) {
	client := dittohttp.NewClient(0)
	client.HTTPClient.Transport = dittohttp.MockSequenceRoundTripper(
		dittohttp.MockHTTPResult{StatusCode: http.StatusTooManyRequests, Body: "retry", Header: http.Header{"Retry-After": []string{"1"}}},
		dittohttp.MockHTTPResult{StatusCode: http.StatusOK, Body: "done"},
	)

	first, firstErr := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)
	second, secondErr := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	assert.NoError(t, firstErr)
	assert.Equal(t, http.StatusTooManyRequests, first.StatusCode)
	assert.Equal(t, "1", first.Header.Get("Retry-After"))
	assert.NoError(t, secondErr)
	assert.Equal(t, http.StatusOK, second.StatusCode)
	assert.Equal(t, "done", second.Body)
}

func TestMockSequenceRoundTripper_failsWhenSequenceIsExhausted(t *testing.T) {
	client := dittohttp.NewClient(0)
	client.HTTPClient.Transport = dittohttp.MockSequenceRoundTripper()

	_, err := client.NewRequest(context.Background(), http.MethodGet, "http://example.test", nil)

	assert.ErrorIs(t, err, dittohttp.FailedToExecuteRequest)
}
