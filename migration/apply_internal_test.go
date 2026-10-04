package migration

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestPendingFileFinder_success(t *testing.T) {
	records := []Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
		{Name: "002_third.sql", Applied: false},
	}
	mockStatus := MockStatus(records, nil)

	findPendingFile := makePendingFileFinder(mockStatus)

	want := "001_second.sql"
	got, _ := findPendingFile(context.Background())

	assert.Equal(t, want, got)
}

func TestPendingFileFinder_failsWhenStatusFails(t *testing.T) {
	want := errors.New("status failed")
	mockStatus := MockStatus(nil, want)

	findPendingFile := makePendingFileFinder(mockStatus)

	_, got := findPendingFile(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestPendingFileFinder_failsWhenNoFileIsPending(t *testing.T) {
	records := []Record{{Name: "000_foundation.sql", Applied: true}}
	mockStatus := MockStatus(records, nil)

	findPendingFile := makePendingFileFinder(mockStatus)

	_, got := findPendingFile(context.Background())

	assert.Error(t, got)
}

func TestPgErrorCode_success(t *testing.T) {
	err := fmt.Errorf("%w: %w", ErrFailedToExecute, &pgconn.PgError{Code: "42601"})

	want := "42601"
	got := pgErrorCode(err)

	assert.Equal(t, want, got)
}

func TestPgErrorCode_successWhenErrorIsNotPostgres(t *testing.T) {
	err := errors.New("connection refused")

	want := ""
	got := pgErrorCode(err)

	assert.Equal(t, want, got)
}
