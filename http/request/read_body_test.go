package request_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/request"
)

func TestReadBody_success(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader("a;b\n1;2\n"))
	w := httptest.NewRecorder()

	want := []byte("a;b\n1;2\n")
	got, _ := request.ReadBody(w, req, 16)

	assert.Equal(t, want, got)
}

func TestReadBody_successWhenTheBodyIsExactlyTheLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader("12345"))
	w := httptest.NewRecorder()

	want := []byte("12345")
	got, _ := request.ReadBody(w, req, 5)

	assert.Equal(t, want, got)
}

func TestReadBody_failsWhenTheBodyIsLargerThanTheLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader("123456"))
	w := httptest.NewRecorder()

	want := request.ErrBodyTooLarge
	_, got := request.ReadBody(w, req, 5)

	assert.ErrorIs(t, got, want)
}

func TestReadBody_failsWhenTheBodyCantBeRead(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/files", request.MockFailingBody())
	w := httptest.NewRecorder()

	want := request.ErrCantReadBody
	_, got := request.ReadBody(w, req, 5)

	assert.ErrorIs(t, got, want)
}
