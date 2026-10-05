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

func TestMakeRunner_failsWhenBeginFails(t *testing.T) {
	db := migration.MockConnection(nil, errors.New("begin failed"))
	runner := migration.MakeRunner(db, t.TempDir())

	want := migration.ErrFailedToApply
	got := runner(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeRunner_failsWhenLockFails(t *testing.T) {
	tx := migration.MockTx(migration.MockTxConfig{LockErr: errors.New("lock failed")})
	db := migration.MockConnection(tx, nil)
	runner := migration.MakeRunner(db, t.TempDir())

	want := migration.ErrFailedToApply
	got := runner(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeRunner_failsWhenCreateTableFails(t *testing.T) {
	tx := migration.MockTx(migration.MockTxConfig{CreateTableErr: errors.New("create failed")})
	db := migration.MockConnection(tx, nil)
	runner := migration.MakeRunner(db, t.TempDir())

	want := migration.ErrFailedToCreateTable
	got := runner(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeRunner_failsWhenMigrationFileFails(t *testing.T) {
	tx := migration.MockTx(migration.MockTxConfig{FileSQL: "SELECT broken;", FileErr: errors.New("syntax error")})
	db := migration.MockConnection(tx, nil)
	dir := migration.MockMigrationDir(t, map[string]string{"001_a.sql": "SELECT broken;"})
	runner := migration.MakeRunner(db, dir)

	want := migration.ErrFailedToExecute
	got := runner(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeRunner_successWhenFileIsAlreadyApplied(t *testing.T) {
	tx := migration.MockTx(migration.MockTxConfig{Applied: true})
	db := migration.MockConnection(tx, nil)
	dir := migration.MockMigrationDir(t, map[string]string{"001_a.sql": "CREATE TABLE a();"})
	runner := migration.MakeRunner(db, dir)

	got := runner(context.Background())

	assert.NoError(t, got)
}

func TestMakeRunner_successWhenFileIsAlreadyAppliedSkipsExecution(t *testing.T) {
	tx := migration.MockTx(migration.MockTxConfig{Applied: true})
	db := migration.MockConnection(tx, nil)
	dir := migration.MockMigrationDir(t, map[string]string{"001_a.sql": "CREATE TABLE a();"})
	runner := migration.MakeRunner(db, dir)

	_ = runner(context.Background())

	tx.AssertNotCalled(t, "Exec", mock.Anything, "CREATE TABLE a();", mock.Anything)
}

func TestMakeRunner_successWhenRunIsRepeated(t *testing.T) {
	pool := migration.MockPostgresPool(t)
	dir := migration.MockMigrationDir(t, map[string]string{
		"001_audit.sql": "CREATE TABLE mig_audit (id SERIAL PRIMARY KEY, note TEXT);",
		"002_mark.sql":  "INSERT INTO mig_audit (note) VALUES ('applied');",
	})
	runner := migration.MakeRunner(pool, dir)

	_ = runner(context.Background())
	_ = runner(context.Background())

	want := 1
	got := migration.MockRowCount(t, pool, "mig_audit")

	assert.Equal(t, want, got)
}

func TestMakeRunner_successWhenRunIsRepeatedTracksEachFileOnce(t *testing.T) {
	pool := migration.MockPostgresPool(t)
	dir := migration.MockMigrationDir(t, map[string]string{
		"001_audit.sql": "CREATE TABLE mig_audit (id SERIAL PRIMARY KEY, note TEXT);",
		"002_mark.sql":  "INSERT INTO mig_audit (note) VALUES ('applied');",
	})
	runner := migration.MakeRunner(pool, dir)

	_ = runner(context.Background())
	_ = runner(context.Background())

	want := 2
	got := migration.MockRowCount(t, pool, "migrations")

	assert.Equal(t, want, got)
}

func TestMakeRunner_successWhenReplicasRunConcurrently(t *testing.T) {
	pool := migration.MockPostgresPool(t)
	dir := migration.MockMigrationDir(t, map[string]string{
		"001_audit.sql": "CREATE TABLE mig_audit (id SERIAL PRIMARY KEY, note TEXT);",
		"002_mark.sql":  "INSERT INTO mig_audit (note) VALUES ('applied');",
	})
	runner := migration.MakeRunner(pool, dir)

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

	want := 1
	got := migration.MockRowCount(t, pool, "mig_audit")

	assert.Equal(t, want, got)
}

func TestMakeRunner_successWhenReplicasRunConcurrentlyWithoutErrors(t *testing.T) {
	pool := migration.MockPostgresPool(t)
	dir := migration.MockMigrationDir(t, map[string]string{
		"001_audit.sql": "CREATE TABLE mig_audit (id SERIAL PRIMARY KEY, note TEXT);",
	})
	runner := migration.MakeRunner(pool, dir)

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

	want := 0
	got := 0
	for err := range errs {
		if err != nil {
			got++
		}
	}

	assert.Equal(t, want, got)
}

func TestMakeRunner_failsWhenFileFailsLeavesNoTableBehind(t *testing.T) {
	pool := migration.MockPostgresPool(t)
	dir := migration.MockMigrationDir(t, map[string]string{
		"001_partial.sql": "CREATE TABLE mig_partial (id INT); SELECT * FROM does_not_exist;",
	})
	runner := migration.MakeRunner(pool, dir)

	_ = runner(context.Background())

	got := migration.MockTableInDatabase(t, pool, "mig_partial")

	assert.False(t, got)
}

func TestMakeRunner_failsWhenFileFailsLeavesNoTrackingRow(t *testing.T) {
	pool := migration.MockPostgresPool(t)
	dir := migration.MockMigrationDir(t, map[string]string{
		"001_partial.sql": "CREATE TABLE mig_partial (id INT); SELECT * FROM does_not_exist;",
	})
	runner := migration.MakeRunner(pool, dir)

	_ = runner(context.Background())

	want := 0
	got := migration.MockRowCount(t, pool, "migrations")

	assert.Equal(t, want, got)
}
