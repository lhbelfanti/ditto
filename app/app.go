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
	defaultStartupTimeout  = 5 * time.Second
	defaultPingTimeout     = 2 * time.Second
	defaultShutdownTimeout = 5 * time.Second
)

// Run configures logging, validates the environment, opens the database and applies migrations
// when the service declares one, then serves HTTP until SIGINT or SIGTERM and shuts down
// gracefully. It returns an error for any failure before the server starts, and nil after a clean
// shutdown.
func Run(opts Options) error {
	ctx := context.Background()
	log.NewCustomLogger(os.Stdout, zerolog.InfoLevel)
	timeouts := opts.Timeouts.withDefaults()

	port, err := env.RequirePort(os.LookupEnv, opts.PortEnv)
	if err != nil {
		return err
	}

	var pg *database.Postgres
	var ping dittohttp.DatabasePing
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

		err = pg.MakeCheck(timeouts.Startup)(ctx)
		if err != nil {
			return err
		}

		err = migration.MakeRunner(pg.Database(), opts.MigrationsDir)(ctx)
		if err != nil {
			return err
		}

		ping = dittohttp.DatabasePing(pg.MakeCheck(timeouts.Ping))
	}

	mux := http.NewServeMux()
	dittohttp.MountSystemRoutes(mux, ping)
	if opts.Routes != nil {
		opts.Routes(mux, pg)
	}

	var handler http.Handler = mux
	if opts.Wrap != nil {
		handler = opts.Wrap(mux)
	}

	addr := fmt.Sprintf(":%d", port)
	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info(ctx, "starting "+opts.Name+" on "+addr)
	return dittohttp.Listen(signalCtx, addr, handler, timeouts.Shutdown)
}

func (t Timeouts) withDefaults() Timeouts {
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
