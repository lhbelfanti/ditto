package database

import (
	"context"
	"time"

	"github.com/lhbelfanti/ditto/v2/env"
)

type (
	// Ping verifies that a database connection is reachable.
	Ping func(ctx context.Context) error

	// Check proves database connectivity within a bounded deadline.
	Check func(ctx context.Context) error
)

// MakeCheck creates a Check function that bounds ping to timeout and, on failure, returns it
// hidden behind ErrDatabaseUnavailable. It never logs itself — Check runs both behind HTTP
// handlers (which already log via response.Send) and in non-HTTP startup paths, so logging is
// the caller's decision, not this primitive's.
func MakeCheck(ping Ping, timeout time.Duration) Check {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := ping(ctx)
		if err != nil {
			return WrapUnavailable(err)
		}

		return nil
	}
}

// MakeCheck creates a Check function bounding pg's own connection pool ping to timeout, so a
// caller never needs to reach past Postgres into its underlying pool to build one.
func (pg *Postgres) MakeCheck(timeout time.Duration) Check {
	return MakeCheck(pg.Database().Ping, timeout)
}

// RequireEnv validates the exact environment variables resolveDatabaseURL/InitPostgres consume —
// POSTGRES_DB_PORT (a valid TCP port) and POSTGRES_DB_NAME/USER/PASS (non-empty) — so a caller can
// fail fast with a clear, credential-safe message before ever attempting a connection.
func RequireEnv(lookup env.Lookup) error {
	_, err := env.RequirePort(lookup, "POSTGRES_DB_PORT")
	if err != nil {
		return err
	}

	for _, key := range []string{"POSTGRES_DB_NAME", "POSTGRES_DB_USER", "POSTGRES_DB_PASS"} {
		_, err := env.RequireValue(lookup, key)
		if err != nil {
			return err
		}
	}

	return nil
}
