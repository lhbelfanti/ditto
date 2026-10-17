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
	mockNext := middleware.MockOKHandler()
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(42, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer a-valid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := http.StatusOK
	got := rr.Code

	assert.Equal(t, want, got)
}

func TestAuth_successWhenTokenIsValidInjectsUserID(t *testing.T) {
	var gotUserID int
	mockNext := middleware.MockUserIDHandler(&gotUserID)
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(42, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer a-valid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := 42
	got := gotUserID

	assert.Equal(t, want, got)
}

func TestAuth_failsWhenAuthorizationHeaderMissing(t *testing.T) {
	mockNext := middleware.MockOKHandler()
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := http.StatusUnauthorized
	got := rr.Code

	assert.Equal(t, want, got)
}

func TestAuth_failsWhenAuthorizationHeaderMissingDoesNotCallNext(t *testing.T) {
	nextCalled := false
	mockNext := middleware.MockNextHandler(&nextCalled)
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.False(t, nextCalled)
}

func TestAuth_failsWhenAuthorizationHeaderMissingExplainsWhy(t *testing.T) {
	mockNext := middleware.MockOKHandler()
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Contains(t, rr.Body.String(), "authorization header required")
}

func TestAuth_failsWhenSchemeIsNotBearer(t *testing.T) {
	mockNext := middleware.MockOKHandler()
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic a-valid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := http.StatusUnauthorized
	got := rr.Code

	assert.Equal(t, want, got)
}

func TestAuth_failsWhenSchemeIsNotBearerDoesNotCallNext(t *testing.T) {
	nextCalled := false
	mockNext := middleware.MockNextHandler(&nextCalled)
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic a-valid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.False(t, nextCalled)
}

func TestAuth_failsWhenSchemeIsNotBearerExplainsWhy(t *testing.T) {
	mockNext := middleware.MockOKHandler()
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic a-valid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Contains(t, rr.Body.String(), middleware.ErrMsgInvalidToken)
}

func TestAuth_failsWhenAuthorizationHeaderIsMalformed(t *testing.T) {
	mockNext := middleware.MockOKHandler()
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := http.StatusUnauthorized
	got := rr.Code

	assert.Equal(t, want, got)
}

func TestAuth_failsWhenAuthorizationHeaderIsMalformedDoesNotCallNext(t *testing.T) {
	nextCalled := false
	mockNext := middleware.MockNextHandler(&nextCalled)
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, nil)

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.False(t, nextCalled)
}

func TestAuth_failsWhenTokenLookupFails(t *testing.T) {
	mockNext := middleware.MockOKHandler()
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, errors.New("session not found"))

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer an-expired-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	want := http.StatusUnauthorized
	got := rr.Code

	assert.Equal(t, want, got)
}

func TestAuth_failsWhenTokenLookupFailsDoesNotCallNext(t *testing.T) {
	nextCalled := false
	mockNext := middleware.MockNextHandler(&nextCalled)
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, errors.New("session not found"))

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer an-expired-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.False(t, nextCalled)
}

func TestAuth_failsWhenTokenLookupFailsExplainsWhy(t *testing.T) {
	mockNext := middleware.MockOKHandler()
	mockSelectUserIDByToken := middleware.MockSelectUserIDByToken(0, errors.New("session not found"))

	auth := middleware.Auth(mockSelectUserIDByToken)
	handler := auth(mockNext)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer an-expired-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Contains(t, rr.Body.String(), middleware.ErrMsgInvalidToken)
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
