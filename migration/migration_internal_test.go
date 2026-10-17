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
	mockTx := MockTx(MockTxConfig{})
	mockDB := MockConnection(mockTx, nil)

	want := mockTx
	got, _ := beginLocked(context.Background(), mockDB)

	assert.Equal(t, want, got)
}

func TestBeginLocked_failsWhenBeginFails(t *testing.T) {
	mockDB := MockConnection(nil, errors.New("begin failed"))

	want := ErrFailedToBeginTransaction
	_, got := beginLocked(context.Background(), mockDB)

	assert.ErrorIs(t, got, want)
}

func TestBeginLocked_failsWhenLockFails(t *testing.T) {
	mockTx := MockTx(MockTxConfig{LockErr: errors.New("lock failed")})
	mockDB := MockConnection(mockTx, nil)

	want := ErrFailedToLockMigrations
	_, got := beginLocked(context.Background(), mockDB)

	assert.ErrorIs(t, got, want)
}

func TestBeginLocked_failsWhenCreateTableFails(t *testing.T) {
	mockTx := MockTx(MockTxConfig{CreateTableErr: errors.New("create failed")})
	mockDB := MockConnection(mockTx, nil)

	want := ErrFailedToCreateTable
	_, got := beginLocked(context.Background(), mockDB)

	assert.ErrorIs(t, got, want)
}

func TestBeginLocked_failsWhenCreateTableFailsRollsBack(t *testing.T) {
	mockTx := MockTx(MockTxConfig{CreateTableErr: errors.New("create failed")})
	mockDB := MockConnection(mockTx, nil)

	_, _ = beginLocked(context.Background(), mockDB)

	mockTx.AssertCalled(t, "Rollback", mock.Anything)
}

func TestApplyFile_success(t *testing.T) {
	file := MockMigrationFile(t, "CREATE TABLE a();")
	mockTx := MockTx(MockTxConfig{})
	mockDB := MockConnection(mockTx, nil)

	_ = applyFile(context.Background(), mockDB, file)

	mockTx.AssertCalled(t, "Exec", mock.Anything, "CREATE TABLE a();", mock.Anything)
}

func TestApplyFile_successWhenFileIsAlreadyApplied(t *testing.T) {
	file := MockMigrationFile(t, "CREATE TABLE a();")
	mockTx := MockTx(MockTxConfig{Applied: true})
	mockDB := MockConnection(mockTx, nil)

	got := applyFile(context.Background(), mockDB, file)

	assert.NoError(t, got)
}

func TestApplyFile_failsWhenBeginFails(t *testing.T) {
	mockDB := MockConnection(nil, errors.New("begin failed"))

	want := ErrFailedToBeginTransaction
	got := applyFile(context.Background(), mockDB, "001_a.sql")

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenAppliedCheckFails(t *testing.T) {
	mockTx := MockTx(MockTxConfig{AppliedErr: errors.New("scan failed")})
	mockDB := MockConnection(mockTx, nil)

	want := ErrFailedToCheckApplied
	got := applyFile(context.Background(), mockDB, "001_a.sql")

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenFileCannotBeRead(t *testing.T) {
	mockTx := MockTx(MockTxConfig{})
	mockDB := MockConnection(mockTx, nil)

	want := ErrUnableToReadFile
	got := applyFile(context.Background(), mockDB, filepath.Join(t.TempDir(), "missing.sql"))

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenFileExecutionFails(t *testing.T) {
	file := MockMigrationFile(t, "SELECT broken;")
	mockTx := MockTx(MockTxConfig{FileSQL: "SELECT broken;", FileErr: errors.New("syntax error")})
	mockDB := MockConnection(mockTx, nil)

	want := ErrFailedToExecute
	got := applyFile(context.Background(), mockDB, file)

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenTrackingInsertFails(t *testing.T) {
	file := MockMigrationFile(t, "CREATE TABLE a();")
	mockTx := MockTx(MockTxConfig{InsertErr: errors.New("insert failed")})
	mockDB := MockConnection(mockTx, nil)

	want := ErrFailedToInsertApplied
	got := applyFile(context.Background(), mockDB, file)

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenCommittingAnAppliedFileFails(t *testing.T) {
	mockTx := MockTx(MockTxConfig{Applied: true, CommitErr: errors.New("commit failed")})
	mockDB := MockConnection(mockTx, nil)

	want := ErrFailedToCommitFile
	got := applyFile(context.Background(), mockDB, "001_a.sql")

	assert.ErrorIs(t, got, want)
}

func TestApplyFile_failsWhenCommittingTheFileFails(t *testing.T) {
	file := MockMigrationFile(t, "SELECT 1;")
	mockTx := MockTx(MockTxConfig{CommitErr: errors.New("commit failed")})
	mockDB := MockConnection(mockTx, nil)

	want := ErrFailedToCommitFile
	got := applyFile(context.Background(), mockDB, file)

	assert.ErrorIs(t, got, want)
}
