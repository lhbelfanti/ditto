package migration

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBeginLocked_success(t *testing.T) {
	tx := MockTx(MockTxConfig{})
	db := MockConnection(tx, nil)

	want := tx
	got, _ := beginLocked(context.Background(), db)

	assert.Equal(t, want, got)
}

func TestBeginLocked_failsWhenBeginFails(t *testing.T) {
	db := MockConnection(nil, errors.New("begin failed"))

	want := ErrFailedToBeginTransaction
	_, got := beginLocked(context.Background(), db)

	assert.ErrorIs(t, got, want)
}

func TestBeginLocked_failsWhenLockFails(t *testing.T) {
	tx := MockTx(MockTxConfig{LockErr: errors.New("lock failed")})
	db := MockConnection(tx, nil)

	want := ErrFailedToLockMigrations
	_, got := beginLocked(context.Background(), db)

	assert.ErrorIs(t, got, want)
}

func TestBeginLocked_failsWhenCreateTableFails(t *testing.T) {
	tx := MockTx(MockTxConfig{CreateTableErr: errors.New("create failed")})
	db := MockConnection(tx, nil)

	want := ErrFailedToCreateTable
	_, got := beginLocked(context.Background(), db)

	assert.ErrorIs(t, got, want)
}

func TestBeginLocked_failsWhenCreateTableFailsRollsBack(t *testing.T) {
	tx := MockTx(MockTxConfig{CreateTableErr: errors.New("create failed")})
	db := MockConnection(tx, nil)

	_, _ = beginLocked(context.Background(), db)

	tx.AssertCalled(t, "Rollback", mock.Anything)
}

func TestApplyFile_success(t *testing.T) {
	file := MockMigrationFile(t, "CREATE TABLE a();")
	tx := MockTx(MockTxConfig{})
	db := MockConnection(tx, nil)

	_ = applyFile(context.Background(), db, file)

	tx.AssertCalled(t, "Exec", mock.Anything, "CREATE TABLE a();", mock.Anything)
}

func TestApplyFile_successWhenFileIsAlreadyApplied(t *testing.T) {
	file := MockMigrationFile(t, "CREATE TABLE a();")
	tx := MockTx(MockTxConfig{Applied: true})
	db := MockConnection(tx, nil)

	got := applyFile(context.Background(), db, file)

	assert.NoError(t, got)
}

func TestApplyFile_failsWhenBeginFails(t *testing.T) {
	db := MockConnection(nil, errors.New("begin failed"))

	want := ErrFailedToBeginTransaction
	got := applyFile(context.Background(), db, "001_a.sql")

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenAppliedCheckFails(t *testing.T) {
	tx := MockTx(MockTxConfig{AppliedErr: errors.New("scan failed")})
	db := MockConnection(tx, nil)

	want := ErrFailedToCheckApplied
	got := applyFile(context.Background(), db, "001_a.sql")

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenFileCannotBeRead(t *testing.T) {
	tx := MockTx(MockTxConfig{})
	db := MockConnection(tx, nil)

	want := ErrUnableToReadFile
	got := applyFile(context.Background(), db, filepath.Join(t.TempDir(), "missing.sql"))

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenFileExecutionFails(t *testing.T) {
	file := MockMigrationFile(t, "SELECT broken;")
	tx := MockTx(MockTxConfig{FileSQL: "SELECT broken;", FileErr: errors.New("syntax error")})
	db := MockConnection(tx, nil)

	want := ErrFailedToExecute
	got := applyFile(context.Background(), db, file)

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenTrackingInsertFails(t *testing.T) {
	file := MockMigrationFile(t, "CREATE TABLE a();")
	tx := MockTx(MockTxConfig{InsertErr: errors.New("insert failed")})
	db := MockConnection(tx, nil)

	want := ErrFailedToInsertApplied
	got := applyFile(context.Background(), db, file)

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenCommittingAnAppliedFileFails(t *testing.T) {
	tx := MockTx(MockTxConfig{Applied: true, CommitErr: errors.New("commit failed")})
	db := MockConnection(tx, nil)

	want := ErrFailedToCommitFile
	got := applyFile(context.Background(), db, "001_a.sql")

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenCommittingTheFileFails(t *testing.T) {
	file := MockMigrationFile(t, "SELECT 1;")
	tx := MockTx(MockTxConfig{CommitErr: errors.New("commit failed")})
	db := MockConnection(tx, nil)

	want := ErrFailedToCommitFile
	got := applyFile(context.Background(), db, file)

	assert.ErrorIs(t, got, want)
}
