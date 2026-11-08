package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
	"github.com/lhbelfanti/ditto/v2/http/response"
)

func TestRegisterSystemRoutes_success(t *testing.T) {
	mux := http.NewServeMux()
	_http.RegisterSystemRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := response.Recorded[any]{Status: http.StatusOK, Message: "pong"}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestRegisterSystemRoutes_successWhenDatabasePingIsNotOptedIn(t *testing.T) {
	mux := http.NewServeMux()
	_http.RegisterSystemRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := response.Recorded[any]{
		Status: http.StatusNotFound,
		Header: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
		Body: "404 page not found\n",
	}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestMountSystemRoutes_successWhenPingIsNil(t *testing.T) {
	mux := http.NewServeMux()
	_http.MountSystemRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	want := response.Recorded[any]{
		Status: http.StatusNotFound,
		Header: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
		Body: "404 page not found\n",
	}
	got := response.Read[any](rec)

	assert.Equal(t, want, got)
}

func TestMountSystemRoutes_successWhenPingIsSet(t *testing.T) {
	mux := http.NewServeMux()
	_http.MountSystemRoutes(mux, _http.MockDatabasePing(nil))

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	want := response.Recorded[any]{Status: http.StatusOK, Message: "pong"}
	got := response.Read[any](rec)

	assert.Equal(t, want, got)
}
