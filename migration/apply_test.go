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

func TestApply_success(t *testing.T) {
	mockRunner := migration.MockRunner(nil)
	mockFinder := migration.MockPendingFileFinder("", errors.New("must not be called"))

	apply := migration.MakeApply(mockRunner, mockFinder)

	got := apply(context.Background())

	assert.NoError(t, got)
}

func TestApply_failsWhenExecutionFails(t *testing.T) {
	mockErr := fmt.Errorf("%w: %w", migration.ErrFailedToExecute, &pgconn.PgError{Code: "42601"})
	mockRunner := migration.MockRunner(mockErr)
	mockFinder := migration.MockPendingFileFinder("001_second.sql", nil)

	apply := migration.MakeApply(mockRunner, mockFinder)

	err := apply(context.Background())
	var target *migration.ApplyError
	_ = errors.As(err, &target)

	want := (&migration.ApplyError{File: "001_second.sql", Code: "42601"}).Error()
	got := target.Error()
	wantCause := mockErr

	assert.Equal(t, want, got)
	assert.ErrorIs(t, err, wantCause)
}

func TestApply_failsWhenTrackingTableSetupFails(t *testing.T) {
	mockErr := fmt.Errorf("%w: %w", migration.ErrFailedToCreateTable, errors.New("connection refused"))
	mockRunner := migration.MockRunner(mockErr)
	mockFinder := migration.MockPendingFileFinder("", errors.New("must not be called"))

	apply := migration.MakeApply(mockRunner, mockFinder)

	var target *migration.ApplyError
	_ = errors.As(apply(context.Background()), &target)

	want := (&migration.ApplyError{}).Error()
	got := target.Error()

	assert.Equal(t, want, got)
}

func TestApply_failsWhenDiagnosticSnapshotFails(t *testing.T) {
	mockErr := fmt.Errorf("%w: %w", migration.ErrFailedToExecute, errors.New("execution error"))
	mockRunner := migration.MockRunner(mockErr)
	mockFinder := migration.MockPendingFileFinder("", errors.New("status query failed"))

	apply := migration.MakeApply(mockRunner, mockFinder)

	var target *migration.ApplyError
	_ = errors.As(apply(context.Background()), &target)

	want := (&migration.ApplyError{AttributionUnavailable: true}).Error()
	got := target.Error()

	assert.Equal(t, want, got)
}

func TestPendingFileFinder_success(t *testing.T) {
	mockStatus := migration.MockStatus(migration.MockRecords(), nil)

	findPendingFile := migration.MakePendingFileFinder(mockStatus)

	want := "001_second.sql"
	got, _ := findPendingFile(context.Background())

	assert.Equal(t, want, got)
}

func TestPendingFileFinder_failsWhenStatusFails(t *testing.T) {
	mockErr := errors.New("status failed")
	mockStatus := migration.MockStatus(nil, mockErr)

	findPendingFile := migration.MakePendingFileFinder(mockStatus)

	_, got := findPendingFile(context.Background())
	want := migration.ErrFailedToFindPendingFile
	wantCause := mockErr

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestPendingFileFinder_failsWhenNoFileIsPending(t *testing.T) {
	records := []migration.Record{{Name: "000_foundation.sql", Applied: true}}
	mockStatus := migration.MockStatus(records, nil)

	findPendingFile := migration.MakePendingFileFinder(mockStatus)

	_, got := findPendingFile(context.Background())
	want := migration.ErrNoPendingFile

	assert.ErrorIs(t, got, want)
}
