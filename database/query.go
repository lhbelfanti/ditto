package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// MakeSelect creates a Select[T] backed by db.Query + collectRows.
func MakeSelect[T any](db Connection, collectRows CollectRows[T]) Select[T] {
	return func(ctx context.Context, query string, args ...any) ([]T, error) {
		rows, err := db.Query(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrQuery, err)
		}

		results, err := collectRows(rows)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCollect, err)
		}

		return results, nil
	}
}

// MakeSelectOne creates a SelectOne[T] backed by db.Query + pgx.CollectOneRow.
// Pass nil for fn to use pgx.RowToStructByPos[T] (struct fields mapped by position).
// Pass a custom pgx.RowToFunc[T] to scan scalar types (bool, int, …) or custom structs.
func MakeSelectOne[T any](db Connection, fn pgx.RowToFunc[T]) SelectOne[T] {
	rowToFunc := fn
	if rowToFunc == nil {
		rowToFunc = pgx.RowToStructByPos[T]
	}

	return func(ctx context.Context, query string, args ...any) (T, error) {
		var zero T

		rows, err := db.Query(ctx, query, args...)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrQuery, err)
		}

		result, err := pgx.CollectOneRow(rows, rowToFunc)
		if errors.Is(err, pgx.ErrNoRows) {
			return zero, fmt.Errorf("%w: %w", ErrNoRows, err)
		}
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrQuery, err)
		}

		return result, nil
	}
}

// MakeInsert creates an Insert[T] backed by db.QueryRow + Scan for a single scalar return.
func MakeInsert[T any](db Connection) Insert[T] {
	return func(ctx context.Context, query string, args ...any) (T, error) {
		var result T
		err := db.QueryRow(ctx, query, args...).Scan(&result)
		if err != nil {
			var zero T

			return zero, fmt.Errorf("%w: %w", ErrQuery, err)
		}

		return result, nil
	}
}

// MakeDelete creates a Delete backed by db.Exec.
func MakeDelete(db Connection) Delete {
	return func(ctx context.Context, query string, args ...any) error {
		_, err := db.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrQuery, err)
		}

		return nil
	}
}

// MakeUpdate creates an Update backed by db.Exec.
func MakeUpdate(db Connection) Update {
	return func(ctx context.Context, query string, args ...any) error {
		_, err := db.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrQuery, err)
		}

		return nil
	}
}
