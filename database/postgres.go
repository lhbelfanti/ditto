package database

import (
	"context"
	"net"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lhbelfanti/ditto/v2/env"
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

// resolveDatabaseURL reads the connection target from the environment. POSTGRES_DB_HOST defaults
// to the compose service name postgres_db, so services that do not set it keep working unchanged.
func resolveDatabaseURL() string {
	dbUser := env.Get("POSTGRES_DB_USER", "")
	dbPass := env.Get("POSTGRES_DB_PASS", "")
	dbHost := env.Get("POSTGRES_DB_HOST", "postgres_db")
	dbName := env.Get("POSTGRES_DB_NAME", "")
	dbPort := env.Get("POSTGRES_DB_PORT", "")

	return buildURL(dbUser, dbPass, dbHost, dbPort, dbName)
}

// buildURL assembles the connection URL, escaping the credentials so a password holding spaces,
// "@", "/", ":" or "=" cannot break it.
func buildURL(user, pass, host, port, name string) string {
	u := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(user, pass),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + name,
		RawQuery: "sslmode=disable",
	}

	return u.String()
}

// resolveAdminURL reads the administrator connection target from the environment: the same host
// and port as the service database, the maintenance database, and POSTGRES_ADMIN_USER/PASS. It
// returns false when either credential is empty.
func resolveAdminURL() (string, bool) {
	// The service's own database may not exist yet, so the administrator connects to the maintenance one.
	const adminDatabase string = "postgres"

	adminUser := env.Get("POSTGRES_ADMIN_USER", "")
	adminPass := env.Get("POSTGRES_ADMIN_PASS", "")
	if adminUser == "" || adminPass == "" {
		return "", false
	}

	dbHost := env.Get("POSTGRES_DB_HOST", "postgres_db")
	dbPort := env.Get("POSTGRES_DB_PORT", "")

	return buildURL(adminUser, adminPass, dbHost, dbPort, adminDatabase), true
}
