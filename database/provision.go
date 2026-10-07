package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	createRole     string = `SELECT format('CREATE ROLE %I LOGIN', $1::text) WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = $1::text)`
	syncPassword   string = `SELECT format('ALTER ROLE %I PASSWORD %L', $1::text, $2::text)`
	createDatabase string = `SELECT format('CREATE DATABASE %I OWNER %I', $1::text, $2::text) WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = $1::text)`
	revokePublic   string = `SELECT format('REVOKE ALL ON DATABASE %I FROM PUBLIC', $1::text)`
)

// MakeProvision returns a Provision that, through admin, creates target's role and database when
// they do not exist, makes the role the database owner, syncs the role's password and revokes the
// database from PUBLIC so no other role can connect to it. Running it again changes nothing. Each
// step fails with its own sentinel (ErrFailedToCreateRole, ErrFailedToSyncPassword,
// ErrFailedToCreateDatabase, ErrFailedToRevokePublic), wrapping the cause.
func MakeProvision(admin Connection, target Target) Provision {
	return func(ctx context.Context) error {
		err := execFormatted(ctx, admin, createRole, target.Role)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToCreateRole, err)
		}

		err = execFormatted(ctx, admin, syncPassword, target.Role, target.Pass)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToSyncPassword, err)
		}

		err = execFormatted(ctx, admin, createDatabase, target.Name, target.Role)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToCreateDatabase, err)
		}

		err = execFormatted(ctx, admin, revokePublic, target.Name)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToRevokePublic, err)
		}

		return nil
	}
}

// execFormatted runs the statement query builds with format(), letting the server quote the
// identifiers and the password. A query with no row means there is nothing to do. It fails with
// ErrFailedToBuildStatement when the server cannot build the statement and with
// ErrFailedToExecuteStatement when it cannot run it.
func execFormatted(ctx context.Context, admin Connection, query string, args ...any) error {
	var statement string
	err := admin.QueryRow(ctx, query, args...).Scan(&statement)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToBuildStatement, err)
	}

	_, err = admin.Exec(ctx, statement)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToExecuteStatement, err)
	}

	return nil
}
