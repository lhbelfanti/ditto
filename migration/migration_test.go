package migration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lhbelfanti/ditto/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lhbelfanti/ditto/migration"
)

func setupMigrationDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestMakeRunner_AppliesFiles(t *testing.T) {
	dir := setupMigrationDir(t, map[string]string{
		"000_setup.sql": "CREATE TABLE test (id SERIAL);",
	})

	mockConn := new(database.MockPostgresConnection)
	// createTable call
	mockConn.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil).Once()
	// exec SQL file call
	mockConn.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil).Once()

	sel := database.MockSelectOne[bool](false, nil)
	ins := database.MockInsert[int](1, nil)

	runner := migration.MakeRunnerWithDeps(mockConn, sel, ins, dir)
	err := runner(context.Background())

	assert.NoError(t, err)
	mockConn.AssertNumberOfCalls(t, "Exec", 2)
}

func TestMakeRunner_SkipsApplied(t *testing.T) {
	dir := setupMigrationDir(t, map[string]string{
		"000_setup.sql": "CREATE TABLE test (id SERIAL);",
	})

	mockConn := new(database.MockPostgresConnection)
	// createTable call only
	mockConn.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil).Once()

	sel := database.MockSelectOne[bool](true, nil)

	insertCalled := false
	ins := func(ctx context.Context, query string, args ...any) (int, error) {
		insertCalled = true
		return 1, nil
	}

	runner := migration.MakeRunnerWithDeps(mockConn, sel, ins, dir)
	err := runner(context.Background())

	assert.NoError(t, err)
	assert.False(t, insertCalled)
	mockConn.AssertNumberOfCalls(t, "Exec", 1)
}

func TestMakeRunner_CreateTableError(t *testing.T) {
	dir := setupMigrationDir(t, map[string]string{})

	mockConn := new(database.MockPostgresConnection)
	mockConn.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, database.ErrQuery)

	sel := database.MockSelectOne[bool](false, nil)
	ins := database.MockInsert[int](0, nil)

	runner := migration.MakeRunnerWithDeps(mockConn, sel, ins, dir)
	err := runner(context.Background())

	assert.ErrorIs(t, err, migration.ErrFailedToCreateTable)
}
