package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/lhbelfanti/ditto/v2/log"
)

// MakeSelect creates a Select[T] backed by db.Query + collectRows.
func MakeSelect[T any](db Connection, collectRows CollectRows[T]) Select[T] {
	return func(ctx context.Context, query string, args ...any) ([]T, error) {
		rows, err := db.Query(ctx, query, args...)
		if err != nil {
			log.Error(ctx, err.Error())
			return nil, ErrQuery
		}

		results, err := collectRows(rows)
		if err != nil {
			log.Error(ctx, err.Error())
			return nil, ErrCollect
		}

		return results, nil
	}
}

// MakeSelectOne creates a SelectOne[T] backed by db.Query + pgx.CollectOneRow.
// Pass nil for fn to use pgx.RowToStructByPos[T] (struct fields mapped by position).
// Pass a custom pgx.RowToFunc[T] to scan scalar types (bool, int, …) or custom structs.
func MakeSelectOne[T any](db Connection, fn pgx.RowToFunc[T]) SelectOne[T] {
	rowTo := fn
	if rowTo == nil {
		rowTo = pgx.RowToStructByPos[T]
	}

	return func(ctx context.Context, query string, args ...any) (T, error) {
		rows, err := db.Query(ctx, query, args...)
		if err != nil {
			log.Error(ctx, err.Error())
			var zero T
			return zero, ErrQuery
		}

		result, err := pgx.CollectOneRow(rows, rowTo)
		if err != nil {
			log.Error(ctx, err.Error())
			var zero T
			if errors.Is(err, pgx.ErrNoRows) {
				return zero, ErrNoRows
			}
			return zero, ErrQuery
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
			log.Error(ctx, err.Error())
			var zero T
			return zero, ErrQuery
		}

		return result, nil
	}
}

// MakeDelete creates a Delete backed by db.Exec.
func MakeDelete(db Connection) Delete {
	return func(ctx context.Context, query string, args ...any) error {
		_, err := db.Exec(ctx, query, args...)
		if err != nil {
			log.Error(ctx, err.Error())
			return ErrQuery
		}

		return nil
	}
}

// MakeUpdate creates an Update backed by db.Exec.
func MakeUpdate(db Connection) Update {
	return func(ctx context.Context, query string, args ...any) error {
		_, err := db.Exec(ctx, query, args...)
		if err != nil {
			log.Error(ctx, err.Error())
			return ErrQuery
		}

		return nil
	}
}
