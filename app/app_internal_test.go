package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestProvision_successWhenAdminIsNotConfigured(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "")
	t.Setenv("POSTGRES_ADMIN_PASS", "")

	got := provision(context.Background())

	assert.NoError(t, got)
}

func TestProvision_failsWhenAdminIsUnreachable(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "admin")
	t.Setenv("POSTGRES_ADMIN_PASS", "secret")
	t.Setenv("POSTGRES_DB_HOST", "127.0.0.1")
	t.Setenv("POSTGRES_DB_PORT", "1")

	want := database.ErrDatabaseUnavailable
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestProvision_failsWhenPortIsMalformed(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "admin")
	t.Setenv("POSTGRES_ADMIN_PASS", "secret")
	t.Setenv("POSTGRES_DB_PORT", "not-a-port")

	want := database.ErrCantInitDatabase
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}
