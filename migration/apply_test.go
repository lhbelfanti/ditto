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
	mockFinder := migration.MockPendingFileFinder("", errors.New("must not be called"))

	apply := migration.MakeApply(mockRunner, mockFinder)

	got := apply(context.Background())

	assert.NoError(t, got)
}

func TestMakeApply_failsWhenExecutionFails(t *testing.T) {
	runnerErr := fmt.Errorf("%w: %w", migration.ErrFailedToExecute, &pgconn.PgError{Code: "42601"})
	mockRunner := migration.MockRunner(runnerErr)
	mockFinder := migration.MockPendingFileFinder("001_second.sql", nil)

	apply := migration.MakeApply(mockRunner, mockFinder)

	want := &migration.ApplyError{File: "001_second.sql", Code: "42601"}
	var got *migration.ApplyError
	_ = errors.As(apply(context.Background()), &got)

	assert.Equal(t, want.Error(), got.Error())
}

func TestMakeApply_failsWhenExecutionFailsKeepsCause(t *testing.T) {
	want := fmt.Errorf("%w: %w", migration.ErrFailedToExecute, &pgconn.PgError{Code: "42601"})
	mockRunner := migration.MockRunner(want)
	mockFinder := migration.MockPendingFileFinder("001_second.sql", nil)

	apply := migration.MakeApply(mockRunner, mockFinder)

	got := apply(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeApply_failsWhenTrackingTableSetupFails(t *testing.T) {
	runnerErr := fmt.Errorf("%w: %w", migration.ErrFailedToCreateTable, errors.New("connection refused"))
	mockRunner := migration.MockRunner(runnerErr)
	mockFinder := migration.MockPendingFileFinder("", errors.New("must not be called"))

	apply := migration.MakeApply(mockRunner, mockFinder)

	want := &migration.ApplyError{}
	var got *migration.ApplyError
	_ = errors.As(apply(context.Background()), &got)

	assert.Equal(t, want.Error(), got.Error())
}

func TestMakeApply_failsWhenDiagnosticSnapshotFails(t *testing.T) {
	runnerErr := fmt.Errorf("%w: %w", migration.ErrFailedToExecute, errors.New("execution error"))
	mockRunner := migration.MockRunner(runnerErr)
	mockFinder := migration.MockPendingFileFinder("", errors.New("status query failed"))

	apply := migration.MakeApply(mockRunner, mockFinder)

	want := &migration.ApplyError{AttributionUnavailable: true}
	var got *migration.ApplyError
	_ = errors.As(apply(context.Background()), &got)

	assert.Equal(t, want.Error(), got.Error())
}

func TestPendingFileFinder_success(t *testing.T) {
	records := []migration.Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
		{Name: "002_third.sql", Applied: false},
	}
	mockStatus := migration.MockStatus(records, nil)
	findPendingFile := migration.MakePendingFileFinder(mockStatus)

	want := "001_second.sql"
	got, _ := findPendingFile(context.Background())

	assert.Equal(t, want, got)
}

func TestPendingFileFinder_failsWhenStatusFails(t *testing.T) {
	want := errors.New("status failed")
	mockStatus := migration.MockStatus(nil, want)
	findPendingFile := migration.MakePendingFileFinder(mockStatus)

	_, got := findPendingFile(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestPendingFileFinder_failsWhenNoFileIsPending(t *testing.T) {
	records := []migration.Record{{Name: "000_foundation.sql", Applied: true}}
	mockStatus := migration.MockStatus(records, nil)
	findPendingFile := migration.MakePendingFileFinder(mockStatus)

	_, got := findPendingFile(context.Background())

	assert.Error(t, got)
}
