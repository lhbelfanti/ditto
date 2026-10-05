package app

import (
	"net/http"
	"time"

	"github.com/lhbelfanti/ditto/v2/database"
)

type (
	// Options describes one service's bootstrap.
	Options struct {
		// Name appears in the startup log line, e.g. "nebula-earth".
		Name string

		// PortEnv names the environment variable holding the internal listen port, e.g. APP_INTERNAL_PORT.
		PortEnv string

		// MigrationsDir enables the database. When set, Run requires the POSTGRES_DB_* variables,
		// verifies connectivity, applies pending migrations, and mounts GET /database/ping/v1. Leave it
		// empty for a service without a database.
		MigrationsDir string

		// Timeouts bounds each phase of the bootstrap. Zero fields fall back to their defaults.
		Timeouts Timeouts

		// Routes registers the service's own endpoints. pg is nil when the database is disabled.
		Routes Routes

		// Wrap optionally decorates the whole handler, e.g. with request ID or CORS middleware.
		Wrap Wrap
	}

	// Timeouts bounds the startup database check, the database ping route, and graceful shutdown.
	Timeouts struct {
		Startup  time.Duration
		Ping     time.Duration
		Shutdown time.Duration
	}

	// Routes registers a service's endpoints on mux.
	Routes func(mux *http.ServeMux, pg *database.Postgres)

	// Wrap decorates the whole handler before it is served.
	Wrap func(next http.Handler) http.Handler
)

// orDefaults returns t with every zero field replaced by its default.
func (t Timeouts) orDefaults() Timeouts {
	if t.Startup == 0 {
		t.Startup = defaultStartupTimeout
	}
	if t.Ping == 0 {
		t.Ping = defaultPingTimeout
	}
	if t.Shutdown == 0 {
		t.Shutdown = defaultShutdownTimeout
	}
	return t
}
