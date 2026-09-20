package migration_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/migration"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	require := assert.New(t)
	require.NoError(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}

func TestMakeListFiles_Success(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "001_second.sql", "SELECT 1;")
	writeFile(t, dir, "000_foundation.sql", "SELECT 1;")
	writeFile(t, dir, "notes.txt", "not a migration")

	listFiles := migration.MakeListFiles(dir)

	want := []string{"000_foundation.sql", "001_second.sql"}
	got, err := listFiles()

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMakeListFiles_SuccessWhenDirectoryIsEmpty(t *testing.T) {
	dir := t.TempDir()

	listFiles := migration.MakeListFiles(dir)

	want := []string{}
	got, err := listFiles()

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMakeListFiles_ErrorWhenPatternIsMalformed(t *testing.T) {
	listFiles := migration.MakeListFiles("[")
	got, err := listFiles()

	assert.ErrorIs(t, err, migration.ErrUnableToReadFile)
	assert.Nil(t, got)
}

func TestMakeTableExists_Success(t *testing.T) {
	selectOneOp := database.MockSelectOne[bool](true, nil)

	tableExists := migration.MakeTableExists(selectOneOp)

	want := true
	got, err := tableExists(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMakeTableExists_ErrorWhenQueryFails(t *testing.T) {
	underlying := errors.New("query error")
	selectOneOp := database.MockSelectOne[bool](false, underlying)

	tableExists := migration.MakeTableExists(selectOneOp)
	got, err := tableExists(context.Background())

	assert.ErrorIs(t, err, underlying)
	assert.False(t, got)
}

func TestMakeAppliedNames_Success(t *testing.T) {
	selectOp := database.MockSelect[string]([]string{"000_foundation.sql"}, nil)

	appliedNames := migration.MakeAppliedNames(selectOp)

	want := []string{"000_foundation.sql"}
	got, err := appliedNames(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMakeAppliedNames_ErrorWhenQueryFails(t *testing.T) {
	underlying := errors.New("query error")
	selectOp := database.MockSelect[string](nil, underlying)

	appliedNames := migration.MakeAppliedNames(selectOp)
	got, err := appliedNames(context.Background())

	assert.ErrorIs(t, err, underlying)
	assert.Nil(t, got)
}

func TestMakeStatus_Success(t *testing.T) {
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

func TestMakeStatus_SuccessWhenTableDoesNotExist(t *testing.T) {
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

func TestMakeStatus_ErrorWhenListFilesFails(t *testing.T) {
	want := errors.New("list files failed")
	mockListFiles := migration.MockListFiles(nil, want)
	mockTableExists := migration.MockTableExists(false, errors.New("must not be called"))
	mockAppliedNames := migration.MockAppliedNames(nil, errors.New("must not be called"))

	status := migration.MakeStatus(mockListFiles, mockTableExists, mockAppliedNames)
	got, err := status(context.Background())

	assert.ErrorIs(t, err, want)
	assert.Nil(t, got)
}

func TestMakeStatus_ErrorWhenTableExistsFails(t *testing.T) {
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

func TestMakeStatus_ErrorWhenAppliedNamesFails(t *testing.T) {
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
