package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/middleware"
)

func TestAuth_success(t *testing.T) {
	mockNext := middleware.MockUserIDResponseHandler()
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(42, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer a-valid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := "42"
	got := w.Body.String()

	assert.Equal(t, want, got)
}

func TestAuth_failsWhenAuthorizationHeaderMissing(t *testing.T) {
	gotCalls := 0
	mockNext := middleware.MockCountingHandler(&gotCalls)
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := `{"code":401,"message":"` + middleware.ErrMsgMissingAuthHeader + `"}` + "\n"
	wantCalls := 0
	got := w.Body.String()

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}

func TestAuth_failsWhenSchemeIsNotBearer(t *testing.T) {
	gotCalls := 0
	mockNext := middleware.MockCountingHandler(&gotCalls)
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic a-valid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := `{"code":401,"message":"` + middleware.ErrMsgInvalidToken + `"}` + "\n"
	wantCalls := 0
	got := w.Body.String()

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}

func TestAuth_failsWhenAuthorizationHeaderIsMalformed(t *testing.T) {
	gotCalls := 0
	mockNext := middleware.MockCountingHandler(&gotCalls)
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := `{"code":401,"message":"` + middleware.ErrMsgInvalidToken + `"}` + "\n"
	wantCalls := 0
	got := w.Body.String()

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}

func TestAuth_failsWhenBearerTokenIsEmpty(t *testing.T) {
	gotCalls := 0
	mockNext := middleware.MockCountingHandler(&gotCalls)
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(42, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := `{"code":401,"message":"` + middleware.ErrMsgInvalidToken + `"}` + "\n"
	wantCalls := 0
	got := w.Body.String()

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}

func TestAuth_failsWhenTokenLookupFails(t *testing.T) {
	gotCalls := 0
	mockNext := middleware.MockCountingHandler(&gotCalls)
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, errors.New("session not found"))

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer an-expired-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := `{"code":401,"message":"` + middleware.ErrMsgInvalidToken + `"}` + "\n"
	wantCalls := 0
	got := w.Body.String()

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}

func TestUserIDFromContext_successWhenUserIDAbsent(t *testing.T) {
	want := 0
	got := middleware.UserIDFromContext(context.Background())

	assert.Equal(t, want, got)
}

func TestContextWithUserID_success(t *testing.T) {
	ctx := middleware.ContextWithUserID(context.Background(), 7)

	want := 7
	got := middleware.UserIDFromContext(ctx)

	assert.Equal(t, want, got)
}
