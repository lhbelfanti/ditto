package app

import (
	"net/http"
	"time"

	"github.com/lhbelfanti/ditto/v2/database"
)

type (
	// Options describes one service's bootstrap. The README's Options table explains each field.
	Options struct {
		Name       string
		PortEnv    string
		Mux        *http.ServeMux
		Database   *database.Postgres
		NoDatabase bool
		Timeouts   Timeouts
		middleware []Middleware
	}

	// Timeouts bounds the startup database check, the database ping route, and graceful shutdown.
	Timeouts struct {
		Startup  time.Duration
		Ping     time.Duration
		Shutdown time.Duration
	}

	// Middleware decorates the whole handler, e.g. with request ID or CORS.
	Middleware func(next http.Handler) http.Handler
)

const (
	defaultPortEnv         string        = "APP_INTERNAL_PORT"
	defaultStartupTimeout  time.Duration = 5 * time.Second
	defaultPingTimeout     time.Duration = 2 * time.Second
	defaultShutdownTimeout time.Duration = 5 * time.Second
)

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

// decorate wraps h with every WithMiddleware entry, the first one outermost.
func (o Options) decorate(h http.Handler) http.Handler {
	for i := len(o.middleware) - 1; i >= 0; i-- {
		h = o.middleware[i](h)
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
