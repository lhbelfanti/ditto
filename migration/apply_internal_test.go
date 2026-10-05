package migration

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

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
