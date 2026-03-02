package database

import "errors"

var (
	// ErrNoRows is returned when a query finds no matching rows.
	ErrNoRows = errors.New("database: no rows found")

	// ErrQuery is returned when a query execution fails.
	ErrQuery = errors.New("database: query execution failed")

	// ErrCollect is returned when row collection fails after a successful query.
	ErrCollect = errors.New("database: failed to collect rows")
)
