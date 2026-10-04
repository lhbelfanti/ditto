package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestInit_failsWhenEnvIsIncomplete(t *testing.T) {
	t.Setenv("POSTGRES_DB_PORT", "")

	_, got := database.Init(time.Second)

	assert.Error(t, got)
}
