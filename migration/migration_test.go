package migration_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/migration"
)

func setupMigrationDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestMakeRunner_failsWhenBeginFails(t *testing.T) {
	want := migration.ErrFailedToApply
	errBegin := errors.New("begin failed")
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return((*database.MockPgxTx)(nil), errBegin)

	got := migration.MakeRunner(db, t.TempDir())(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeRunner_failsWhenLockFails(t *testing.T) {
	want := migration.ErrFailedToApply
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, "SELECT pg_advisory_xact_lock($1)", mock.Anything).Return(pgconn.CommandTag{}, errors.New("lock failed"))
	tx.On("Rollback", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	got := migration.MakeRunner(db, t.TempDir())(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeRunner_failsWhenCreateTableFails(t *testing.T) {
	want := migration.ErrFailedToCreateTable
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, "SELECT pg_advisory_xact_lock($1)", mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, errors.New("create failed"))
	tx.On("Rollback", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	got := migration.MakeRunner(db, t.TempDir())(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeRunner_failsWhenMigrationFileFails(t *testing.T) {
	want := migration.ErrFailedToExecute
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, "SELECT broken;", mock.Anything).Return(pgconn.CommandTag{}, errors.New("syntax error"))
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(migration.MockPgxRowBool(false))
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)
	dir := setupMigrationDir(t, map[string]string{"001_a.sql": "SELECT broken;"})

	got := migration.MakeRunner(db, dir)(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeRunner_successWhenFileIsAlreadyApplied(t *testing.T) {
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(migration.MockPgxRowBool(true))
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)
	dir := setupMigrationDir(t, map[string]string{"001_a.sql": "CREATE TABLE a();"})

	err := migration.MakeRunner(db, dir)(context.Background())

	assert.NoError(t, err)
	tx.AssertNotCalled(t, "Exec", mock.Anything, "CREATE TABLE a();", mock.Anything)
}

// The tests below need a reachable Postgres. They are skipped unless DITTO_TEST_DATABASE_URL is set,
// for example postgres://user:pass@localhost:5433/ditto_test?sslmode=disable

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DITTO_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("DITTO_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, "DROP TABLE IF EXISTS migrations, mig_audit, mig_partial CASCADE")
	require.NoError(t, err)
	return pool
}

func countRows(t *testing.T, pool *pgxpool.Pool, table string) int {
	t.Helper()
	var n int
	err := pool.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&n)
	require.NoError(t, err)
	return n
}

func TestMakeRunner_successWhenRunIsRepeated(t *testing.T) {
	pool := testPool(t)
	dir := setupMigrationDir(t, map[string]string{
		"001_audit.sql": "CREATE TABLE mig_audit (id SERIAL PRIMARY KEY, note TEXT);",
		"002_mark.sql":  "INSERT INTO mig_audit (note) VALUES ('applied');",
	})
	runner := migration.MakeRunner(pool, dir)

	require.NoError(t, runner(context.Background()))
	require.NoError(t, runner(context.Background()))

	assert.Equal(t, 1, countRows(t, pool, "mig_audit"))
	assert.Equal(t, 2, countRows(t, pool, "migrations"))
}

func TestMakeRunner_successWhenReplicasRunConcurrently(t *testing.T) {
	pool := testPool(t)
	dir := setupMigrationDir(t, map[string]string{
		"001_audit.sql": "CREATE TABLE mig_audit (id SERIAL PRIMARY KEY, note TEXT);",
		"002_mark.sql":  "INSERT INTO mig_audit (note) VALUES ('applied');",
	})
	runner := migration.MakeRunner(pool, dir)

	const replicas = 6
	var wg sync.WaitGroup
	errs := make(chan error, replicas)
	for i := 0; i < replicas; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- runner(context.Background())
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		assert.NoError(t, err)
	}
	assert.Equal(t, 1, countRows(t, pool, "mig_audit"))
	assert.Equal(t, 2, countRows(t, pool, "migrations"))
}

func TestMakeRunner_failsWhenFileFailsLeavesNothingBehind(t *testing.T) {
	pool := testPool(t)
	dir := setupMigrationDir(t, map[string]string{
		"001_partial.sql": "CREATE TABLE mig_partial (id INT); SELECT * FROM does_not_exist;",
	})

	err := migration.MakeRunner(pool, dir)(context.Background())

	assert.ErrorIs(t, err, migration.ErrFailedToExecute)
	var exists bool
	err = pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'mig_partial')").Scan(&exists)
	require.NoError(t, err)
	assert.False(t, exists, "the table created before the failing statement must be rolled back")
	assert.Equal(t, 0, countRows(t, pool, "migrations"))
}
