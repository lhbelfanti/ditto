package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lhbelfanti/ditto/v2/env"
)

type (
	// Connection is an interface created as an abstraction of pgxpool.Pool to be able to mock it
	Connection interface {
		Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
		Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
		QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
		Begin(ctx context.Context) (pgx.Tx, error)
	}

	// Postgres is the representation of a postgres database connection
	Postgres struct {
		db *pgxpool.Pool
	}

	// CollectRows is a wrapper created to be able to mock pgx.CollectRows function
	CollectRows[T any] func(rows pgx.Rows) ([]T, error)

	// Ping verifies that a database connection is reachable.
	Ping func(ctx context.Context) error

	// Check proves database connectivity within a bounded deadline.
	Check func(ctx context.Context) error

	// Select[T] executes a query returning multiple rows.
	Select[T any] func(ctx context.Context, query string, args ...any) ([]T, error)

	// SelectOne[T] executes a query returning exactly one row.
	SelectOne[T any] func(ctx context.Context, query string, args ...any) (T, error)

	// Insert[T] executes an INSERT with a RETURNING clause, scanning the scalar result into T.
	// Use T=int for RETURNING id.
	Insert[T any] func(ctx context.Context, query string, args ...any) (T, error)

	// Delete executes a DELETE statement.
	Delete func(ctx context.Context, query string, args ...any) error

	// Update executes an UPDATE statement.
	// Also use Update for INSERT … ON CONFLICT DO NOTHING (no RETURNING clause).
	Update func(ctx context.Context, query string, args ...any) error

	// SafeError renders only its package-owned sentinel message, while still letting errors.Is
	// identify both the sentinel and the original cause it hides — so driver errors that may carry
	// a credential-bearing DSN never reach a log line or an HTTP response body. Reserved for
	// ErrDatabaseUnavailable/ErrCantInitDatabase, whose cause can be a raw pgx dial/ping error — not
	// a blanket rule for every sentinel: ErrNoRows/ErrQuery/ErrCollect stay bare because a query
	// execution/row-collection failure never carries connection-string detail.
	SafeError struct {
		sentinel error
		cause    error
	}
)

var (
	pgInstance *Postgres
	pgInitErr  error
	pgOnce     sync.Once
)

const databaseURL string = "postgresql://%s:%s@%s:%s/%s?sslmode=disable"

// resolveDatabaseURL reads the connection target from the environment. POSTGRES_DB_HOST defaults
// to the compose service name postgres_db, so services that do not set it keep working unchanged.
func resolveDatabaseURL() string {
	dbUser := env.Get("POSTGRES_DB_USER", "")
	dbPass := env.Get("POSTGRES_DB_PASS", "")
	dbHost := env.Get("POSTGRES_DB_HOST", "postgres_db")
	dbName := env.Get("POSTGRES_DB_NAME", "")
	dbPort := env.Get("POSTGRES_DB_PORT", "")

	return fmt.Sprintf(databaseURL, dbUser, dbPass, dbHost, dbPort, dbName)
}
