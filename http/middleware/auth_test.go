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

func mockSelectUserIDByToken(userID int, err error) middleware.SelectUserIDByToken {
	return func(ctx context.Context, token string) (int, error) {
		return userID, err
	}
}

func TestAuth_successInjectsUserIDIntoContext(t *testing.T) {
	var gotUserID int
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = middleware.UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Auth(mockSelectUserIDByToken(42, nil))(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer a-valid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 42, gotUserID)
}

func TestAuth_failsWhenAuthorizationHeaderMissing(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true })

	handler := middleware.Auth(mockSelectUserIDByToken(0, nil))(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, nextCalled)
	assert.Contains(t, rr.Body.String(), "authorization header required")
}

func TestAuth_failsWhenSchemeIsNotBearer(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true })

	handler := middleware.Auth(mockSelectUserIDByToken(0, nil))(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic a-valid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, nextCalled)
	assert.Contains(t, rr.Body.String(), middleware.ErrMsgInvalidToken)
}

func TestAuth_failsWhenAuthorizationHeaderIsMalformed(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true })

	handler := middleware.Auth(mockSelectUserIDByToken(0, nil))(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, nextCalled)
}

func TestAuth_failsWhenTokenLookupFails(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true })

	handler := middleware.Auth(mockSelectUserIDByToken(0, errors.New("session not found")))(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer an-expired-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, nextCalled)
	assert.Contains(t, rr.Body.String(), middleware.ErrMsgInvalidToken)
}

func TestUserIDFromContext_returnsZeroWhenNotSet(t *testing.T) {
	got := middleware.UserIDFromContext(context.Background())

	assert.Equal(t, 0, got)
}

func TestContextWithUserID_roundTrips(t *testing.T) {
	ctx := middleware.ContextWithUserID(context.Background(), 7)

	got := middleware.UserIDFromContext(ctx)

	assert.Equal(t, 7, got)
}
