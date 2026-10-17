package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpen_success(t *testing.T) {
	pg, _ := open(context.Background(), "postgresql://u:p@127.0.0.1:5432/d?sslmode=disable")
	defer pg.Close()

	got := pg.Database()

	assert.NotNil(t, got)
}

func TestOpen_failsWhenURLIsMalformed(t *testing.T) {
	want := ErrCantInitDatabase
	_, got := open(context.Background(), "postgresql://u:p@127.0.0.1:not-a-port/d")

	assert.ErrorIs(t, got, want)
}

func TestResolveDatabaseURL_success(t *testing.T) {
	t.Setenv("POSTGRES_DB_USER", "user")
	t.Setenv("POSTGRES_DB_PASS", "pass")
	t.Setenv("POSTGRES_DB_NAME", "app")
	t.Setenv("POSTGRES_DB_PORT", "5432")
	t.Setenv("POSTGRES_DB_HOST", "nebula-postgres")

	want := "postgresql://user:pass@nebula-postgres:5432/app?sslmode=disable"
	got := resolveDatabaseURL()

	assert.Equal(t, want, got)
}

func TestResolveDatabaseURL_successWhenHostIsUnset(t *testing.T) {
	t.Setenv("POSTGRES_DB_USER", "user")
	t.Setenv("POSTGRES_DB_PASS", "pass")
	t.Setenv("POSTGRES_DB_NAME", "app")
	t.Setenv("POSTGRES_DB_PORT", "5432")
	t.Setenv("POSTGRES_DB_HOST", "")

	want := "postgresql://user:pass@postgres_db:5432/app?sslmode=disable"
	got := resolveDatabaseURL()

	assert.Equal(t, want, got)
}

func TestBuildURL_success(t *testing.T) {
	want := "postgresql://user:pass@host:5432/app?sslmode=disable"
	got := buildURL("user", "pass", "host", "5432", "app")

	assert.Equal(t, want, got)
}

func TestBuildURL_successWhenPasswordHasSpecialCharacters(t *testing.T) {
	want := "postgresql://user:p%20w=1%40%2F%3A@host:5432/app?sslmode=disable"
	got := buildURL("user", "p w=1@/:", "host", "5432", "app")

	assert.Equal(t, want, got)
}

func TestResolveAdminURL_success(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "admin")
	t.Setenv("POSTGRES_ADMIN_PASS", "secret")
	t.Setenv("POSTGRES_DB_HOST", "")
	t.Setenv("POSTGRES_DB_PORT", "5432")

	want := "postgresql://admin:secret@postgres_db:5432/postgres?sslmode=disable"
	got, _ := resolveAdminURL()

	assert.Equal(t, want, got)
}

func TestResolveAdminURL_failsWhenCredentialsAreMissing(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "admin")
	t.Setenv("POSTGRES_ADMIN_PASS", "")

	_, got := resolveAdminURL()

	assert.False(t, got)
}
