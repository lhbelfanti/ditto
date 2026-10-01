package migration

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/lhbelfanti/ditto/v2/database"
)

const (
	queryTableExists  = `SELECT to_regclass('migrations') IS NOT NULL`
	queryAppliedNames = `SELECT name FROM migrations`
)

type (
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

// MakeListFiles creates a ListFiles function that globs *.sql files under dir and returns their
// sorted basenames, matching MakeRunner's discovery order.
func MakeListFiles(dir string) ListFiles {
	return func() ([]string, error) {
		pattern := filepath.Join(dir, "*.sql")
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnableToReadFile, err)
		}

		names := make([]string, len(files))
		for i, file := range files {
			names[i] = filepath.Base(file)
		}
		sort.Strings(names)

		return names, nil
	}
}

// MakeTableExists creates a TableExists function backed by the injected read-only scalar select.
func MakeTableExists(selectOne database.SelectOne[bool]) TableExists {
	return func(ctx context.Context) (bool, error) {
		return selectOne(ctx, queryTableExists)
	}
}

// MakeAppliedNames creates an AppliedNames function backed by the injected read-only select.
func MakeAppliedNames(selectMany database.Select[string]) AppliedNames {
	return func(ctx context.Context) ([]string, error) {
		return selectMany(ctx, queryAppliedNames)
	}
}

// MakeStatus creates a Status function that classifies every migration file as applied or
// pending. A tracking table that does not yet exist is a normal all-pending state, not an error,
// so a genuinely clean database can be inspected without issuing any DDL or DML.
func MakeStatus(listFiles ListFiles, tableExists TableExists, appliedNames AppliedNames) Status {
	return func(ctx context.Context) ([]Record, error) {
		names, err := listFiles()
		if err != nil {
			return nil, err
		}

		exists, err := tableExists(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToCheckTableExists, err)
		}

		applied := map[string]bool{}
		if exists {
			appliedList, err := appliedNames(ctx)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrFailedToSelectAppliedNames, err)
			}
			for _, name := range appliedList {
				applied[name] = true
			}
		}

		records := make([]Record, len(names))
		for i, name := range names {
			records[i] = Record{Name: name, Applied: applied[name]}
		}

		return records, nil
	}
}
