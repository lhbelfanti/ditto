package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

// InitPostgres backs a package-level singleton (sync.Once) — this is the only test exercising it,
// since any other test calling it first would decide the singleton's outcome for the rest of this
// package's test binary. A malformed port fails pgxpool.New's own config parsing immediately, with
// no real network I/O needed to exercise the failure path.
func TestInitPostgres_returnsSameErrorOnRetryAfterFailure(t *testing.T) {
	t.Setenv("POSTGRES_DB_USER", "u")
	t.Setenv("POSTGRES_DB_PASS", "p")
	t.Setenv("POSTGRES_DB_NAME", "d")
	t.Setenv("POSTGRES_DB_PORT", "not-a-port")

	firstInstance, firstErr := database.InitPostgres()
	secondInstance, secondErr := database.InitPostgres()

	assert.ErrorIs(t, firstErr, database.ErrCantInitDatabase)
	assert.ErrorIs(t, secondErr, database.ErrCantInitDatabase)
	assert.Equal(t, firstInstance, secondInstance)
}
