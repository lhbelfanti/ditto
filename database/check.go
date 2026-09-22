package database

import (
	"context"
	"time"

	"github.com/lhbelfanti/ditto/v2/env"
	"github.com/lhbelfanti/ditto/v2/log"
)

type (
	// Ping verifies that a database connection is reachable.
	Ping func(ctx context.Context) error

	// Check proves database connectivity within a bounded deadline.
	Check func(ctx context.Context) error
)

// MakeCheck creates a Check function that bounds ping to timeout and, on failure, logs the
// underlying cause once and returns it hidden behind ErrDatabaseUnavailable.
func MakeCheck(ping Ping, timeout time.Duration) Check {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if err := ping(ctx); err != nil {
			safeErr := WrapUnavailable(err)
			log.Err(ctx, safeErr, ErrDatabaseUnavailable.Error())
			return safeErr
		}

		return nil
	}
}

// RequireEnv validates the exact environment variables resolveDatabaseURL/InitPostgres consume —
// POSTGRES_DB_PORT (a valid TCP port) and POSTGRES_DB_NAME/USER/PASS (non-empty) — so a caller can
// fail fast with a clear, credential-safe message before ever attempting a connection.
func RequireEnv(lookup env.Lookup) error {
	if _, err := env.RequirePort(lookup, "POSTGRES_DB_PORT"); err != nil {
		return err
	}

	for _, key := range []string{"POSTGRES_DB_NAME", "POSTGRES_DB_USER", "POSTGRES_DB_PASS"} {
		if _, err := env.RequireValue(lookup, key); err != nil {
			return err
		}
	}

	return nil
}
