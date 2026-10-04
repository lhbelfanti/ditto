package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	dittohttp "github.com/lhbelfanti/ditto/v2/http"
)

func TestRegisterSystemRoutes_ping(t *testing.T) {
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestRegisterSystemRoutes_migrationsSkippedWhenNotOptedIn(t *testing.T) {
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusNotFound
	got := w.Code

	assert.Equal(t, want, got)
}

func TestRegisterSystemRoutes_databasePingSkippedWhenNotOptedIn(t *testing.T) {
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusNotFound
	got := w.Code

	assert.Equal(t, want, got)
}

func TestMountSystemRoutes_successWhenPingIsNil(t *testing.T) {
	mux := http.NewServeMux()
	dittohttp.MountSystemRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	want := http.StatusNotFound
	got := rec.Code

	assert.Equal(t, want, got)
}

func TestMountSystemRoutes_successWhenPingIsSet(t *testing.T) {
	mux := http.NewServeMux()
	dittohttp.MountSystemRoutes(mux, dittohttp.DatabasePing(func(context.Context) error { return nil }))

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	want := http.StatusOK
	got := rec.Code

	assert.Equal(t, want, got)
}
