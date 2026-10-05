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

	got, _ := tableExists(context.Background())

	assert.True(t, got)
}

func TestMakeTableExists_failsWhenQueryFails(t *testing.T) {
	want := errors.New("query error")
	selectOneOp := database.MockSelectOne[bool](false, want)

	tableExists := migration.MakeTableExists(selectOneOp)

	_, got := tableExists(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeAppliedNames_success(t *testing.T) {
	selectOp := database.MockSelect[string]([]string{"000_foundation.sql"}, nil)

	appliedNames := migration.MakeAppliedNames(selectOp)

	want := []string{"000_foundation.sql"}
	got, _ := appliedNames(context.Background())

	assert.Equal(t, want, got)
}

func TestMakeAppliedNames_failsWhenQueryFails(t *testing.T) {
	want := errors.New("query error")
	selectOp := database.MockSelect[string](nil, want)

	appliedNames := migration.MakeAppliedNames(selectOp)

	_, got := appliedNames(context.Background())

	assert.ErrorIs(t, got, want)
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
	got, _ := status(context.Background())

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
	got, _ := status(context.Background())

	assert.Equal(t, want, got)
}

func TestMakeStatus_failsWhenListFilesFails(t *testing.T) {
	want := errors.New("list files failed")
	mockListFiles := migration.MockListFiles(nil, want)
	mockTableExists := migration.MockTableExists(false, errors.New("must not be called"))
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("must not be called"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	_, got := status(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeStatus_failsWhenTableExistsFails(t *testing.T) {
	mockListFiles := migration.MockListFiles([]string{"000_foundation.sql"}, nil)
	mockTableExists := migration.MockTableExists(false, errors.New("table check failed"))
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("must not be called"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	want := migration.ErrFailedToCheckTableExists
	_, got := status(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeStatus_failsWhenTableExistsFailsKeepsCause(t *testing.T) {
	mockListFiles := migration.MockListFiles([]string{"000_foundation.sql"}, nil)
	want := errors.New("table check failed")
	mockTableExists := migration.MockTableExists(false, want)
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("must not be called"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	_, got := status(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeStatus_failsWhenAppliedNamesFails(t *testing.T) {
	mockListFiles := migration.MockListFiles([]string{"000_foundation.sql"}, nil)
	mockTableExists := migration.MockTableExists(true, nil)
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("select failed"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	want := migration.ErrFailedToSelectAppliedNames
	_, got := status(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeStatus_failsWhenAppliedNamesFailsKeepsCause(t *testing.T) {
	mockListFiles := migration.MockListFiles([]string{"000_foundation.sql"}, nil)
	mockTableExists := migration.MockTableExists(true, nil)
	want := errors.New("select failed")
	mockAppliedNames := migration.MockAppliedNames(nil, want)

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	_, got := status(context.Background())

	assert.ErrorIs(t, got, want)
}
