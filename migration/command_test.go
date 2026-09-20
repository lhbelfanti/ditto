package migration_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/migration"
)

func TestDispatch_SuccessWhenNoArguments(t *testing.T) {
	called := false
	mockApply := func(context.Context) error {
		called = true
		return nil
	}
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))
	out := &bytes.Buffer{}

	got := migration.Dispatch(context.Background(), []string{}, mockApply, mockStatus, out)

	assert.NoError(t, got)
	assert.True(t, called)
}

func TestDispatch_SuccessWhenExplicitApply(t *testing.T) {
	want := errors.New("apply failed")
	mockApply := migration.MockApply(want)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	got := migration.Dispatch(context.Background(), []string{migration.CommandApply}, mockApply, mockStatus, &bytes.Buffer{})

	assert.ErrorIs(t, got, want)
}

func TestDispatch_SuccessWhenStatusReportsRecords(t *testing.T) {
	statusCalled := false
	mockApply := func(context.Context) error {
		return errors.New("must not be called")
	}
	mockStatus := func(context.Context) ([]migration.Record, error) {
		statusCalled = true
		return []migration.Record{
			{Name: "000_foundation.sql", Applied: true},
			{Name: "001_second.sql", Applied: false},
		}, nil
	}
	out := &bytes.Buffer{}

	want := "applied 000_foundation.sql\npending 001_second.sql\n"
	err := migration.Dispatch(context.Background(), []string{migration.CommandStatus}, mockApply, mockStatus, out)
	got := out.String()

	assert.NoError(t, err)
	assert.True(t, statusCalled)
	assert.Equal(t, want, got)
}

func TestDispatch_ErrorWhenStatusQueryFails(t *testing.T) {
	mockApply := migration.MockApply(nil)
	want := errors.New("status query failed")
	mockStatus := migration.MockStatus(nil, want)

	got := migration.Dispatch(context.Background(), []string{migration.CommandStatus}, mockApply, mockStatus, &bytes.Buffer{})

	assert.ErrorIs(t, got, want)
}

func TestDispatch_ErrorWhenCommandIsUnknown(t *testing.T) {
	mockApply := migration.MockApply(nil)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	got := migration.Dispatch(context.Background(), []string{"rollback"}, mockApply, mockStatus, &bytes.Buffer{})

	assert.ErrorIs(t, got, migration.ErrUnknownCommand)
}

func TestDispatch_ErrorWhenTooManyArguments(t *testing.T) {
	mockApply := migration.MockApply(nil)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	got := migration.Dispatch(context.Background(), []string{migration.CommandApply, "extra"}, mockApply, mockStatus, &bytes.Buffer{})

	assert.ErrorIs(t, got, migration.ErrUnknownCommand)
}
