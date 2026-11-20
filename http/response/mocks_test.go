package response_test

import (
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
