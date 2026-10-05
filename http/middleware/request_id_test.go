package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/middleware"
	"github.com/lhbelfanti/ditto/v2/log"
)

func TestRequestID_successWhenHeaderIsAbsent(t *testing.T) {
	handler := middleware.RequestID(middleware.MockNoopHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	got := w.Header().Get("X-Request-ID")

	assert.NotEmpty(t, got)
}

func TestRequestID_successWhenRequestsAreDistinct(t *testing.T) {
	handler := middleware.RequestID(middleware.MockNoopHandler())
	firstW := httptest.NewRecorder()
	handler.ServeHTTP(firstW, httptest.NewRequest(http.MethodGet, "/", nil))
	secondW := httptest.NewRecorder()
	handler.ServeHTTP(secondW, httptest.NewRequest(http.MethodGet, "/", nil))

	notWant := firstW.Header().Get("X-Request-ID")
	got := secondW.Header().Get("X-Request-ID")

	assert.NotEqual(t, notWant, got)
}

func TestRequestID_successWhenInboundHeaderPresent(t *testing.T) {
	handler := middleware.RequestID(middleware.MockNoopHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "upstream-id-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := "upstream-id-123"
	got := w.Header().Get("X-Request-ID")

	assert.Equal(t, want, got)
}

func TestRequestID_successWhenContextIsLogged(t *testing.T) {
	buf := log.MockLogOutput(t)
	handler := middleware.RequestID(middleware.MockRequestIDLoggingHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	got := buf.String()

	assert.Contains(t, got, "request_id")
}
