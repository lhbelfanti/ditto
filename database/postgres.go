package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InitPostgres creates a new postgres instance. A connection failure is returned hidden behind
// the credential-safe ErrCantInitDatabase sentinel — callers never see the raw driver error.
// The attempt runs at most once: every call, including ones after a failed attempt, returns the
// same cached instance and the same cached error, rather than silently reporting a nil error on
// retry.
func InitPostgres() (*Postgres, error) {
	pgOnce.Do(func() {
		db, err := pgxpool.New(context.Background(), resolveDatabaseURL())
		if err != nil {
			pgInitErr = WrapInitFailure(err)
			return
		}
		pgInstance = &Postgres{db}
	})

	return pgInstance, pgInitErr
}

// OpenAdmin opens a pool for the database administrator, using POSTGRES_ADMIN_USER and
// POSTGRES_ADMIN_PASS against the maintenance database. It returns ErrAdminNotConfigured when either
// is empty. Unlike InitPostgres it is not cached: the caller closes the returned pool.
func OpenAdmin() (*Postgres, error) {
	url, ok := resolveAdminURL()
	if !ok {
		return nil, ErrAdminNotConfigured
	}

	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, WrapInitFailure(err)
	}

	return &Postgres{db}, nil
}

// MakeCollectRows creates a new CollectRows
func MakeCollectRows[T any](fn pgx.RowToFunc[T]) CollectRows[T] {
	return func(rows pgx.Rows) ([]T, error) {
		if fn == nil {
			return pgx.CollectRows(rows, pgx.RowToStructByPos[T])
		}

		return pgx.CollectRows(rows, fn)
	}
}
