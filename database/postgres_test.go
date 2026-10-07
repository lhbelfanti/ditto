package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestInitPostgres_failsWhenCalledAgainAfterInitializationFailure(t *testing.T) {
	// InitPostgres backs a singleton (sync.Once): the first call in this test binary decides every
	// later one, so each test here sets the same failing environment. A malformed port fails
	// pgxpool.New's config parsing immediately, with no network I/O.
	t.Setenv("POSTGRES_DB_USER", "u")
	t.Setenv("POSTGRES_DB_PASS", "p")
	t.Setenv("POSTGRES_DB_NAME", "d")
	t.Setenv("POSTGRES_DB_PORT", "not-a-port")
	_, _ = database.InitPostgres()

	want := database.ErrCantInitDatabase
	_, got := database.InitPostgres()

	assert.ErrorIs(t, got, want)
}

func TestInitPostgres_failsWhenInitializationFails(t *testing.T) {
	t.Setenv("POSTGRES_DB_USER", "u")
	t.Setenv("POSTGRES_DB_PASS", "p")
	t.Setenv("POSTGRES_DB_NAME", "d")
	t.Setenv("POSTGRES_DB_PORT", "not-a-port")

	want := database.ErrCantInitDatabase
	_, got := database.InitPostgres()

	assert.ErrorIs(t, got, want)
}

func TestInitPostgres_failsWhenCalledAgainAfterInitializationFailureReturnsSameInstance(t *testing.T) {
	first, _ := database.InitPostgres()

	want := first
	got, _ := database.InitPostgres()

	assert.Equal(t, want, got)
}

func TestOpenAdmin_success(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "admin")
	t.Setenv("POSTGRES_ADMIN_PASS", "secret")
	t.Setenv("POSTGRES_DB_PORT", "5432")
	pg, _ := database.OpenAdmin()
	defer pg.Close()

	got := pg.Database()

	assert.NotNil(t, got)
}

func TestOpenAdmin_failsWhenAdminIsNotConfigured(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "")
	t.Setenv("POSTGRES_ADMIN_PASS", "")

	want := database.ErrAdminNotConfigured
	_, got := database.OpenAdmin()

	assert.ErrorIs(t, got, want)
}

func TestOpenAdmin_failsWhenPortIsMalformed(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "admin")
	t.Setenv("POSTGRES_ADMIN_PASS", "secret")
	t.Setenv("POSTGRES_DB_PORT", "not-a-port")

	want := database.ErrCantInitDatabase
	_, got := database.OpenAdmin()

	assert.ErrorIs(t, got, want)
}
