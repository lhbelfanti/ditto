package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestPostgres_MakeCheck_failsWhenDatabaseIsUnreachable(t *testing.T) {
	pg := database.MockPostgres()
	defer pg.Close()

	check := pg.MakeCheck(time.Second)

	want := database.ErrDatabaseUnavailable
	got := check(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestPostgres_Database_success(t *testing.T) {
	pg := database.MockPostgres()
	defer pg.Close()

	got := pg.Database()

	assert.NotNil(t, got)
}

func TestPostgres_Close_success(t *testing.T) {
	pg := database.MockPostgres()
	pool := pg.Database()

	pg.Close()

	got := pool.Ping(context.Background()) != nil

	assert.True(t, got)
}

func TestTargetFromEnv_success(t *testing.T) {
	t.Setenv("POSTGRES_DB_NAME", "svc_db")
	t.Setenv("POSTGRES_DB_USER", "svc")
	t.Setenv("POSTGRES_DB_PASS", "pw")

	want := database.Target{Name: "svc_db", Role: "svc", Pass: "pw"}
	got := database.TargetFromEnv()

	assert.Equal(t, want, got)
}
