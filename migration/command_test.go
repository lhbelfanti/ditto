package migration_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/migration"
)

func TestDispatch_successWhenNoArguments(t *testing.T) {
	called := false
	mockApply := migration.MockApplyCapturing(nil, &called)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	_ = migration.Dispatch(context.Background(), []string{}, mockApply, mockStatus, &bytes.Buffer{})

	assert.True(t, called)
}

func TestDispatch_successWhenApplyIsExplicit(t *testing.T) {
	want := errors.New("apply failed")
	mockApply := migration.MockApply(want)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	got := migration.Dispatch(context.Background(), []string{migration.CommandApply}, mockApply, mockStatus, &bytes.Buffer{})

	assert.ErrorIs(t, got, want)
}

func TestDispatch_successWhenStatusReportsRecords(t *testing.T) {
	records := []migration.Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
	}
	mockApply := migration.MockApply(errors.New("must not be called"))
	mockStatus := migration.MockStatus(records, nil)
	out := &bytes.Buffer{}

	_ = migration.Dispatch(context.Background(), []string{migration.CommandStatus}, mockApply, mockStatus, out)

	want := "applied 000_foundation.sql\npending 001_second.sql\n"
	got := out.String()

	assert.Equal(t, want, got)
}

func TestDispatch_failsWhenStatusQueryFails(t *testing.T) {
	mockApply := migration.MockApply(nil)
	want := errors.New("status query failed")
	mockStatus := migration.MockStatus(nil, want)

	got := migration.Dispatch(context.Background(), []string{migration.CommandStatus}, mockApply, mockStatus, &bytes.Buffer{})

	assert.ErrorIs(t, got, want)
}

func TestDispatch_failsWhenCommandIsUnknown(t *testing.T) {
	mockApply := migration.MockApply(nil)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	got := migration.Dispatch(context.Background(), []string{"rollback"}, mockApply, mockStatus, &bytes.Buffer{})

	assert.ErrorIs(t, got, migration.ErrUnknownCommand)
}

func TestDispatch_failsWhenTooManyArguments(t *testing.T) {
	mockApply := migration.MockApply(nil)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	got := migration.Dispatch(context.Background(), []string{migration.CommandApply, "extra"}, mockApply, mockStatus, &bytes.Buffer{})

	assert.ErrorIs(t, got, migration.ErrUnknownCommand)
}
