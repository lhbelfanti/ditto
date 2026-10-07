package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestInitPostgres_success(t *testing.T) {
	t.Setenv("POSTGRES_DB_USER", "u")
	t.Setenv("POSTGRES_DB_PASS", "p")
	t.Setenv("POSTGRES_DB_NAME", "d")
	t.Setenv("POSTGRES_DB_PORT", "5432")
	pg, _ := database.InitPostgres(context.Background())
	defer pg.Close()

	got := pg.Database()

	assert.NotNil(t, got)
}

func TestInitPostgres_failsWhenPortIsMalformed(t *testing.T) {
	t.Setenv("POSTGRES_DB_USER", "u")
	t.Setenv("POSTGRES_DB_PASS", "p")
	t.Setenv("POSTGRES_DB_NAME", "d")
	t.Setenv("POSTGRES_DB_PORT", "not-a-port")

	want := database.ErrCantInitDatabase
	_, got := database.InitPostgres(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestInitPostgres_successWhenCalledAgainAfterAFailure(t *testing.T) {
	t.Setenv("POSTGRES_DB_USER", "u")
	t.Setenv("POSTGRES_DB_PASS", "p")
	t.Setenv("POSTGRES_DB_NAME", "d")
	t.Setenv("POSTGRES_DB_PORT", "not-a-port")
	_, _ = database.InitPostgres(context.Background())
	t.Setenv("POSTGRES_DB_PORT", "5432")
	pg, got := database.InitPostgres(context.Background())
	defer pg.Close()

	assert.NoError(t, got)
}

func TestOpenAdmin_success(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "admin")
	t.Setenv("POSTGRES_ADMIN_PASS", "secret")
	t.Setenv("POSTGRES_DB_PORT", "5432")
	pg, _ := database.OpenAdmin(context.Background())
	defer pg.Close()

	got := pg.Database()

	assert.NotNil(t, got)
}

func TestOpenAdmin_failsWhenAdminIsNotConfigured(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "")
	t.Setenv("POSTGRES_ADMIN_PASS", "")

	want := database.ErrAdminNotConfigured
	_, got := database.OpenAdmin(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestOpenAdmin_failsWhenPortIsMalformed(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "admin")
	t.Setenv("POSTGRES_ADMIN_PASS", "secret")
	t.Setenv("POSTGRES_DB_PORT", "not-a-port")

	want := database.ErrCantInitDatabase
	_, got := database.OpenAdmin(context.Background())

	assert.ErrorIs(t, got, want)
}
