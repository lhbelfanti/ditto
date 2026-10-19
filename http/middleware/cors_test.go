package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/middleware"
)

func TestCORS_successWhenOriginIsAllowed(t *testing.T) {
	mockNext := middleware.MockOKHandler()

	cors := middleware.CORS()
	handler := cors(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := middleware.MockCORSHeaders("http://localhost:3000")
	got := w.Header()

	assert.Equal(t, want, got)
}

func TestCORS_successWhenOriginIsNotAllowed(t *testing.T) {
	mockNext := middleware.MockOKHandler()

	cors := middleware.CORS()
	handler := cors(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://not-allowed.example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := middleware.MockCORSHeaders("")
	got := w.Header()

	assert.Equal(t, want, got)
}

func TestCORS_successWhenOriginIsMissing(t *testing.T) {
	mockNext := middleware.MockOKHandler()

	cors := middleware.CORS()
	handler := cors(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := middleware.MockCORSHeaders("")
	got := w.Header()

	assert.Equal(t, want, got)
}

func TestCORS_successWhenRequestIsPreflight(t *testing.T) {
	gotCalls := 0
	mockNext := middleware.MockCountingHandler(&gotCalls)

	cors := middleware.CORS()
	handler := cors(mockNext)
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := http.StatusNoContent
	wantCalls := 0
	got := w.Result().StatusCode

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}

func TestCORS_successWhenOriginIsConfigured(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://example.com")
	mockNext := middleware.MockOKHandler()

	cors := middleware.CORS()
	handler := cors(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := middleware.MockCORSHeaders("http://example.com")
	got := w.Header()

	assert.Equal(t, want, got)
}

func TestCORS_successWhenMultipleOriginsConfigured(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://a.example.com, http://b.example.com")
	mockNext := middleware.MockOKHandler()

	cors := middleware.CORS()
	handler := cors(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://b.example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := middleware.MockCORSHeaders("http://b.example.com")
	got := w.Header()

	assert.Equal(t, want, got)
}
