package http

import (
	"context"
	"net/http"

	"github.com/lhbelfanti/ditto/http/response"
)

// MigrationRunner is a function that executes pending database migrations.
type MigrationRunner func(ctx context.Context) error

// RegisterSystemRoutes attaches standard system endpoints to the provided mux:
//   - GET /ping/v1
//   - POST /migrations/run/v1 (only if runner is non-nil)
func RegisterSystemRoutes(mux *http.ServeMux, runner MigrationRunner) {
	mux.HandleFunc("GET /ping/v1", pingHandlerV1())
	if runner != nil {
		mux.HandleFunc("POST /migrations/run/v1", migrationsRunHandlerV1(runner))
	}
}

func pingHandlerV1() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.Send(r.Context(), w, http.StatusOK, "pong", nil, nil)
	}
}

func migrationsRunHandlerV1(run MigrationRunner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := run(r.Context()); err != nil {
			response.Send(r.Context(), w, http.StatusInternalServerError, "Failed to run migrations", nil, err)
			return
		}
		response.Send(r.Context(), w, http.StatusOK, "Migrations applied successfully", nil, nil)
	}
}
