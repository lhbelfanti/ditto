package migration

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestFirstPendingFile_success(t *testing.T) {
	records := []Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
		{Name: "002_third.sql", Applied: false},
	}
	mockStatus := MockStatus(records, nil)

	want := "001_second.sql"
	got, _ := firstPendingFile(context.Background(), mockStatus)

	assert.Equal(t, want, got)
}

func TestFirstPendingFile_failsWhenStatusFails(t *testing.T) {
	want := errors.New("status failed")
	mockStatus := MockStatus(nil, want)

	_, got := firstPendingFile(context.Background(), mockStatus)

	assert.ErrorIs(t, got, want)
}

func TestFirstPendingFile_failsWhenNoFileIsPending(t *testing.T) {
	records := []Record{{Name: "000_foundation.sql", Applied: true}}
	mockStatus := MockStatus(records, nil)

	_, got := firstPendingFile(context.Background(), mockStatus)

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
