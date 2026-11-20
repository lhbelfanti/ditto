package response_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/response"
)

func TestMockRedirected_success(t *testing.T) {
	w := httptest.NewRecorder()
	http.SetCookie(w, &http.Cookie{Name: "state", Value: "", MaxAge: -1})
	http.Redirect(w, httptest.NewRequest(http.MethodGet, "/", nil), "http://localhost:4000/?error=denied&restorable=true", http.StatusFound)

	want := response.MockRedirected("http://localhost:4000/?error=denied&restorable=true", "state=; Max-Age=0")
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestMockRedirected_successWhenNoCookiesAreSet(t *testing.T) {
	w := httptest.NewRecorder()
	http.Redirect(w, httptest.NewRequest(http.MethodGet, "/", nil), "/home", http.StatusFound)

	want := response.MockRedirected("/home")
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestMockOK_success(t *testing.T) {
	mockData := map[string]int{"id": 1}
	w := httptest.NewRecorder()
	response.Send(context.Background(), w, http.StatusOK, "updated", mockData, nil)

	want := response.MockOK("updated", mockData)
	got := response.Read[map[string]int](w)

	assert.Equal(t, want, got)
}

func TestMockCreated_success(t *testing.T) {
	mockData := map[string]int{"id": 1}
	w := httptest.NewRecorder()
	response.Send(context.Background(), w, http.StatusCreated, "created", mockData, nil)

	want := response.MockCreated("created", mockData)
	got := response.Read[map[string]int](w)

	assert.Equal(t, want, got)
}

func TestMockBadRequest_success(t *testing.T) {
	w := httptest.NewRecorder()
	response.Send(context.Background(), w, http.StatusBadRequest, "invalid id", nil, errors.New("cause"))

	want := response.MockBadRequest("invalid id")
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestMockNotFound_success(t *testing.T) {
	w := httptest.NewRecorder()
	response.Send(context.Background(), w, http.StatusNotFound, "account not found", nil, errors.New("cause"))

	want := response.MockNotFound("account not found")
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestMockInternalServerError_success(t *testing.T) {
	w := httptest.NewRecorder()
	response.Send(context.Background(), w, http.StatusInternalServerError, "can't update account", nil, errors.New("cause"))

	want := response.MockInternalServerError("can't update account")
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}
