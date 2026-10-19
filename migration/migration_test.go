package migration_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lhbelfanti/ditto/v2/migration"
)

func TestRunner_failsWhenBeginFails(t *testing.T) {
	mockDB := migration.MockConnection(nil, errors.New("begin failed"))

	runner := migration.MakeRunner(mockDB, t.TempDir())

	want := migration.ErrFailedToBeginTransaction
	got := runner(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestRunner_failsWhenLockFails(t *testing.T) {
	mockTx := migration.MockTx(migration.MockTxConfig{LockErr: errors.New("lock failed")})
	mockDB := migration.MockConnection(mockTx, nil)

	runner := migration.MakeRunner(mockDB, t.TempDir())

	want := migration.ErrFailedToLockMigrations
	got := runner(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestRunner_failsWhenCreateTableFails(t *testing.T) {
	mockTx := migration.MockTx(migration.MockTxConfig{CreateTableErr: errors.New("create failed")})
	mockDB := migration.MockConnection(mockTx, nil)

	runner := migration.MakeRunner(mockDB, t.TempDir())

	want := migration.ErrFailedToCreateTable
	got := runner(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestRunner_failsWhenMigrationFileFails(t *testing.T) {
	mockTx := migration.MockTx(migration.MockTxConfig{FileSQL: "SELECT broken;", FileErr: errors.New("syntax error")})
	mockDB := migration.MockConnection(mockTx, nil)
	mockDir := migration.MockMigrationDir(t, map[string]string{"001_a.sql": "SELECT broken;"})

	runner := migration.MakeRunner(mockDB, mockDir)

	want := migration.ErrFailedToExecute
	got := runner(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestRunner_successWhenFileIsAlreadyApplied(t *testing.T) {
	mockTx := migration.MockTx(migration.MockTxConfig{Applied: true})
	mockDB := migration.MockConnection(mockTx, nil)
	mockDir := migration.MockMigrationDir(t, map[string]string{"001_a.sql": "CREATE TABLE a();"})

	runner := migration.MakeRunner(mockDB, mockDir)

	_ = runner(context.Background())

	mockTx.AssertNotCalled(t, "Exec", mock.Anything, "CREATE TABLE a();", mock.Anything)
}

func TestRunner_successWhenRunIsRepeated(t *testing.T) {
	mockPool := migration.MockPostgresPool(t)
	mockDir := migration.MockMigrationDir(t, map[string]string{
		"001_audit.sql": "CREATE TABLE mig_audit (id SERIAL PRIMARY KEY, note TEXT);",
		"002_mark.sql":  "INSERT INTO mig_audit (note) VALUES ('applied');",
	})

	runner := migration.MakeRunner(mockPool, mockDir)

	_ = runner(context.Background())
	_ = runner(context.Background())

	want := map[string]int{"mig_audit": 1, "migrations": 2}
	got := map[string]int{"mig_audit": migration.MockRowCount(t, mockPool, "mig_audit"), "migrations": migration.MockRowCount(t, mockPool, "migrations")}

	assert.Equal(t, want, got)
}

func TestRunner_successWhenReplicasRunConcurrently(t *testing.T) {
	mockPool := migration.MockPostgresPool(t)
	mockDir := migration.MockMigrationDir(t, map[string]string{
		"001_audit.sql": "CREATE TABLE mig_audit (id SERIAL PRIMARY KEY, note TEXT);",
		"002_mark.sql":  "INSERT INTO mig_audit (note) VALUES ('applied');",
	})

	runner := migration.MakeRunner(mockPool, mockDir)

	const replicas = 6
	var wg sync.WaitGroup
	errs := make(chan error, replicas)
	for range replicas {
		wg.Go(func() {
			errs <- runner(context.Background())
		})
	}

	wg.Wait()
	close(errs)

	failures := 0
	for err := range errs {
		if err != nil {
			failures++
		}
	}

	want := map[string]int{"failures": 0, "mig_audit": 1}
	got := map[string]int{"failures": failures, "mig_audit": migration.MockRowCount(t, mockPool, "mig_audit")}

	assert.Equal(t, want, got)
}

func TestRunner_failsWhenFileFailsAfterCreatingATable(t *testing.T) {
	mockPool := migration.MockPostgresPool(t)
	mockDir := migration.MockMigrationDir(t, map[string]string{
		"001_partial.sql": "CREATE TABLE mig_partial (id INT); SELECT * FROM does_not_exist;",
	})

	runner := migration.MakeRunner(mockPool, mockDir)

	_ = runner(context.Background())

	want := map[string]any{"mig_partial exists": false, "migrations rows": 0}
	got := map[string]any{"mig_partial exists": migration.MockTableInDatabase(t, mockPool, "mig_partial"), "migrations rows": migration.MockRowCount(t, mockPool, "migrations")}

	assert.Equal(t, want, got)
}

func TestRunner_failsWhenCommittingTheTrackingTableFails(t *testing.T) {
	mockTx := migration.MockTx(migration.MockTxConfig{CommitErr: errors.New("commit failed")})
	mockDB := migration.MockConnection(mockTx, nil)

	runner := migration.MakeRunner(mockDB, t.TempDir())

	want := migration.ErrFailedToCommitTrackingTable
	got := runner(context.Background())

	assert.ErrorIs(t, got, want)
}
