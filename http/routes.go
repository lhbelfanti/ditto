package http

import (
	"context"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/http/response"
	"github.com/lhbelfanti/ditto/v2/log"
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
		mux.HandleFunc("GET /database/ping/v1", DatabasePingHandler(dbPing))
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

// DatabasePingHandler builds a handler that reports 200 on a successful ping and 503 on failure —
// exported so a caller can mount it under a project-specific route (e.g. a "/ready/v1" readiness
// contract) in addition to, or instead of, RegisterSystemRoutes' own "/database/ping/v1".
func DatabasePingHandler(ping DatabasePing) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := ping(ctx)
		if err != nil {
			// response.Send already logs the message and err below (code >= 400) — do not log again here.
			response.Send(ctx, w, http.StatusServiceUnavailable, ErrMsgDatabaseUnavailable, nil, ErrDatabaseUnavailable)
			return
		}
		response.Send(ctx, w, http.StatusOK, "pong", nil, nil)
	}
}

func migrationsRunHandlerV1(run MigrationRunner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := run(ctx)
		if err != nil {
			log.Err(ctx, err, ErrMsgMigrationsFailed)
			response.Send(ctx, w, http.StatusInternalServerError, ErrMsgMigrationsFailed, nil, ErrMigrationsFailed)
			return
		}
		response.Send(ctx, w, http.StatusOK, "Migrations applied successfully", nil, nil)
	}
}
