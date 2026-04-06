package migration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/lhbelfanti/ditto/database"
	dittohttp "github.com/lhbelfanti/ditto/http"
	"github.com/lhbelfanti/ditto/log"
)

// MakeRunner returns a MigrationRunner that applies all *.sql files from migrationsDir
// in lexicographic order, skipping already-applied files.
func MakeRunner(db database.Connection, migrationsDir string) dittohttp.MigrationRunner {
	sel := database.MakeSelectOne[bool](db, func(row pgx.CollectableRow) (bool, error) {
		var v bool
		return v, row.Scan(&v)
	})
	ins := database.MakeInsert[int](db)
	return MakeRunnerWithDeps(db, sel, ins, migrationsDir)
}

// MakeRunnerWithDeps is the injectable variant of MakeRunner used in tests.
func MakeRunnerWithDeps(db database.Connection, sel database.SelectOne[bool], ins database.Insert[int], migrationsDir string) dittohttp.MigrationRunner {
	createTable := MakeCreateTable(db)
	isApplied := MakeIsApplied(sel)
	insertApplied := MakeInsertApplied(ins)

	return func(ctx context.Context) error {
		if err := createTable(ctx); err != nil {
			return err
		}

		pattern := filepath.Join(migrationsDir, "*.sql")
		files, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrUnableToReadFile, err)
		}
		sort.Strings(files)

		for _, file := range files {
			name := filepath.Base(file)

			applied, err := isApplied(ctx, name)
			if err != nil {
				return err
			}
			if applied {
				continue
			}

			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrUnableToReadFile, err)
			}

			if _, execErr := db.Exec(ctx, string(content)); execErr != nil {
				log.Error(ctx, execErr.Error())
				return fmt.Errorf("%w: %w", ErrFailedToExecute, execErr)
			}

			if err := insertApplied(ctx, name); err != nil {
				return err
			}
		}

		return nil
	}
}
