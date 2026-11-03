package migration_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/migration"
)

func TestDispatch_successWhenNoArguments(t *testing.T) {
	want := errors.New("apply failed")
	mockApply := migration.MockApply(want)
	mockRunStatus := migration.MockStatusRunner(errors.New("must not be called"))

	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	got := dispatch(context.Background(), []string{})

	assert.ErrorIs(t, got, want)
}

func TestDispatch_successWhenApplyIsExplicit(t *testing.T) {
	want := errors.New("apply failed")
	mockApply := migration.MockApply(want)
	mockRunStatus := migration.MockStatusRunner(errors.New("must not be called"))

	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	got := dispatch(context.Background(), []string{migration.CommandApply})

	assert.ErrorIs(t, got, want)
}

func TestDispatch_successWhenStatusIsRequested(t *testing.T) {
	mockApply := migration.MockApply(errors.New("must not be called"))
	mockRunStatus := migration.MockStatusRunner(nil)

	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	got := dispatch(context.Background(), []string{migration.CommandStatus})

	assert.NoError(t, got)
}

func TestDispatch_failsWhenStatusRunnerFails(t *testing.T) {
	mockErr := errors.New("status query failed")
	mockApply := migration.MockApply(nil)
	mockRunStatus := migration.MockStatusRunner(mockErr)

	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	want := migration.ErrFailedToRunStatus
	wantCause := mockErr
	got := dispatch(context.Background(), []string{migration.CommandStatus})

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestDispatch_failsWhenCommandIsUnknown(t *testing.T) {
	mockApply := migration.MockApply(nil)
	mockRunStatus := migration.MockStatusRunner(nil)

	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	want := migration.ErrUnknownCommand
	got := dispatch(context.Background(), []string{"rollback"})

	assert.ErrorIs(t, got, want)
}

func TestDispatch_failsWhenTooManyArguments(t *testing.T) {
	mockApply := migration.MockApply(nil)
	mockRunStatus := migration.MockStatusRunner(nil)

	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	want := migration.ErrUnknownCommand
	got := dispatch(context.Background(), []string{migration.CommandApply, "extra"})

	assert.ErrorIs(t, got, want)
}

func TestStatusRunner_success(t *testing.T) {
	mockStatus := migration.MockStatus(migration.MockRecords(), nil)
	out := &bytes.Buffer{}

	runStatus := migration.MakeStatusRunner(mockStatus, out)

	_ = runStatus(context.Background())

	want := "applied 000_foundation.sql\npending 001_second.sql\npending 002_third.sql\n"
	got := out.String()

	assert.Equal(t, want, got)
}

func TestStatusRunner_failsWhenStatusFails(t *testing.T) {
	mockErr := errors.New("status failed")
	mockStatus := migration.MockStatus(nil, mockErr)

	runStatus := migration.MakeStatusRunner(mockStatus, &bytes.Buffer{})

	want := migration.ErrFailedToGetStatus
	wantCause := mockErr
	got := runStatus(context.Background())

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestStatusRunner_failsWhenWriterFails(t *testing.T) {
	records := []migration.Record{{Name: "000_foundation.sql", Applied: true}}
	mockStatus := migration.MockStatus(records, nil)
	mockOut := migration.MockClosedWriter()

	runStatus := migration.MakeStatusRunner(mockStatus, mockOut)

	want := io.ErrClosedPipe
	got := runStatus(context.Background())

	assert.ErrorIs(t, got, want)
}
