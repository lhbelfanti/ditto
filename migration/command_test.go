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
	called := false
	mockApply := migration.MockApplyCapturing(nil, &called)
	mockRunStatus := migration.MockStatusRunner(errors.New("must not be called"), new(bool))
	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	_ = dispatch(context.Background(), []string{})

	assert.True(t, called)
}

func TestDispatch_successWhenApplyIsExplicit(t *testing.T) {
	want := errors.New("apply failed")
	mockApply := migration.MockApply(want)
	mockRunStatus := migration.MockStatusRunner(errors.New("must not be called"), new(bool))
	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	got := dispatch(context.Background(), []string{migration.CommandApply})

	assert.ErrorIs(t, got, want)
}

func TestDispatch_successWhenStatusIsRequested(t *testing.T) {
	called := false
	mockApply := migration.MockApply(errors.New("must not be called"))
	mockRunStatus := migration.MockStatusRunner(nil, &called)
	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	_ = dispatch(context.Background(), []string{migration.CommandStatus})

	assert.True(t, called)
}

func TestDispatch_failsWhenStatusRunnerFails(t *testing.T) {
	want := errors.New("status query failed")
	mockApply := migration.MockApply(nil)
	mockRunStatus := migration.MockStatusRunner(want, new(bool))
	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	got := dispatch(context.Background(), []string{migration.CommandStatus})

	assert.ErrorIs(t, got, want)
}

func TestDispatch_failsWhenCommandIsUnknown(t *testing.T) {
	mockApply := migration.MockApply(nil)
	mockRunStatus := migration.MockStatusRunner(nil, new(bool))
	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	want := migration.ErrUnknownCommand
	got := dispatch(context.Background(), []string{"rollback"})

	assert.ErrorIs(t, got, want)
}

func TestDispatch_failsWhenTooManyArguments(t *testing.T) {
	mockApply := migration.MockApply(nil)
	mockRunStatus := migration.MockStatusRunner(nil, new(bool))
	dispatch := migration.MakeDispatch(mockApply, mockRunStatus)

	want := migration.ErrUnknownCommand
	got := dispatch(context.Background(), []string{migration.CommandApply, "extra"})

	assert.ErrorIs(t, got, want)
}

func TestStatusRunner_success(t *testing.T) {
	records := []migration.Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
	}
	mockStatus := migration.MockStatus(records, nil)
	out := &bytes.Buffer{}
	runStatus := migration.MakeStatusRunner(mockStatus, out)

	_ = runStatus(context.Background())

	want := "applied 000_foundation.sql\npending 001_second.sql\n"
	got := out.String()

	assert.Equal(t, want, got)
}

func TestStatusRunner_failsWhenStatusFails(t *testing.T) {
	want := errors.New("status failed")
	mockStatus := migration.MockStatus(nil, want)
	runStatus := migration.MakeStatusRunner(mockStatus, &bytes.Buffer{})

	got := runStatus(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestStatusRunner_failsWhenWriterFails(t *testing.T) {
	records := []migration.Record{{Name: "000_foundation.sql", Applied: true}}
	mockStatus := migration.MockStatus(records, nil)
	runStatus := migration.MakeStatusRunner(mockStatus, migration.MockClosedWriter())

	want := io.ErrClosedPipe
	got := runStatus(context.Background())

	assert.ErrorIs(t, got, want)
}
