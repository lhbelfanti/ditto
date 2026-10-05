package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/middleware"
)

func TestCORS_successWhenOriginIsAllowed(t *testing.T) {
	handler := middleware.CORS()(middleware.MockOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := "http://localhost:3000"
	got := rr.Header().Get("Access-Control-Allow-Origin")

	assert.Equal(t, want, got)
}

func TestCORS_successWhenOriginIsAllowedVariesByOrigin(t *testing.T) {
	handler := middleware.CORS()(middleware.MockOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Contains(t, rr.Header().Values("Vary"), "Origin")
}

func TestCORS_successWhenOriginIsNotAllowed(t *testing.T) {
	handler := middleware.CORS()(middleware.MockOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://not-allowed.example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Access-Control-Allow-Origin")

	assert.Empty(t, got)
}

func TestCORS_successWhenOriginIsMissing(t *testing.T) {
	handler := middleware.CORS()(middleware.MockOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Access-Control-Allow-Origin")

	assert.Empty(t, got)
}

func TestCORS_successWhenRequestIsPreflight(t *testing.T) {
	handler := middleware.CORS()(middleware.MockOKHandler())
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := http.StatusNoContent
	got := rr.Code

	assert.Equal(t, want, got)
}

func TestCORS_successWhenRequestIsPreflightDoesNotCallNext(t *testing.T) {
	nextCalled := false
	handler := middleware.CORS()(middleware.MockNextHandler(&nextCalled))
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.False(t, nextCalled)
}

func TestCORS_successWhenCredentialsAllowed(t *testing.T) {
	handler := middleware.CORS()(middleware.MockOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := "true"
	got := rr.Header().Get("Access-Control-Allow-Credentials")

	assert.Equal(t, want, got)
}

func TestCORS_successWhenOriginIsConfigured(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://example.com")
	handler := middleware.CORS()(middleware.MockOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := "http://example.com"
	got := rr.Header().Get("Access-Control-Allow-Origin")

	assert.Equal(t, want, got)
}

func TestCORS_successWhenMultipleOriginsConfigured(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://a.example.com, http://b.example.com")
	handler := middleware.CORS()(middleware.MockOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://b.example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := "http://b.example.com"
	got := rr.Header().Get("Access-Control-Allow-Origin")

	assert.Equal(t, want, got)
}
