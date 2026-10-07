package database

import (
	"context"
	"fmt"
)

const (
	createRole     string = `SELECT format('CREATE ROLE %I LOGIN', $1::text) WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = $1::text)`
	syncPassword   string = `SELECT format('ALTER ROLE %I PASSWORD %L', $1::text, $2::text)`
	createDatabase string = `SELECT format('CREATE DATABASE %I OWNER %I', $1::text, $2::text) WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = $1::text)`
	revokePublic   string = `SELECT format('REVOKE ALL ON DATABASE %I FROM PUBLIC', $1::text)`
)

// MakeProvision returns a Provision that, through execFormatted, creates target's role and database when
// they do not exist, makes the role the database owner, syncs the role's password and revokes the
// database from PUBLIC so no other role can connect to it. Running it again changes nothing. Each
// step fails with its own sentinel (ErrFailedToCreateRole, ErrFailedToSyncPassword,
// ErrFailedToCreateDatabase, ErrFailedToRevokePublic), wrapping the cause.
func MakeProvision(execFormatted ExecFormatted, target Target) Provision {
	return func(ctx context.Context) error {
		err := execFormatted(ctx, createRole, target.Role)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToCreateRole, err)
		}

		err = execFormatted(ctx, syncPassword, target.Role, target.Pass)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToSyncPassword, err)
		}

		err = execFormatted(ctx, createDatabase, target.Name, target.Role)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToCreateDatabase, err)
		}

		err = execFormatted(ctx, revokePublic, target.Name)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToRevokePublic, err)
		}

		return nil
	}
}
