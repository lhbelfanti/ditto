package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	dittohttp "github.com/lhbelfanti/ditto/http"
)

func TestNewRequest_successWithJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, `{"symbol":"BTC/USDT"}`, readAll(t, r.Body))
		w.Header().Set("X-Test-Header", "present")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := dittohttp.NewClient(5 * time.Second)
	resp, err := client.NewRequest(context.Background(), http.MethodPost, server.URL, map[string]string{"symbol": "BTC/USDT"})

	assert.NoError(t, err)
	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, `{"ok":true}`, resp.Body)
	assert.Equal(t, "present", resp.Header.Get("X-Test-Header"))
}

func TestNewRequest_successWithNilBodyOmitsContentTypeAndSendsNoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "", r.Header.Get("Content-Type"))
		assert.Equal(t, "", readAll(t, r.Body))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := dittohttp.NewClient(5 * time.Second)
	resp, err := client.NewRequest(context.Background(), http.MethodGet, server.URL+"?symbol=BTCUSDT", nil)

	assert.NoError(t, err)
	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNewRequest_exposesResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-MBX-USED-WEIGHT-1M", "42")
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := dittohttp.NewClient(5 * time.Second)
	resp, err := client.NewRequest(context.Background(), http.MethodGet, server.URL, nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, "42", resp.Header.Get("X-MBX-USED-WEIGHT-1M"))
	assert.Equal(t, "5", resp.Header.Get("Retry-After"))
}

func TestNewRequest_failsWhenURLIsInvalid(t *testing.T) {
	client := dittohttp.NewClient(5 * time.Second)
	resp, err := client.NewRequest(context.Background(), http.MethodGet, "://invalid-url", nil)

	assert.ErrorIs(t, err, dittohttp.FailedToCreateRequest)
	assert.Equal(t, dittohttp.Response{}, resp)
}

func TestNewRequest_failsWhenContextIsCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := dittohttp.NewClient(5 * time.Second)
	resp, err := client.NewRequest(ctx, http.MethodGet, server.URL, nil)

	assert.ErrorIs(t, err, dittohttp.FailedToExecuteRequest)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, dittohttp.Response{}, resp)
}

func readAll(t *testing.T, r interface{ Read([]byte) (int, error) }) string {
	t.Helper()
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	return string(buf[:n])
}
