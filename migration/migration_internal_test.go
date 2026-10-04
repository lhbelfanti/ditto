package migration

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
)

const lockQuery = "SELECT pg_advisory_xact_lock($1)"

func writeMigrationFile(t *testing.T, content string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "001_a.sql")
	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func TestBeginLocked_success(t *testing.T) {
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	want := tx
	got, _ := beginLocked(context.Background(), db)

	assert.Equal(t, want, got)
}

func TestBeginLocked_failsWhenBeginFails(t *testing.T) {
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return((*database.MockPgxTx)(nil), errors.New("begin failed"))

	want := ErrFailedToApply
	_, got := beginLocked(context.Background(), db)

	assert.ErrorIs(t, got, want)
}

func TestBeginLocked_failsWhenLockFails(t *testing.T) {
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, lockQuery, mock.Anything).Return(pgconn.CommandTag{}, errors.New("lock failed"))
	tx.On("Rollback", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	want := ErrFailedToApply
	_, got := beginLocked(context.Background(), db)

	assert.ErrorIs(t, got, want)
}

func TestBeginLocked_failsWhenCreateTableFails(t *testing.T) {
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, lockQuery, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, errors.New("create failed"))
	tx.On("Rollback", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	want := ErrFailedToCreateTable
	_, got := beginLocked(context.Background(), db)

	assert.ErrorIs(t, got, want)
}

func TestBeginLocked_failsWhenCreateTableFailsRollsBack(t *testing.T) {
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, lockQuery, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, errors.New("create failed"))
	tx.On("Rollback", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	_, _ = beginLocked(context.Background(), db)

	tx.AssertCalled(t, "Rollback", mock.Anything)
}

func TestApplyFile_success(t *testing.T) {
	file := writeMigrationFile(t, "CREATE TABLE a();")
	insertRow := &database.MockPgxRow{}
	insertRow.On("Scan", mock.Anything).Return(nil)
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(MockPgxRowBool(false)).Once()
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(insertRow)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	_ = applyFile(context.Background(), db, file)

	tx.AssertCalled(t, "Exec", mock.Anything, "CREATE TABLE a();", mock.Anything)
}

func TestApplyFile_successWhenFileIsAlreadyApplied(t *testing.T) {
	file := writeMigrationFile(t, "CREATE TABLE a();")
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(MockPgxRowBool(true))
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	got := applyFile(context.Background(), db, file)

	assert.NoError(t, got)
}

func TestApplyFile_failsWhenBeginFails(t *testing.T) {
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return((*database.MockPgxTx)(nil), errors.New("begin failed"))

	want := ErrFailedToApply
	got := applyFile(context.Background(), db, "001_a.sql")

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenAppliedCheckFails(t *testing.T) {
	row := &database.MockPgxRow{}
	row.On("Scan", mock.Anything).Return(errors.New("scan failed"))
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(row)
	tx.On("Rollback", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	want := ErrFailedToCheckApplied
	got := applyFile(context.Background(), db, "001_a.sql")

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenFileCannotBeRead(t *testing.T) {
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(MockPgxRowBool(false))
	tx.On("Rollback", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	want := ErrUnableToReadFile
	got := applyFile(context.Background(), db, filepath.Join(t.TempDir(), "missing.sql"))

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenFileExecutionFails(t *testing.T) {
	file := writeMigrationFile(t, "SELECT broken;")
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, "SELECT broken;", mock.Anything).Return(pgconn.CommandTag{}, errors.New("syntax error"))
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(MockPgxRowBool(false))
	tx.On("Rollback", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	want := ErrFailedToExecute
	got := applyFile(context.Background(), db, file)

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenTrackingInsertFails(t *testing.T) {
	file := writeMigrationFile(t, "CREATE TABLE a();")
	insertRow := &database.MockPgxRow{}
	insertRow.On("Scan", mock.Anything).Return(errors.New("insert failed"))
	tx := &database.MockPgxTx{}
	tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(MockPgxRowBool(false)).Once()
	tx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(insertRow)
	tx.On("Rollback", mock.Anything).Return(nil)
	db := &database.MockPostgresConnection{}
	db.On("Begin", mock.Anything).Return(tx, nil)

	want := ErrFailedToInsertApplied
	got := applyFile(context.Background(), db, file)

	assert.ErrorIs(t, got, want)
}
