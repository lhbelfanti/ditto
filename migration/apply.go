package migration

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// MakeApply wraps runner with deterministic, credential-safe failure attribution: on failure it
// extracts the allowlisted PostgreSQL SQLSTATE (never the raw driver error) and, when the failure
// happened during file execution, the name of the file that failed.
func MakeApply(runner Runner, findPendingFile PendingFileFinder) Apply {
	return func(ctx context.Context) error {
		err := runner(ctx)
		if err == nil {
			return nil
		}

		code := pgErrorCode(err)

		if !errors.Is(err, ErrFailedToExecute) {
			return &ApplyError{Code: code, cause: err}
		}

		file, attribErr := findPendingFile(ctx)
		if attribErr != nil {
			return &ApplyError{Code: code, AttributionUnavailable: true, cause: err}
		}

		return &ApplyError{File: file, Code: code, cause: err}
	}
}

// MakePendingFileFinder creates a PendingFileFinder that takes a read-only post-failure status
// snapshot and returns the first file still pending, relying on MakeRunner's guaranteed
// lexicographic, sequential apply order to identify the file whose execution failed.
func MakePendingFileFinder(status Status) PendingFileFinder {
	return func(ctx context.Context) (string, error) {
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
}

// pgErrorCode extracts the allowlisted PostgreSQL SQLSTATE classification from err, if present. It
// never returns arbitrary dependency error text.
func pgErrorCode(err error) string {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		return pgErr.Code
	}

	return ""
}
