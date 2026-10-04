package migration_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
