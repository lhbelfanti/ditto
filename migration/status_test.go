package migration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/migration"
)

func TestMakeListFiles_success(t *testing.T) {
	dir := migration.MockMigrationDir(t, map[string]string{
		"001_second.sql":     "SELECT 1;",
		"000_foundation.sql": "SELECT 1;",
		"notes.txt":          "not a migration",
	})
	listFiles := migration.MakeListFiles(dir)

	want := []string{"000_foundation.sql", "001_second.sql"}
	got, _ := listFiles()

	assert.Equal(t, want, got)
}

func TestMakeListFiles_successWhenDirectoryIsEmpty(t *testing.T) {
	dir := migration.MockMigrationDir(t, nil)
	listFiles := migration.MakeListFiles(dir)

	want := []string{}
	got, _ := listFiles()

	assert.Equal(t, want, got)
}

func TestMakeListFiles_failsWhenPatternIsMalformed(t *testing.T) {
	listFiles := migration.MakeListFiles("[")

	want := migration.ErrUnableToReadFile
	_, got := listFiles()

	assert.ErrorIs(t, got, want)
}

func TestMakeTableExists_success(t *testing.T) {
	selectOneOp := database.MockSelectOne[bool](true, nil)

	tableExists := migration.MakeTableExists(selectOneOp)

	got, err := tableExists(context.Background())

	assert.NoError(t, err)
	assert.True(t, got)
}

func TestMakeTableExists_failsWhenQueryFails(t *testing.T) {
	underlying := errors.New("query error")
	selectOneOp := database.MockSelectOne[bool](false, underlying)

	tableExists := migration.MakeTableExists(selectOneOp)
	got, err := tableExists(context.Background())

	assert.ErrorIs(t, err, underlying)
	assert.False(t, got)
}

func TestMakeAppliedNames_success(t *testing.T) {
	selectOp := database.MockSelect[string]([]string{"000_foundation.sql"}, nil)

	appliedNames := migration.MakeAppliedNames(selectOp)

	want := []string{"000_foundation.sql"}
	got, err := appliedNames(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMakeAppliedNames_failsWhenQueryFails(t *testing.T) {
	underlying := errors.New("query error")
	selectOp := database.MockSelect[string](nil, underlying)

	appliedNames := migration.MakeAppliedNames(selectOp)
	got, err := appliedNames(context.Background())

	assert.ErrorIs(t, err, underlying)
	assert.Nil(t, got)
}

func TestMakeStatus_success(t *testing.T) {
	mockListFiles := migration.MockListFiles([]string{"000_foundation.sql", "001_second.sql"}, nil)
	mockTableExists := migration.MockTableExists(true, nil)
	mockAppliedNames := migration.MockAppliedNames([]string{"000_foundation.sql"}, nil)

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	want := []migration.Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
	}
	got, err := status(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMakeStatus_successWhenTableDoesNotExist(t *testing.T) {
	mockListFiles := migration.MockListFiles([]string{"000_foundation.sql"}, nil)
	mockTableExists := migration.MockTableExists(false, nil)
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("must not be called"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	want := []migration.Record{
		{Name: "000_foundation.sql", Applied: false},
	}
	got, err := status(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMakeStatus_failsWhenListFilesFails(t *testing.T) {
	want := errors.New("list files failed")
	mockListFiles := migration.MockListFiles(nil, want)
	mockTableExists := migration.MockTableExists(false, errors.New("must not be called"))
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("must not be called"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)
	got, err := status(context.Background())

	assert.ErrorIs(t, err, want)
	assert.Nil(t, got)
}

func TestMakeStatus_failsWhenTableExistsFails(t *testing.T) {
	mockListFiles := migration.MockListFiles([]string{"000_foundation.sql"}, nil)
	underlying := errors.New("table check failed")
	mockTableExists := migration.MockTableExists(false, underlying)
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("must not be called"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)
	got, err := status(context.Background())

	assert.ErrorIs(t, err, migration.ErrFailedToCheckTableExists)
	assert.ErrorIs(t, err, underlying)
	assert.Nil(t, got)
}

func TestMakeStatus_failsWhenAppliedNamesFails(t *testing.T) {
	mockListFiles := migration.MockListFiles([]string{"000_foundation.sql"}, nil)
	mockTableExists := migration.MockTableExists(true, nil)
	underlying := errors.New("select failed")
	mockAppliedNames := migration.MockAppliedNames(nil, underlying)

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)
	got, err := status(context.Background())

	assert.ErrorIs(t, err, migration.ErrFailedToSelectAppliedNames)
	assert.ErrorIs(t, err, underlying)
	assert.Nil(t, got)
}
