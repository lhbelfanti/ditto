package http

import (
	"net/http"

	"github.com/lhbelfanti/ditto/v2/http/response"
)

// RegisterSystemRoutes mounts GET /ping/v1 (unconditional liveness — never touches the database)
// and returns a SystemRoutes to opt into the routes a given service actually needs:
//
//	dittohttp.RegisterSystemRoutes(mux).WithMigrationRunner(runner).WithDatabasePing(dbPing)
func RegisterSystemRoutes(mux *http.ServeMux) *SystemRoutes {
	mux.HandleFunc("GET /ping/v1", pingHandlerV1())
	return &SystemRoutes{mux: mux}
}

// MountSystemRoutes mounts GET /ping/v1, and GET /database/ping/v1 when ping is not nil.
func MountSystemRoutes(mux *http.ServeMux, ping DatabasePing) {
	routes := RegisterSystemRoutes(mux)
	if ping != nil {
		routes.WithDatabasePing(ping)
	}
}

// WithMigrationRunner mounts POST /migrations/run/v1.
//
// Deprecated: exposes DDL over HTTP without authentication. app.Run applies migrations at boot.
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
			// response.Send already logs the message and err below (code >= 400) — do not log again here.
			response.Send(ctx, w, http.StatusInternalServerError, ErrMsgMigrationsFailed, nil, ErrMigrationsFailed)
			return
		}
		response.Send(ctx, w, http.StatusOK, "Migrations applied successfully", nil, nil)
	}
}
