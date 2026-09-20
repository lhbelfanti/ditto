package migration_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"

	dittohttp "github.com/lhbelfanti/ditto/v2/http"
	"github.com/lhbelfanti/ditto/v2/migration"
)

func TestMakeApply_Success(t *testing.T) {
	mockRunner := dittohttp.MockMigrationRunner(nil)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	apply := migration.MakeApply(mockRunner, mockStatus)
	got := apply(context.Background())

	assert.NoError(t, got)
}

func TestMakeApply_ErrorWithFileAttributionWhenExecutionFails(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "42601"}
	runnerErr := fmt.Errorf("%w: %w", migration.ErrFailedToExecute, pgErr)
	mockRunner := dittohttp.MockMigrationRunner(runnerErr)
	records := []migration.Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
	}
	mockStatus := migration.MockStatus(records, nil)

	apply := migration.MakeApply(mockRunner, mockStatus)
	got := apply(context.Background())

	var applyErr *migration.ApplyError
	assert.ErrorAs(t, got, &applyErr)
	assert.ErrorIs(t, got, migration.ErrFailedToApply)
	assert.ErrorIs(t, got, migration.ErrFailedToExecute)
	assert.Equal(t, "001_second.sql", applyErr.File)
	assert.Equal(t, "42601", applyErr.Code)
}

func TestMakeApply_ErrorWithoutFileAttributionWhenFailureIsPreFile(t *testing.T) {
	runnerErr := fmt.Errorf("%w: %w", migration.ErrFailedToCreateTable, errors.New("connection refused"))
	mockRunner := dittohttp.MockMigrationRunner(runnerErr)
	mockStatus := migration.MockStatus(nil, errors.New("must not be called"))

	apply := migration.MakeApply(mockRunner, mockStatus)
	got := apply(context.Background())

	var applyErr *migration.ApplyError
	assert.ErrorAs(t, got, &applyErr)
	assert.ErrorIs(t, got, migration.ErrFailedToApply)
	assert.Equal(t, "", applyErr.File)
}

func TestMakeApply_ErrorWithoutFileAttributionWhenDiagnosticSnapshotFails(t *testing.T) {
	runnerErr := fmt.Errorf("%w: %w", migration.ErrFailedToExecute, errors.New("execution error"))
	mockRunner := dittohttp.MockMigrationRunner(runnerErr)
	mockStatus := migration.MockStatus(nil, errors.New("status query failed"))

	apply := migration.MakeApply(mockRunner, mockStatus)
	got := apply(context.Background())

	var applyErr *migration.ApplyError
	assert.ErrorAs(t, got, &applyErr)
	assert.ErrorIs(t, got, migration.ErrFailedToApply)
	assert.True(t, applyErr.AttributionUnavailable)
	assert.Equal(t, "", applyErr.File)
}
