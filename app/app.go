// Package app bootstraps a service: InitDatabase opens the pool and applies migrations, and Run
// serves the service's mux with the system routes, middleware and graceful shutdown.
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
	defaultStartupTimeout  time.Duration = 5 * time.Second
	defaultPingTimeout     time.Duration = 2 * time.Second
	defaultShutdownTimeout time.Duration = 5 * time.Second
)

// InitDatabase validates the POSTGRES_DB_* variables, opens the pool, verifies connectivity and
// applies the pending migrations in ./migrations, or in the first of migrationsDir when given. The
// returned pool's Database method is the connection to inject into the service's makers; the caller
// closes the pool.
func InitDatabase(ctx context.Context, migrationsDir ...string) (*database.Postgres, error) {
	dir := defaultMigrationsDir
	if len(migrationsDir) > 0 {
		dir = migrationsDir[0]
	}

	err := database.RequireEnv(os.LookupEnv)
	if err != nil {
		return nil, err
	}

	pg, err := database.InitPostgres()
	if err != nil {
		return nil, err
	}

	err = pg.MakeCheck(defaultStartupTimeout)(ctx)
	if err != nil {
		pg.Close()
		return nil, err
	}

	err = migration.MakeRunner(pg.Database(), dir)(ctx)
	if err != nil {
		pg.Close()
		return nil, err
	}

	return pg, nil
}

// Run opens the database unless opts.Database is set or opts.NoDatabase says the service has none,
// then serves opts.Mux with the system routes mounted and the middleware applied, until SIGINT or
// SIGTERM, then shuts down gracefully. It returns an error for any failure before the server starts,
// and nil after a clean shutdown.
func Run(opts Options) error {
	ctx := context.Background()
	log.NewCustomLogger(os.Stdout, zerolog.InfoLevel)
	opts = opts.orDefaults()
	timeouts := opts.Timeouts.orDefaults()

	port, err := env.RequirePort(os.LookupEnv, opts.PortEnv)
	if err != nil {
		return err
	}

	pg := opts.Database
	if pg == nil && !opts.NoDatabase {
		pg, err = InitDatabase(ctx, opts.MigrationsDir)
		if err != nil {
			return err
		}
		defer pg.Close()
	}

	var ping dittohttp.DatabasePing
	if pg != nil {
		ping = dittohttp.DatabasePing(pg.MakeCheck(timeouts.Ping))
	}

	mux := opts.Mux
	if mux == nil {
		mux = http.NewServeMux()
	}
	dittohttp.MountSystemRoutes(mux, ping)
	if opts.Routes != nil {
		opts.Routes(mux, pg)
	}

	addr := fmt.Sprintf(":%d", port)
	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info(ctx, "starting "+opts.Name+" on "+addr)
	return dittohttp.Listen(signalCtx, addr, opts.decorate(mux), timeouts.Shutdown)
}
