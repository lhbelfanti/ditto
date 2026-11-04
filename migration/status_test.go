package migration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/migration"
)

func TestListFiles_success(t *testing.T) {
	mockFiles := map[string]string{
		"001_second.sql":     "SELECT 1;",
		"000_foundation.sql": "SELECT 1;",
		"notes.txt":          "not a migration",
	}
	mockDir := migration.MockMigrationDir(t, mockFiles)

	listFiles := migration.MakeListFiles(mockDir)

	want := []string{"000_foundation.sql", "001_second.sql"}
	got, _ := listFiles()

	assert.Equal(t, want, got)
}

func TestListFiles_successWhenDirectoryIsEmpty(t *testing.T) {
	mockDir := migration.MockMigrationDir(t, nil)

	listFiles := migration.MakeListFiles(mockDir)

	want := []string{}
	got, _ := listFiles()

	assert.Equal(t, want, got)
}

func TestListFiles_failsWhenPatternIsMalformed(t *testing.T) {
	listFiles := migration.MakeListFiles("[")

	want := migration.ErrUnableToReadFile
	_, got := listFiles()

	assert.ErrorIs(t, got, want)
}

func TestTableExists_success(t *testing.T) {
	mockSelectOneOp := database.MockSelectOne[bool](true, nil)

	tableExists := migration.MakeTableExists(mockSelectOneOp)

	got, _ := tableExists(context.Background())

	assert.True(t, got)
}

func TestTableExists_failsWhenQueryFails(t *testing.T) {
	want := errors.New("query error")
	mockSelectOneOp := database.MockSelectOne[bool](false, want)

	tableExists := migration.MakeTableExists(mockSelectOneOp)

	_, got := tableExists(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestAppliedNames_success(t *testing.T) {
	mockNames := []string{"000_foundation.sql"}
	mockSelectOp := database.MockSelect[string](mockNames, nil)

	appliedNames := migration.MakeAppliedNames(mockSelectOp)

	want := []string{"000_foundation.sql"}
	got, _ := appliedNames(context.Background())

	assert.Equal(t, want, got)
}

func TestAppliedNames_failsWhenQueryFails(t *testing.T) {
	want := errors.New("query error")
	mockSelectOp := database.MockSelect[string](nil, want)

	appliedNames := migration.MakeAppliedNames(mockSelectOp)

	_, got := appliedNames(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestStatus_success(t *testing.T) {
	mockFiles := []string{"000_foundation.sql", "001_second.sql"}
	mockListFiles := migration.MockListFiles(mockFiles, nil)
	mockTableExists := migration.MockTableExists(true, nil)
	mockNames := []string{"000_foundation.sql"}
	mockAppliedNames := migration.MockAppliedNames(mockNames, nil)

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	want := []migration.Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
	}
	got, _ := status(context.Background())

	assert.Equal(t, want, got)
}

func TestStatus_successWhenTableDoesNotExist(t *testing.T) {
	mockFiles := []string{"000_foundation.sql"}
	mockListFiles := migration.MockListFiles(mockFiles, nil)
	mockTableExists := migration.MockTableExists(false, nil)
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("must not be called"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	want := []migration.Record{
		{Name: "000_foundation.sql", Applied: false},
	}
	got, _ := status(context.Background())

	assert.Equal(t, want, got)
}

func TestStatus_failsWhenListFilesFails(t *testing.T) {
	mockErr := errors.New("list files failed")
	mockListFiles := migration.MockListFiles(nil, mockErr)
	mockTableExists := migration.MockTableExists(false, errors.New("must not be called"))
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("must not be called"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	want := migration.ErrFailedToListFiles
	_, got := status(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestStatus_failsWhenTableExistsFails(t *testing.T) {
	mockFiles := []string{"000_foundation.sql"}
	mockListFiles := migration.MockListFiles(mockFiles, nil)
	mockErr := errors.New("table check failed")
	mockTableExists := migration.MockTableExists(false, mockErr)
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("must not be called"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	want := migration.ErrFailedToCheckTableExists
	_, got := status(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestStatus_failsWhenAppliedNamesFails(t *testing.T) {
	mockFiles := []string{"000_foundation.sql"}
	mockListFiles := migration.MockListFiles(mockFiles, nil)
	mockTableExists := migration.MockTableExists(true, nil)
	mockErr := errors.New("select failed")
	mockAppliedNames := migration.MockAppliedNames(nil, mockErr)

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)

	want := migration.ErrFailedToSelectAppliedNames
	_, got := status(context.Background())

	assert.ErrorIs(t, got, want)
}
