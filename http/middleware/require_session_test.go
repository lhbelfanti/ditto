package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/middleware"
)

func TestRequireSession_success(t *testing.T) {
	gotUserID := 0
	mockGetSession := middleware.MockGetSessionByToken(middleware.Session{UserID: 42}, nil)
	handler := middleware.RequireSession(mockGetSession)(middleware.CaptureUserIDHandler(&gotUserID))
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "bearer valid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := 42
	got := gotUserID

	assert.Equal(t, want, got)
}

func TestRequireSession_successWhenCalledTwice(t *testing.T) {
	mockGetSession := middleware.MockGetSessionByTokenExpecting("valid-token", middleware.Session{UserID: 42})
	handler := middleware.RequireSession(mockGetSession)(middleware.MockOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	first := httptest.NewRecorder()
	second := httptest.NewRecorder()

	handler.ServeHTTP(first, req)
	handler.ServeHTTP(second, req)

	want := []int{http.StatusOK, http.StatusOK}
	got := []int{first.Code, second.Code}

	assert.Equal(t, want, got)
}

func TestRequireSession_failsWhenAuthorizationHeaderMissing(t *testing.T) {
	gotCalls := 0
	mockGetSession := middleware.MockGetSessionByToken(middleware.Session{UserID: 42}, nil)
	handler := middleware.RequireSession(mockGetSession)(middleware.MockCountingHandler(&gotCalls))
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := `{"code":401,"message":"` + middleware.ErrMsgMissingAuthHeader + `"}` + "\n"
	wantCalls := 0
	got := w.Body.String()

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}

func TestRequireSession_failsWhenBearerTokenMalformed(t *testing.T) {
	for _, header := range []string{"Basic token", "Bearer", "Bearer ", "Bearer a token", "Bearer a\ttoken"} {
		gotCalls := 0
		mockGetSession := middleware.MockGetSessionByToken(middleware.Session{UserID: 42}, nil)
		handler := middleware.RequireSession(mockGetSession)(middleware.MockCountingHandler(&gotCalls))
		req := httptest.NewRequest(http.MethodGet, "/private", nil)
		req.Header.Set("Authorization", header)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		want := `{"code":401,"message":"` + middleware.ErrMsgInvalidSession + `"}` + "\n"
		wantCalls := 0
		got := w.Body.String()

		assert.Equal(t, want, got)
		assert.Equal(t, wantCalls, gotCalls)
	}
}

func TestRequireSession_failsWhenSessionLookupFails(t *testing.T) {
	gotCalls := 0
	mockGetSession := middleware.MockGetSessionByToken(middleware.Session{}, errors.New("unavailable"))
	handler := middleware.RequireSession(mockGetSession)(middleware.MockCountingHandler(&gotCalls))
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := `{"code":401,"message":"` + middleware.ErrMsgInvalidSession + `"}` + "\n"
	wantCalls := 0
	got := w.Body.String()

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}

func TestRequireSession_failsWhenSessionHasNoUser(t *testing.T) {
	for _, userID := range []int{0, -1} {
		gotCalls := 0
		mockGetSession := middleware.MockGetSessionByToken(middleware.Session{UserID: userID}, nil)
		handler := middleware.RequireSession(mockGetSession)(middleware.MockCountingHandler(&gotCalls))
		req := httptest.NewRequest(http.MethodGet, "/private", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		want := http.StatusUnauthorized
		wantCalls := 0
		got := w.Code

		assert.Equal(t, want, got)
		assert.Equal(t, wantCalls, gotCalls)
	}
}

func TestRequireSession_failsWhenReadOnlySessionWrites(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		gotCalls := 0
		mockGetSession := middleware.MockGetSessionByToken(middleware.Session{UserID: 42, ReadOnly: true}, nil)
		handler := middleware.RequireSession(mockGetSession)(middleware.MockCountingHandler(&gotCalls))
		req := httptest.NewRequest(method, "/private", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		want := `{"code":403,"message":"` + middleware.ErrMsgReadOnlySession + `"}` + "\n"
		wantCalls := 0
		got := w.Body.String()

		assert.Equal(t, want, got)
		assert.Equal(t, wantCalls, gotCalls)
	}
}

func TestRequireSession_successWhenReadOnlySessionReads(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		gotCalls := 0
		mockGetSession := middleware.MockGetSessionByToken(middleware.Session{UserID: 42, ReadOnly: true}, nil)
		handler := middleware.RequireSession(mockGetSession)(middleware.MockCountingHandler(&gotCalls))
		req := httptest.NewRequest(method, "/private", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		want := http.StatusOK
		wantCalls := 1
		got := w.Code

		assert.Equal(t, want, got)
		assert.Equal(t, wantCalls, gotCalls)
	}
}

func TestRequireSession_successWhenSessionCanWrite(t *testing.T) {
	gotCalls := 0
	mockGetSession := middleware.MockGetSessionByToken(middleware.Session{UserID: 42}, nil)
	handler := middleware.RequireSession(mockGetSession)(middleware.MockCountingHandler(&gotCalls))
	req := httptest.NewRequest(http.MethodPost, "/private", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	want := http.StatusOK
	wantCalls := 1
	got := w.Code

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}
