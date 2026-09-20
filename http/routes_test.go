package http_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	dittohttp "github.com/lhbelfanti/ditto/v2/http"
)

func TestRegisterSystemRoutes_ping(t *testing.T) {
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestRegisterSystemRoutes_migrationsSkippedWhenRunnerNil(t *testing.T) {
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusNotFound
	got := w.Code

	assert.Equal(t, want, got)
}

func TestRegisterSystemRoutes_migrationsSuccess(t *testing.T) {
	mockRunner := dittohttp.MockMigrationRunner(nil)
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, mockRunner, nil)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestRegisterSystemRoutes_migrationsFailure(t *testing.T) {
	mockRunner := dittohttp.MockMigrationRunner(errors.New("migration failed"))
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, mockRunner, nil)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusInternalServerError
	got := w.Code

	assert.Equal(t, want, got)
}

func TestRegisterSystemRoutes_migrationsFailure_bodyExcludesUnderlyingError(t *testing.T) {
	underlyingErrText := "pq: relation \"trades\" already exists"
	mockRunner := dittohttp.MockMigrationRunner(errors.New(underlyingErrText))
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, mockRunner, nil)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	got := w.Body.String()

	assert.NotContains(t, got, underlyingErrText)
}

func TestRegisterSystemRoutes_databasePingSkippedWhenNil(t *testing.T) {
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusNotFound
	got := w.Code

	assert.Equal(t, want, got)
}

func TestRegisterSystemRoutes_databasePingSuccess(t *testing.T) {
	mockPing := dittohttp.MockDatabasePing(nil)
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, nil, mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestRegisterSystemRoutes_databasePingFailure(t *testing.T) {
	mockPing := dittohttp.MockDatabasePing(errors.New("database unreachable"))
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, nil, mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusServiceUnavailable
	got := w.Code

	assert.Equal(t, want, got)
}

func TestRegisterSystemRoutes_databasePingFailure_bodyExcludesUnderlyingError(t *testing.T) {
	underlyingErrText := "pgx: connection refused on host db-internal:5432"
	mockPing := dittohttp.MockDatabasePing(errors.New(underlyingErrText))
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux, nil, mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	got := w.Body.String()

	assert.NotContains(t, got, underlyingErrText)
}
