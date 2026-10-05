package migration

import (
	"context"
	"fmt"
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

	// StatusRunner prints the applied/pending state of every migration file.
	StatusRunner func(ctx context.Context) error

	// PendingFileFinder returns the name of the first migration file still pending.
	PendingFileFinder func(ctx context.Context) (string, error)

	// Dispatch runs the migrations command named by the first argument.
	Dispatch func(ctx context.Context, args []string) error
)

// Error renders only a package-owned message, with the attributed file and SQLSTATE code when present.
func (e *ApplyError) Error() string {
	switch {
	case e.File != "" && e.Code != "":
		return fmt.Sprintf("%s: file %s: postgresql error %s", ErrFailedToApply, e.File, e.Code)
	case e.File != "":
		return fmt.Sprintf("%s: file %s", ErrFailedToApply, e.File)
	case e.AttributionUnavailable && e.Code != "":
		return fmt.Sprintf("%s: file attribution unavailable: postgresql error %s", ErrFailedToApply, e.Code)
	case e.AttributionUnavailable:
		return fmt.Sprintf("%s: file attribution unavailable", ErrFailedToApply)
	case e.Code != "":
		return fmt.Sprintf("%s: postgresql error %s", ErrFailedToApply, e.Code)
	default:
		return ErrFailedToApply.Error()
	}
}

// Unwrap returns ErrFailedToApply and the original runner failure.
func (e *ApplyError) Unwrap() []error {
	return []error{ErrFailedToApply, e.cause}
}
