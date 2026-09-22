package database

import (
	"context"

	"github.com/lhbelfanti/ditto/v2/env"
)

type (
	// Ping verifies that a database connection is reachable.
	Ping func(ctx context.Context) error

	// Check proves database connectivity, hiding any underlying failure behind a credential-safe sentinel.
	Check func(ctx context.Context) error
)

// MakeCheck creates a Check function. Its only job is hiding ping's raw failure — which may embed
// a credential-bearing DSN — behind ErrDatabaseUnavailable; it never logs itself, since Check runs
// both behind HTTP handlers (which already log via response.Send) and in non-HTTP startup paths.
// Bounding ctx with a deadline before calling Check, like any other context.Context-based call, is
// the caller's job — Check has no timeout of its own to configure.
func MakeCheck(ping Ping) Check {
	return func(ctx context.Context) error {
		err := ping(ctx)
		if err != nil {
			return WrapUnavailable(err)
		}

		return nil
	}
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
