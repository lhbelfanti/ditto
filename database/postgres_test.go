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
