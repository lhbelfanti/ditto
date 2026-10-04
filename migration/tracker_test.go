package migration_test

import (
	"context"
	"testing"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/migration"
)

func TestMakeCreateTable_Success(t *testing.T) {
	mockConn := database.MockExecConnection(nil)
	createTable := migration.MakeCreateTable(mockConn)

	got := createTable(context.Background())

	assert.NoError(t, got)
}

func TestMakeCreateTable_Error(t *testing.T) {
	mockConn := database.MockExecConnection(database.ErrQuery)
	createTable := migration.MakeCreateTable(mockConn)

	want := migration.ErrFailedToCreateTable
	got := createTable(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeIsApplied_Applied(t *testing.T) {
	sel := database.MockSelectOne[bool](true, nil)

	isApplied := migration.MakeIsApplied(sel)
	result, err := isApplied(context.Background(), "000_setup.sql")

	assert.NoError(t, err)
	assert.True(t, result)
}

func TestMakeIsApplied_NotApplied(t *testing.T) {
	sel := database.MockSelectOne[bool](false, nil)

	isApplied := migration.MakeIsApplied(sel)
	result, err := isApplied(context.Background(), "000_setup.sql")

	assert.NoError(t, err)
	assert.False(t, result)
}

func TestMakeIsApplied_Error(t *testing.T) {
	sel := database.MockSelectOne[bool](false, database.ErrQuery)

	isApplied := migration.MakeIsApplied(sel)
	result, err := isApplied(context.Background(), "000_setup.sql")

	assert.False(t, result)
	assert.ErrorIs(t, err, migration.ErrFailedToCheckApplied)
}

func TestMakeInsertApplied_Success(t *testing.T) {
	ins := database.MockInsert[int](1, nil)

	insertApplied := migration.MakeInsertApplied(ins)
	err := insertApplied(context.Background(), "000_setup.sql")

	assert.NoError(t, err)
}

func TestMakeInsertApplied_Error(t *testing.T) {
	ins := database.MockInsert[int](0, database.ErrQuery)

	insertApplied := migration.MakeInsertApplied(ins)
	err := insertApplied(context.Background(), "000_setup.sql")

	assert.ErrorIs(t, err, migration.ErrFailedToInsertApplied)
}
