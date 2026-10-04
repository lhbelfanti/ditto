package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingHandlerV1_success(t *testing.T) {
	handler := pingHandlerV1()
	req := httptest.NewRequest(http.MethodGet, "/ping/v1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestDatabasePingHandlerV1_success(t *testing.T) {
	mockPing := MockDatabasePing(nil)
	handler := databasePingHandlerV1(mockPing)
	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestDatabasePingHandlerV1_failsWhenPingFails(t *testing.T) {
	mockPing := MockDatabasePing(errors.New("database unreachable"))
	handler := databasePingHandlerV1(mockPing)
	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	want := http.StatusServiceUnavailable
	got := w.Code

	assert.Equal(t, want, got)
}

func TestMigrationsRunHandlerV1_success(t *testing.T) {
	mockRunner := MockMigrationRunner(nil)
	handler := migrationsRunHandlerV1(mockRunner)
	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestMigrationsRunHandlerV1_failsWhenRunnerFails(t *testing.T) {
	mockRunner := MockMigrationRunner(errors.New("migration failed"))
	handler := migrationsRunHandlerV1(mockRunner)
	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	want := http.StatusInternalServerError
	got := w.Code

	assert.Equal(t, want, got)
}
