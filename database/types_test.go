package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestPostgres_MakeCheck_failsWhenDatabaseIsUnreachable(t *testing.T) {
	mockPostgres := database.MockPostgres()
	defer mockPostgres.Close()

	check := mockPostgres.MakeCheck(time.Second)

	want := database.ErrDatabaseUnavailable
	got := check(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestPostgres_Database_success(t *testing.T) {
	mockPostgres := database.MockPostgres()
	defer mockPostgres.Close()

	got := mockPostgres.Database()

	assert.NotNil(t, got)
}

func TestPostgres_Close_success(t *testing.T) {
	mockPostgres := database.MockPostgres()
	pool := mockPostgres.Database()

	mockPostgres.Close()

	got := pool.Ping(context.Background()) != nil

	assert.True(t, got)
}
