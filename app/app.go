// Package app bootstraps a service: InitDatabase opens the pool and applies migrations, and Run
// serves the service's mux with the system routes, middleware and graceful shutdown.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/env"
	_http "github.com/lhbelfanti/ditto/v2/http"
	"github.com/lhbelfanti/ditto/v2/log"
	"github.com/lhbelfanti/ditto/v2/migration"
)

// provision creates the service's role and database through the administrator credentials in
// POSTGRES_ADMIN_USER and POSTGRES_ADMIN_PASS. Without them it does nothing, so a database provisioned
// by other means keeps working.
func provision(ctx context.Context) error {
	admin, err := database.OpenAdmin(ctx)
	if err != nil {
		if errors.Is(err, database.ErrAdminNotConfigured) {
			return nil
		}

		return fmt.Errorf("%w: %w", ErrCantOpenAdmin, err)
	}

	defer admin.Close()

	check := admin.MakeCheck(defaultStartupTimeout)

	err = check(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAdminUnavailable, err)
	}

	selectStatement := database.MakeSelectOne[string](admin.Database(), pgx.RowTo[string])
	update := database.MakeUpdate(admin.Database())
	execFormatted := database.MakeExecFormatted(selectStatement, update)
	provisionTarget := database.MakeProvision(execFormatted, database.TargetFromEnv())

	err = provisionTarget(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCantProvision, err)
	}

	return nil
}

// InitDatabase opens the database and applies the pending migrations in ./migrations. See
// InitDatabaseWithMigrationFolder for another folder.
func InitDatabase(ctx context.Context) (*database.Postgres, error) {
	const migrationsDir string = "./migrations"

	return InitDatabaseWithMigrationFolder(ctx, migrationsDir)
}

// InitDatabaseWithMigrationFolder validates the POSTGRES_DB_* variables, creates the service's role
// and database when administrator credentials are set (see provision), opens the pool, verifies
// connectivity and applies the pending migrations in dir. The returned pool's Database method is
// the connection to inject into the service's makers; the caller closes the pool.
func InitDatabaseWithMigrationFolder(ctx context.Context, dir string) (*database.Postgres, error) {
	err := database.RequireEnv(os.LookupEnv)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidDatabaseEnv, err)
	}

	err = provision(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCantProvisionDatabase, err)
	}

	pg, err := database.InitPostgres(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCantOpenDatabase, err)
	}

	check := pg.MakeCheck(defaultStartupTimeout)

	err = check(ctx)
	if err != nil {
		pg.Close()
		return nil, fmt.Errorf("%w: %w", ErrCantReachDatabase, err)
	}

	runMigrations := migration.MakeRunner(pg.Database(), dir)

	err = runMigrations(ctx)
	if err != nil {
		pg.Close()
		return nil, fmt.Errorf("%w: %w", ErrCantApplyMigrations, err)
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
		return fmt.Errorf("%w: %w", ErrInvalidPort, err)
	}

	pg := opts.Database
	if pg == nil && !opts.NoDatabase {
		pg, err = InitDatabase(ctx)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCantInitDatabase, err)
		}

		defer pg.Close()
	}

	var ping _http.DatabasePing
	if pg != nil {
		ping = _http.DatabasePing(pg.MakeCheck(timeouts.Ping))
	}

	mux := opts.Mux
	if mux == nil {
		mux = http.NewServeMux()
	}

	_http.MountSystemRoutes(mux, ping)

	addr := fmt.Sprintf(":%d", port)
	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info(ctx, "starting "+opts.Name+" on "+addr)

	err = _http.Listen(signalCtx, addr, opts.decorate(mux), timeouts.Shutdown)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCantServe, err)
	}

	return nil
}
