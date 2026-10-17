package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/env"
)

func TestMakeCheck_success(t *testing.T) {
	mockPing := database.MockPing(nil)

	check := database.MakeCheck(mockPing, time.Second)

	got := check(context.Background())

	assert.NoError(t, got)
}

func TestMakeCheck_failsWhenPingFails(t *testing.T) {
	mockErr := errors.New("ping failed")
	mockPing := database.MockPing(mockErr)

	check := database.MakeCheck(mockPing, time.Second)

	want := database.ErrDatabaseUnavailable
	got := check(context.Background())
	wantCause := mockErr

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestMakeCheck_failsWhenPingExceedsTimeout(t *testing.T) {
	mockPing := database.MockBlockingPing(nil)

	check := database.MakeCheck(mockPing, 10*time.Millisecond)

	want := database.ErrDatabaseUnavailable
	got := check(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeCheck_failsWhenPingExceedsTimeoutReturnsPromptly(t *testing.T) {
	mockPing := database.MockBlockingPing(nil)

	check := database.MakeCheck(mockPing, 10*time.Millisecond)

	start := time.Now()
	_ = check(context.Background())
	got := time.Since(start)

	assert.Less(t, got, time.Second)
}

func TestMakeCheck_failsWhenPingErrorCarriesCredentials(t *testing.T) {
	credentialBearingDSN := errors.New("dial postgres://nebula_earth_user:super-secret@db:5432/nebula_earth failed")
	mockPing := database.MockPing(credentialBearingDSN)

	check := database.MakeCheck(mockPing, time.Second)

	want := database.ErrDatabaseUnavailable.Error()
	got := check(context.Background()).Error()

	assert.Equal(t, want, got)
}

func TestRequireEnv_success(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{
		"POSTGRES_DB_PORT": "5432",
		"POSTGRES_DB_NAME": "nebula_earth",
		"POSTGRES_DB_USER": "nebula_earth_user",
		"POSTGRES_DB_PASS": "correct-horse-battery-staple",
	})

	got := database.RequireEnv(mockLookup)

	assert.NoError(t, got)
}

func TestRequireEnv_failsWhenRequiredKeyIsMissing(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "Port", key: "POSTGRES_DB_PORT"},
		{name: "Name", key: "POSTGRES_DB_NAME"},
		{name: "User", key: "POSTGRES_DB_USER"},
		{name: "Pass", key: "POSTGRES_DB_PASS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{
				"POSTGRES_DB_PORT": "5432",
				"POSTGRES_DB_NAME": "nebula_earth",
				"POSTGRES_DB_USER": "nebula_earth_user",
				"POSTGRES_DB_PASS": "correct-horse-battery-staple",
			}
			delete(values, tt.key)
			mockLookup := env.MockLookup(values)

			got := database.RequireEnv(mockLookup)

			assert.Error(t, got)
		})
	}
}

func TestRequireEnv_failsWhenPortIsMalformed(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{
		"POSTGRES_DB_PORT": "not-a-port",
		"POSTGRES_DB_NAME": "nebula_earth",
		"POSTGRES_DB_USER": "nebula_earth_user",
		"POSTGRES_DB_PASS": "correct-horse-battery-staple",
	})

	got := database.RequireEnv(mockLookup)

	assert.Error(t, got)
}
