package migration_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/migration"
)

func TestMakeApply_success(t *testing.T) {
	mockRunner := migration.MockRunner(nil)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	apply := migration.MakeApply(mockRunner, mockStatus)

	got := apply(context.Background())

	assert.NoError(t, got)
}

func TestMakeApply_failsWhenExecutionFails(t *testing.T) {
	runnerErr := fmt.Errorf("%w: %w", migration.ErrFailedToExecute, &pgconn.PgError{Code: "42601"})
	mockRunner := migration.MockRunner(runnerErr)
	records := []migration.Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
	}
	mockStatus := migration.MockStatus(records, nil)

	apply := migration.MakeApply(mockRunner, mockStatus)

	want := &migration.ApplyError{File: "001_second.sql", Code: "42601"}
	var got *migration.ApplyError
	_ = errors.As(apply(context.Background()), &got)

	assert.Equal(t, want.Error(), got.Error())
}

func TestMakeApply_failsWhenExecutionFailsKeepsCause(t *testing.T) {
	want := fmt.Errorf("%w: %w", migration.ErrFailedToExecute, &pgconn.PgError{Code: "42601"})
	mockRunner := migration.MockRunner(want)
	mockStatus := migration.MockStatus([]migration.Record{{Name: "001_second.sql"}}, nil)

	apply := migration.MakeApply(mockRunner, mockStatus)

	got := apply(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeApply_failsWhenTrackingTableSetupFails(t *testing.T) {
	runnerErr := fmt.Errorf("%w: %w", migration.ErrFailedToCreateTable, errors.New("connection refused"))
	mockRunner := migration.MockRunner(runnerErr)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	apply := migration.MakeApply(mockRunner, mockStatus)

	want := &migration.ApplyError{}
	var got *migration.ApplyError
	_ = errors.As(apply(context.Background()), &got)

	assert.Equal(t, want.Error(), got.Error())
}

func TestMakeApply_failsWhenDiagnosticSnapshotFails(t *testing.T) {
	runnerErr := fmt.Errorf("%w: %w", migration.ErrFailedToExecute, errors.New("execution error"))
	mockRunner := migration.MockRunner(runnerErr)
	mockStatus := migration.MockStatus(nil, errors.New("status query failed"))

	apply := migration.MakeApply(mockRunner, mockStatus)

	want := &migration.ApplyError{AttributionUnavailable: true}
	var got *migration.ApplyError
	_ = errors.As(apply(context.Background()), &got)

	assert.Equal(t, want.Error(), got.Error())
}
