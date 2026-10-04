package migration

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/mock"

	"github.com/lhbelfanti/ditto/v2/database"
)

type (
	MockTxConfig struct {
		LockErr        error
		CreateTableErr error
		FileSQL        string
		FileErr        error
		Applied        bool
		AppliedErr     error
		InsertErr      error
	}
)

// MockListFiles returns a ListFiles that always returns the given names and error.
func MockListFiles(names []string, err error) ListFiles {
	return func() ([]string, error) {
		return names, err
	}
}

// MockTableExists returns a TableExists that always returns the given value and error.
func MockTableExists(exists bool, err error) TableExists {
	return func(context.Context) (bool, error) {
		return exists, err
	}
}

// MockAppliedNames returns an AppliedNames that always returns the given names and error.
func MockAppliedNames(names []string, err error) AppliedNames {
	return func(context.Context) ([]string, error) {
		return names, err
	}
}

// MockStatus returns a Status that always returns the given records and error.
func MockStatus(records []Record, err error) Status {
	return func(context.Context) ([]Record, error) {
		return records, err
	}
}

// MockApply returns an Apply that always returns err.
func MockApply(err error) Apply {
	return func(context.Context) error {
		return err
	}
}

// MockRunner returns a Runner that always returns err.
func MockRunner(err error) Runner {
	return func(context.Context) error {
		return err
	}
}

// MockPgxRowBool returns a pgx.Row whose Scan writes applied into a *bool destination.
func MockPgxRowBool(applied bool) *database.MockPgxRow {
	row := &database.MockPgxRow{}
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).([]any)[0].(*bool) = applied
	}).Return(nil)
	return row
}

func MockApplyCapturing(err error, called *bool) Apply {
	return func(context.Context) error {
		*called = true
		return err
	}
}

func MockStatusCapturing(records []Record, err error, called *bool) Status {
	return func(context.Context) ([]Record, error) {
		*called = true
		return records, err
	}
}

func MockTx(cfg MockTxConfig) *database.MockPgxTx {
	appliedRow := MockPgxRowBool(cfg.Applied)
	if cfg.AppliedErr != nil {
		appliedRow = database.MockRowFailing(cfg.AppliedErr)
	}

	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, "SELECT pg_advisory_xact_lock($1)", mock.Anything).Return(pgconn.CommandTag{}, cfg.LockErr)
	tx.On("Exec", mock.Anything, mock.MatchedBy(func(query string) bool {
		return strings.HasPrefix(query, "CREATE TABLE IF NOT EXISTS migrations")
	}), mock.Anything).Return(pgconn.CommandTag{}, cfg.CreateTableErr)
	if cfg.FileSQL != "" {
		tx.On("Exec", mock.Anything, cfg.FileSQL, mock.Anything).Return(pgconn.CommandTag{}, cfg.FileErr)
	}
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(appliedRow).Once()
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(database.MockRowFailing(cfg.InsertErr))
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)
	return tx
}

func MockConnection(tx *database.MockPgxTx, beginErr error) *database.MockPostgresConnection {
	conn := &database.MockPostgresConnection{}
	conn.On("Begin", mock.Anything).Return(tx, beginErr)
	return conn
}

func MockClosedWriter() io.Writer {
	reader, writer := io.Pipe()
	_ = reader.Close()
	return writer
}

func MockMigrationDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600)
		if err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func MockMigrationFile(t *testing.T, content string) string {
	t.Helper()
	dir := MockMigrationDir(t, map[string]string{"001_a.sql": content})
	return filepath.Join(dir, "001_a.sql")
}

func MockPostgresPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DITTO_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("DITTO_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, "DROP TABLE IF EXISTS migrations, mig_audit, mig_partial CASCADE")
	if err != nil {
		t.Fatal(err)
	}
	return pool
}

func MockRowCount(t *testing.T, pool *pgxpool.Pool, table string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func MockTableInDatabase(t *testing.T, pool *pgxpool.Pool, table string) bool {
	t.Helper()
	var exists bool
	err := pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", table).Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}
	return exists
}
