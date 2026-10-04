package database

import (
	"context"
	"fmt"
	"os"
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
	dbUser := os.Getenv("POSTGRES_DB_USER")
	dbPass := os.Getenv("POSTGRES_DB_PASS")
	dbHost := env.Get("POSTGRES_DB_HOST", "postgres_db")
	dbName := os.Getenv("POSTGRES_DB_NAME")
	dbPort := os.Getenv("POSTGRES_DB_PORT")

	return fmt.Sprintf(databaseURL, dbUser, dbPass, dbHost, dbPort, dbName)
}
