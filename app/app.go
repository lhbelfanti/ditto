// Package app bootstraps a service end to end: listen port, database, migrations at boot, system
// routes and graceful shutdown. A service's main reduces to a single call to Run.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/env"
	dittohttp "github.com/lhbelfanti/ditto/v2/http"
	"github.com/lhbelfanti/ditto/v2/log"
	"github.com/lhbelfanti/ditto/v2/migration"
)

const (
	startupCheckTimeout = 5 * time.Second
	pingCheckTimeout    = 2 * time.Second
	shutdownTimeout     = 5 * time.Second
)

// Options describes one service's bootstrap.
type Options struct {
	// Name appears in the startup log line, e.g. "nebula-earth".
	Name string

	// PortEnv names the environment variable holding the internal listen port, e.g. APP_INTERNAL_PORT.
	PortEnv string

	// MigrationsDir enables the database. When set, Run requires the POSTGRES_DB_* variables,
	// verifies connectivity, applies pending migrations, and mounts GET /database/ping/v1. Leave it
	// empty for a service without a database.
	MigrationsDir string

	// Routes registers the service's own endpoints on mux. pg is nil when the database is disabled.
	Routes func(mux *http.ServeMux, pg *database.Postgres)

	// Wrap optionally decorates the whole handler, e.g. with request ID or CORS middleware.
	Wrap func(http.Handler) http.Handler
}

// Run configures logging, validates the environment, opens the database and applies migrations
// when the service declares one, then serves HTTP until SIGINT or SIGTERM and shuts down
// gracefully. It returns an error for any failure before the server starts, and nil after a clean
// shutdown.
func Run(opts Options) error {
	ctx := context.Background()
	log.NewCustomLogger(os.Stdout, zerolog.InfoLevel)

	port, err := env.RequirePort(os.LookupEnv, opts.PortEnv)
	if err != nil {
		return err
	}

	var pg *database.Postgres
	if opts.MigrationsDir != "" {
		err = database.RequireEnv(os.LookupEnv)
		if err != nil {
			return err
		}

		pg, err = database.InitPostgres()
		if err != nil {
			return err
		}
		defer pg.Close()

		err = pg.MakeCheck(startupCheckTimeout)(ctx)
		if err != nil {
			return err
		}

		err = migration.MakeRunner(pg.Database(), opts.MigrationsDir)(ctx)
		if err != nil {
			return err
		}
	}

	mux := http.NewServeMux()
	systemRoutes := dittohttp.RegisterSystemRoutes(mux)
	if pg != nil {
		systemRoutes.WithDatabasePing(dittohttp.DatabasePing(pg.MakeCheck(pingCheckTimeout)))
	}

	if opts.Routes != nil {
		opts.Routes(mux, pg)
	}

	var handler http.Handler = mux
	if opts.Wrap != nil {
		handler = opts.Wrap(mux)
	}

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{Addr: addr, Handler: handler}

	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info(ctx, "starting "+opts.Name+" on "+addr)
	return dittohttp.GracefulShutdown(signalCtx, server.ListenAndServe, server.Shutdown, shutdownTimeout)
}
