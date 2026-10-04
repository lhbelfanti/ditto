package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/env"
)

func TestRun_failsWhenPortIsMissing(t *testing.T) {
	err := Run(Options{Name: "svc", PortEnv: "APP_TEST_MISSING_PORT"})

	var keyErr *env.KeyError
	require.ErrorAs(t, err, &keyErr)
}

func TestRun_failsWhenDatabaseEnvIsIncomplete(t *testing.T) {
	t.Setenv("APP_TEST_PORT", "4000")
	t.Setenv("POSTGRES_DB_PORT", "")

	err := Run(Options{Name: "svc", PortEnv: "APP_TEST_PORT", MigrationsDir: "./migrations"})

	assert.Error(t, err)
}

func TestNewHandler_successWhenDatabaseIsDisabled(t *testing.T) {
	h := newHandler(Options{}, nil)

	assert.Equal(t, http.StatusOK, serve(t, h, "GET", "/ping/v1").Code)
	assert.Equal(t, http.StatusNotFound, serve(t, h, "GET", "/database/ping/v1").Code)
	assert.Equal(t, http.StatusNotFound, serve(t, h, "POST", "/migrations/run/v1").Code)
}

func TestNewHandler_successWhenRoutesAndWrapAreSet(t *testing.T) {
	opts := Options{
		Routes: func(mux *http.ServeMux, _ *database.Postgres) {
			mux.HandleFunc("GET /items/v1", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})
		},
		Wrap: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Wrapped", "true")
				next.ServeHTTP(w, r)
			})
		},
	}
	h := newHandler(opts, nil)

	rec := serve(t, h, "GET", "/items/v1")

	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, "true", rec.Header().Get("X-Wrapped"))
}

func serve(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}
