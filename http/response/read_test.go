package response_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/response"
)

func TestRead_success(t *testing.T) {
	mockData := map[string]string{"username": "ana"}
	w := httptest.NewRecorder()
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc"})
	response.Send(context.Background(), w, http.StatusOK, "username set", mockData, nil)

	want := response.Recorded[map[string]string]{
		Status: http.StatusOK,
		Header: http.Header{
			"Set-Cookie": {"session=abc"},
		},
		Message: "username set",
		Data: map[string]string{
			"username": "ana",
		},
	}
	got := response.Read[map[string]string](w)

	assert.Equal(t, want, got)
}

func TestRead_successWhenTheResponseHasNoData(t *testing.T) {
	w := httptest.NewRecorder()
	response.Send(context.Background(), w, http.StatusUnauthorized, "unauthenticated", nil, nil)

	want := response.Recorded[any]{Status: http.StatusUnauthorized, Message: "unauthenticated"}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestRead_successWhenTheBodyIsNotJSON(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/auth/google/start/v1", nil)
	w := httptest.NewRecorder()
	http.Redirect(w, request, "https://accounts.example.com/authorize", http.StatusFound)

	want := response.Recorded[any]{
		Status: http.StatusFound,
		Header: http.Header{
			"Content-Type": {"text/html; charset=utf-8"},
			"Location":     {"https://accounts.example.com/authorize"},
		},
		Body: "<a href=\"https://accounts.example.com/authorize\">Found</a>.\n\n",
	}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestRead_successWhenTheJSONBodyIsInvalid(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.WriteString("{")

	want := response.Recorded[any]{Status: http.StatusInternalServerError, Header: http.Header{"Content-Type": {"application/json"}}, Body: "{"}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestRead_successWhenTheCodeIsNotTheStatus(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.WriteString(`{"code":500,"message":"pong"}`)

	want := response.Recorded[any]{
		Status: http.StatusOK,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Body: `{"code":500,"message":"pong"}`,
	}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}
