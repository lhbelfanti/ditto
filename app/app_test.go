package app_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/app"
	"github.com/lhbelfanti/ditto/v2/env"
)

func TestRun_failsWhenPortIsMissing(t *testing.T) {
	want := app.ErrInvalidPort
	wantCause := env.ErrMissingKey
	got := app.Run(app.Options{Name: "svc", PortEnv: "APP_TEST_MISSING_PORT"})

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestRun_failsWhenDatabaseEnvIsIncomplete(t *testing.T) {
	t.Setenv("APP_TEST_PORT", "4000")
	t.Setenv("POSTGRES_DB_PORT", "")

	want := app.ErrCantInitDatabase
	wantCause := app.ErrInvalidDatabaseEnv
	got := app.Run(app.Options{Name: "svc", PortEnv: "APP_TEST_PORT"})

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestRun_successWhenServerStartsAndShutsDown(t *testing.T) {
	service := app.StartService(t)

	got := service.Stop()

	assert.NoError(t, got)
}

func TestRun_successWhenLivenessRouteIsMounted(t *testing.T) {
	service := app.StartService(t)

	want := http.StatusOK
	got := service.Get("/ping/v1").StatusCode

	assert.Equal(t, want, got)
}

func TestRun_successWhenServiceRoutesAreMounted(t *testing.T) {
	service := app.StartService(t)

	want := http.StatusTeapot
	got := service.Get("/items/v1").StatusCode

	assert.Equal(t, want, got)
}

func TestRun_successWhenTwoMiddlewareAreAdded(t *testing.T) {
	service := app.StartService(t)

	want := []string{"outer", "inner"}
	got := service.Get("/items/v1").Header.Values("X-Order")

	assert.Equal(t, want, got)
}

func TestRun_successWhenDatabaseIsDisabled(t *testing.T) {
	service := app.StartService(t)

	want := http.StatusNotFound
	got := service.Get("/database/ping/v1").StatusCode

	assert.Equal(t, want, got)
}

func TestInitDatabase_failsWhenEnvIsIncomplete(t *testing.T) {
	t.Setenv("POSTGRES_DB_PORT", "")

	want := app.ErrInvalidDatabaseEnv
	_, got := app.InitDatabase(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestInitDatabaseWithMigrationFolder_failsWhenEnvIsIncomplete(t *testing.T) {
	t.Setenv("POSTGRES_DB_PORT", "")

	want := app.ErrInvalidDatabaseEnv
	_, got := app.InitDatabaseWithMigrationFolder(context.Background(), "./sql")

	assert.ErrorIs(t, got, want)
}
