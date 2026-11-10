package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/middleware"
	"github.com/lhbelfanti/ditto/v2/http/response"
)

func TestRequireToken_success(t *testing.T) {
	handler := middleware.RequireToken("service-secret")(middleware.MockOKHandler())
	req := httptest.NewRequest(http.MethodDelete, "/internal/route/v1", nil)
	req.Header.Set("Authorization", "Bearer service-secret")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := response.Recorded[any]{Status: http.StatusOK}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestRequireToken_failsWhenAuthorizationHeaderMissing(t *testing.T) {
	gotCalls := 0
	handler := middleware.RequireToken("service-secret")(middleware.MockCountingHandler(&gotCalls))
	req := httptest.NewRequest(http.MethodDelete, "/internal/route/v1", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := response.Recorded[any]{Status: http.StatusUnauthorized, Message: middleware.ErrMsgMissingAuthHeader}
	wantCalls := 0
	got := response.Read[any](w)

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}

func TestRequireToken_failsWhenTokenIsAnother(t *testing.T) {
	for _, header := range []string{"Bearer another-secret", "Bearer service-secret-and-more", "Basic service-secret", "Bearer"} {
		gotCalls := 0
		handler := middleware.RequireToken("service-secret")(middleware.MockCountingHandler(&gotCalls))
		req := httptest.NewRequest(http.MethodDelete, "/internal/route/v1", nil)
		req.Header.Set("Authorization", header)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		want := response.Recorded[any]{Status: http.StatusUnauthorized, Message: middleware.ErrMsgInvalidToken}
		wantCalls := 0
		got := response.Read[any](w)

		assert.Equal(t, want, got, header)
		assert.Equal(t, wantCalls, gotCalls, header)
	}
}

func TestRequireToken_failsWhenNoTokenIsConfigured(t *testing.T) {
	gotCalls := 0
	handler := middleware.RequireToken("")(middleware.MockCountingHandler(&gotCalls))
	req := httptest.NewRequest(http.MethodDelete, "/internal/route/v1", nil)
	req.Header.Set("Authorization", "Bearer anything")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := response.Recorded[any]{Status: http.StatusUnauthorized, Message: middleware.ErrMsgInvalidToken}
	wantCalls := 0
	got := response.Read[any](w)

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}
