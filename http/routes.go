package http

import (
	"context"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/http/response"
)

// MigrationRunner is a function that executes pending database migrations.
type MigrationRunner func(ctx context.Context) error

// DatabasePing checks whether the database dependency is reachable.
type DatabasePing func(ctx context.Context) error

// RegisterSystemRoutes attaches standard system endpoints to the provided mux:
//   - GET /ping/v1 (unconditional liveness — never touches the database)
//   - GET /database/ping/v1 (only if dbPing is non-nil)
//   - POST /migrations/run/v1 (only if runner is non-nil)
func RegisterSystemRoutes(mux *http.ServeMux, runner MigrationRunner, dbPing DatabasePing) {
	mux.HandleFunc("GET /ping/v1", pingHandlerV1())
	if dbPing != nil {
		mux.HandleFunc("GET /database/ping/v1", databasePingHandlerV1(dbPing))
	}
	if runner != nil {
		mux.HandleFunc("POST /migrations/run/v1", migrationsRunHandlerV1(runner))
	}
}

func pingHandlerV1() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.Send(r.Context(), w, http.StatusOK, "pong", nil, nil)
	}
}

func databasePingHandlerV1(ping DatabasePing) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := ping(ctx)
		if err != nil {
			response.Send(ctx, w, http.StatusServiceUnavailable, "database unreachable", nil, err)
			return
		}
		response.Send(ctx, w, http.StatusOK, "pong", nil, nil)
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
