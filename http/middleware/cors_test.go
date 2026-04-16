package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/http/middleware"
)

func TestCORS_successWithHeadersPresent(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.CORS()(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.NotEmpty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_successWithOptionsPreflightReturning204(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.CORS()(next)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := http.StatusNoContent
	got := rr.Code

	assert.Equal(t, want, got)
	assert.False(t, nextCalled)
}

func TestCORS_successWithAllowCredentials(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.CORS()(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := "true"
	got := rr.Header().Get("Access-Control-Allow-Credentials")

	assert.Equal(t, want, got)
}

func TestCORS_successWithOriginFromEnv(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://example.com")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.CORS()(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := "http://example.com"
	got := rr.Header().Get("Access-Control-Allow-Origin")

	assert.Equal(t, want, got)
}
