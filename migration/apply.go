package migration

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	dittohttp "github.com/lhbelfanti/ditto/v2/http"
)

// Apply runs pending migrations through the injected runner and returns a credential-safe,
// filename-attributed error on failure.
type Apply func(ctx context.Context) error

// ApplyError renders only a package-owned message, an optional attributed filename, and an
// optional allowlisted PostgreSQL SQLSTATE classification, while still letting errors.Is/errors.As
// reach the original runner failure through Unwrap.
type ApplyError struct {
	File                   string
	Code                   string
	AttributionUnavailable bool
	cause                  error
}

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

func (e *ApplyError) Unwrap() []error {
	return []error{ErrFailedToApply, e.cause}
}

// MakeApply wraps runner with deterministic, credential-safe failure attribution: on failure it
// extracts the allowlisted PostgreSQL SQLSTATE (never the raw driver error) and, when the failure
// happened during file execution, the name of the file that failed.
func MakeApply(runner dittohttp.MigrationRunner, status Status) Apply {
	return func(ctx context.Context) error {
		err := runner(ctx)
		if err == nil {
			return nil
		}

		code := pgErrorCode(err)

		if !errors.Is(err, ErrFailedToExecute) {
			return &ApplyError{Code: code, cause: err}
		}

		file, attribErr := firstPendingFile(ctx, status)
		if attribErr != nil {
			return &ApplyError{Code: code, AttributionUnavailable: true, cause: err}
		}

		return &ApplyError{File: file, Code: code, cause: err}
	}
}

// firstPendingFile takes a read-only post-failure status snapshot and returns the first file
// still pending, relying on MakeRunner's guaranteed lexicographic, sequential apply order to
// identify the file whose execution failed.
func firstPendingFile(ctx context.Context, status Status) (string, error) {
	records, err := status(ctx)
	if err != nil {
		return "", err
	}

	for _, record := range records {
		if !record.Applied {
			return record.Name, nil
		}
	}

	return "", errors.New("migration: no pending migration found in post-failure status snapshot")
}

// pgErrorCode extracts the allowlisted PostgreSQL SQLSTATE classification from err, if present. It
// never returns arbitrary dependency error text.
func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}

	return ""
}
