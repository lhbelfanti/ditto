package migration

import (
	"context"
)

type (
	// Apply runs pending migrations through the injected runner and returns a credential-safe,
	// filename-attributed error on failure.
	Apply func(ctx context.Context) error

	// ApplyError renders only a package-owned message, an optional attributed filename, and an
	// optional allowlisted PostgreSQL SQLSTATE classification, while still letting errors.Is/errors.As
	// reach the original runner failure through Unwrap.
	ApplyError struct {
		File                   string
		Code                   string
		AttributionUnavailable bool
		cause                  error
	}

	CreateTable   func(ctx context.Context) error
	IsApplied     func(ctx context.Context, name string) (bool, error)
	InsertApplied func(ctx context.Context, name string) error

	// Runner applies all pending migrations and returns an error on failure.
	Runner func(ctx context.Context) error

	// ListFiles returns the basenames of every top-level migration SQL file, sorted
	// lexicographically to match MakeRunner's apply order.
	ListFiles func() ([]string, error)

	// TableExists reports whether the migrations tracking table is visible on the current
	// search path, without creating it.
	TableExists func(ctx context.Context) (bool, error)

	// AppliedNames returns the basenames recorded as applied in the migrations tracking table.
	AppliedNames func(ctx context.Context) ([]string, error)

	// Record describes one migration file's applied/pending state.
	Record struct {
		Name    string
		Applied bool
	}

	// Status reports the applied/pending state of every migration file without mutating the
	// database.
	Status func(ctx context.Context) ([]Record, error)
)
