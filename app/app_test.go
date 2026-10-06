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
	got := app.Run(app.Options{Name: "svc", PortEnv: "APP_TEST_MISSING_PORT"})

	var keyErr *env.KeyError
	assert.ErrorAs(t, got, &keyErr)
}

func TestRun_failsWhenDatabaseEnvIsIncomplete(t *testing.T) {
	t.Setenv("APP_TEST_PORT", "4000")
	t.Setenv("POSTGRES_DB_PORT", "")

	got := app.Run(app.Options{Name: "svc", PortEnv: "APP_TEST_PORT"})

	assert.Error(t, got)
}

func TestRun_successWhenServerStartsAndShutsDown(t *testing.T) {
	service := app.MockRunningService(t)

	got := service.Stop()

	assert.NoError(t, got)
}

func TestRun_successWhenLivenessRouteIsMounted(t *testing.T) {
	service := app.MockRunningService(t)

	want := http.StatusOK
	got := service.Get("/ping/v1").StatusCode

	assert.Equal(t, want, got)
}

func TestRun_successWhenServiceRoutesAreMounted(t *testing.T) {
	service := app.MockRunningService(t)

	want := http.StatusTeapot
	got := service.Get("/items/v1").StatusCode

	assert.Equal(t, want, got)
}

func TestRun_successWhenMiddlewareRunsOutermostFirst(t *testing.T) {
	service := app.MockRunningService(t)

	want := []string{"outer", "inner"}
	got := service.Get("/items/v1").Header.Values("X-Order")

	assert.Equal(t, want, got)
}

func TestRun_successWhenDatabaseIsDisabled(t *testing.T) {
	service := app.MockRunningService(t)

	want := http.StatusNotFound
	got := service.Get("/database/ping/v1").StatusCode

	assert.Equal(t, want, got)
}

func TestRun_successWhenDeprecatedRoutesAreMounted(t *testing.T) {
	service := app.MockRunningServiceWith(t, app.Options{Routes: app.MockRoutes("GET /legacy/v1", http.StatusAccepted)})

	want := http.StatusAccepted
	got := service.Get("/legacy/v1").StatusCode

	assert.Equal(t, want, got)
}

func TestInitDatabase_failsWhenEnvIsIncomplete(t *testing.T) {
	t.Setenv("POSTGRES_DB_PORT", "")

	_, got := app.InitDatabase(context.Background(), "./migrations")

	assert.Error(t, got)
}
