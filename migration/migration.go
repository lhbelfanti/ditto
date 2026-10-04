package migration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"

	"github.com/lhbelfanti/ditto/v2/database"
)

// migrationLockKey identifies ditto's migration lock. Replicas starting together queue on it, and
// whichever wins the lock finds the file already applied.
const migrationLockKey int64 = 7261046501783541

// Runner applies all pending migrations and returns an error on failure.
type Runner func(ctx context.Context) error

// MakeRunner returns a Runner that applies all *.sql files from migrationsDir in lexicographic
// order, skipping already-applied files. Each file runs in its own transaction together with its
// tracking row, under a transaction-scoped advisory lock: concurrent replicas never apply the same
// file twice, and a failing file leaves nothing half-applied.
func MakeRunner(db database.Connection, migrationsDir string) Runner {
	return func(ctx context.Context) error {
		tx, err := beginLocked(ctx, db)
		if err != nil {
			return err
		}
		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToApply, err)
		}

		pattern := filepath.Join(migrationsDir, "*.sql")
		files, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrUnableToReadFile, err)
		}
		sort.Strings(files)

		for _, file := range files {
			err = applyFile(ctx, db, file)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// beginLocked opens a transaction, takes the migration lock and ensures the tracking table exists.
// The caller owns the transaction and must commit or roll it back.
func beginLocked(ctx context.Context, db database.Connection) (pgx.Tx, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToApply, err)
	}

	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", migrationLockKey)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("%w: %w", ErrFailedToApply, err)
	}

	const createTable = `CREATE TABLE IF NOT EXISTS migrations (
		id         SERIAL PRIMARY KEY,
		name       TEXT UNIQUE NOT NULL,
		applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	)`
	_, err = tx.Exec(ctx, createTable)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("%w: %w", ErrFailedToCreateTable, err)
	}

	return tx, nil
}

func applyFile(ctx context.Context, db database.Connection, file string) error {
	tx, err := beginLocked(ctx, db)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	name := filepath.Base(file)

	var applied bool
	err = tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM migrations WHERE name = $1)`, name).Scan(&applied)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToCheckApplied, err)
	}
	if applied {
		return tx.Commit(ctx)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnableToReadFile, err)
	}

	_, err = tx.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToExecute, err)
	}

	var id int
	err = tx.QueryRow(ctx, `INSERT INTO migrations (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToInsertApplied, err)
	}

	return tx.Commit(ctx)
}
