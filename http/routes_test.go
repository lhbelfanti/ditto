package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	dittohttp "github.com/lhbelfanti/ditto/http"
)

func TestRegisterSystemRoutes_ping(t *testing.T) {
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegisterSystemRoutes_migrationsSkippedWhenRunnerNil(t *testing.T) {
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRegisterSystemRoutes_migrationsSuccess(t *testing.T) {
	mux := http.NewServeMux()
	mockRunner := func(ctx context.Context) error { return nil }
	dittohttp.RegisterSystemRoutes(mux, mockRunner)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegisterSystemRoutes_migrationsFailure(t *testing.T) {
	mux := http.NewServeMux()
	mockRunner := func(ctx context.Context) error { return errors.New("migration failed") }
	dittohttp.RegisterSystemRoutes(mux, mockRunner)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
