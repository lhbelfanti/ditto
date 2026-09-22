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

// SystemRoutes mounts ditto's standard system endpoints on a mux, one opt-in route at a time.
type SystemRoutes struct {
	mux *http.ServeMux
}

// RegisterSystemRoutes mounts GET /ping/v1 (unconditional liveness — never touches the database)
// and returns a SystemRoutes to opt into the routes a given service actually needs:
//
//	dittohttp.RegisterSystemRoutes(mux).WithMigrationRunner(runner).WithDatabasePing(dbPing)
func RegisterSystemRoutes(mux *http.ServeMux) *SystemRoutes {
	mux.HandleFunc("GET /ping/v1", pingHandlerV1())
	return &SystemRoutes{mux: mux}
}

// WithMigrationRunner mounts POST /migrations/run/v1.
func (s *SystemRoutes) WithMigrationRunner(runner MigrationRunner) *SystemRoutes {
	s.mux.HandleFunc("POST /migrations/run/v1", migrationsRunHandlerV1(runner))
	return s
}

// WithDatabasePing mounts GET /database/ping/v1.
func (s *SystemRoutes) WithDatabasePing(dbPing DatabasePing) *SystemRoutes {
	s.mux.HandleFunc("GET /database/ping/v1", databasePingHandlerV1(dbPing))
	return s
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
