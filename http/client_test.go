package http_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http"
)

func TestNewClient_success(t *testing.T) {
	want := 5 * time.Second
	client := http.NewClient(want)

	got := client.HTTPClient.Timeout

	assert.Equal(t, want, got)
}

func TestWithHeader_success(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.test", nil)

	http.WithHeader("Authorization", "Bearer token")(req)

	want := "Bearer token"
	got := req.Header.Get("Authorization")

	assert.Equal(t, want, got)
}

func TestWithHeader_successWhenHeaderAlreadyHasValue(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.test", nil)
	req.Header.Set("Authorization", "Bearer old")

	http.WithHeader("Authorization", "Bearer new")(req)

	want := []string{"Bearer new"}
	got := req.Header.Values("Authorization")

	assert.Equal(t, want, got)
}
