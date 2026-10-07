package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitPostgres opens the connection pool for the service database, whose settings it reads from
// the POSTGRES_DB_* variables. The pool connects lazily, so a bad host only shows up on first use
// or in a Check. A failure to build the pool is returned hidden behind the credential-safe
// ErrCantInitDatabase sentinel — callers never see the raw driver error. Each call opens a new
// pool: call it once per process and close the result.
func InitPostgres(ctx context.Context) (*Postgres, error) {
	return open(ctx, resolveDatabaseURL())
}

// OpenAdmin opens a pool for the database administrator, using POSTGRES_ADMIN_USER and
// POSTGRES_ADMIN_PASS against the maintenance database. It returns ErrAdminNotConfigured when either
// is empty. The caller closes the returned pool.
func OpenAdmin(ctx context.Context) (*Postgres, error) {
	url, ok := resolveAdminURL()
	if !ok {
		return nil, ErrAdminNotConfigured
	}

	return open(ctx, url)
}

// open builds a pool for url, hiding any failure behind ErrCantInitDatabase.
func open(ctx context.Context, url string) (*Postgres, error) {
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, &SafeError{sentinel: ErrCantInitDatabase, cause: err}
	}

	return &Postgres{db}, nil
}
