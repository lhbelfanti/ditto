package migration

import (
	"context"
	"fmt"

	"github.com/lhbelfanti/ditto/database"
)

type (
	CreateTable   func(ctx context.Context) error
	IsApplied     func(ctx context.Context, name string) (bool, error)
	InsertApplied func(ctx context.Context, name string) error
)

func MakeCreateTable(db database.Connection) CreateTable {
	return func(ctx context.Context) error {
		const q = `CREATE TABLE IF NOT EXISTS migrations (
			id         SERIAL PRIMARY KEY,
			name       TEXT UNIQUE NOT NULL,
			applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`
		_, err := db.Exec(ctx, q)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToCreateTable, err)
		}
		return nil
	}
}

func MakeIsApplied(sel database.SelectOne[bool]) IsApplied {
	return func(ctx context.Context, name string) (bool, error) {
		const q = `SELECT EXISTS (SELECT 1 FROM migrations WHERE name = $1)`
		applied, err := sel(ctx, q, name)
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrFailedToCheckApplied, err)
		}
		return applied, nil
	}
}

func MakeInsertApplied(ins database.Insert[int]) InsertApplied {
	return func(ctx context.Context, name string) error {
		const q = `INSERT INTO migrations (name) VALUES ($1) RETURNING id`
		_, err := ins(ctx, q, name)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToInsertApplied, err)
		}
		return nil
	}
}
