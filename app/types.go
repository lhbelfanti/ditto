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

		// PortEnv names the environment variable holding the internal listen port. Defaults to APP_INTERNAL_PORT.
		PortEnv string

		// Mux holds the service's own endpoints. Run mounts its system routes on it. Run creates an
		// empty one when nil.
		Mux *http.ServeMux

		// Database is the pool the service opened with InitDatabase. When set, Run mounts GET
		// /database/ping/v1 and leaves closing the pool to the caller.
		Database *database.Postgres

		// MigrationsDir makes Run open the database and apply migrations itself.
		//
		// Deprecated: open it with InitDatabase and set Database.
		MigrationsDir string

		// Timeouts bounds each phase of the bootstrap. Zero fields fall back to their defaults.
		Timeouts Timeouts

		// Routes registers the service's own endpoints. pg is nil when the database is disabled.
		//
		// Deprecated: build the endpoints on Mux.
		Routes Routes

		// Wrap optionally decorates the whole handler. It runs outside every WithMiddleware entry.
		//
		// Deprecated: use WithMiddleware.
		Wrap Wrap

		middleware []Middleware
	}

	// Timeouts bounds the startup database check, the database ping route, and graceful shutdown.
	Timeouts struct {
		Startup  time.Duration
		Ping     time.Duration
		Shutdown time.Duration
	}

	// Routes registers a service's endpoints on mux.
	Routes func(mux *http.ServeMux, pg *database.Postgres)

	// Middleware decorates the whole handler, e.g. with request ID or CORS.
	Middleware func(next http.Handler) http.Handler

	// Wrap decorates the whole handler before it is served.
	//
	// Deprecated: use Middleware.
	Wrap = Middleware
)

const defaultPortEnv string = "APP_INTERNAL_PORT"

// WithMiddleware returns a copy of o that also decorates the handler with m. The first middleware
// added is the outermost, so it sees the request first.
func (o Options) WithMiddleware(m Middleware) Options {
	o.middleware = append(append([]Middleware(nil), o.middleware...), m)
	return o
}

// orDefaults returns o with an empty PortEnv replaced by its default.
func (o Options) orDefaults() Options {
	if o.PortEnv == "" {
		o.PortEnv = defaultPortEnv
	}
	return o
}

// decorate wraps h with every WithMiddleware entry, the first one outermost, and then with Wrap.
func (o Options) decorate(h http.Handler) http.Handler {
	for i := len(o.middleware) - 1; i >= 0; i-- {
		h = o.middleware[i](h)
	}
	if o.Wrap != nil {
		h = o.Wrap(h)
	}
	return h
}

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
