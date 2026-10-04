package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	dittohttp "github.com/lhbelfanti/ditto/v2/http"
)

func TestCustomClient_NewRequest_successWhenBodyIsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, `{"symbol":"BTC/USDT"}`, readAll(t, r.Body))
		w.Header().Set("X-Test-Header", "present")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := dittohttp.NewClient(5 * time.Second)
	resp, err := client.NewRequest(context.Background(), http.MethodPost, server.URL, map[string]string{"symbol": "BTC/USDT"})

	assert.NoError(t, err)
	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, `{"ok":true}`, resp.Body)
	assert.Equal(t, "present", resp.Header.Get("X-Test-Header"))
}

func TestCustomClient_NewRequest_successWhenBodyIsNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "", r.Header.Get("Content-Type"))
		assert.Equal(t, "", readAll(t, r.Body))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := dittohttp.NewClient(5 * time.Second)
	resp, err := client.NewRequest(context.Background(), http.MethodGet, server.URL+"?symbol=BTCUSDT", nil)

	assert.NoError(t, err)
	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCustomClient_NewRequest_successWhenResponseHasHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-MBX-USED-WEIGHT-1M", "42")
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := dittohttp.NewClient(5 * time.Second)
	resp, err := client.NewRequest(context.Background(), http.MethodGet, server.URL, nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, "42", resp.Header.Get("X-MBX-USED-WEIGHT-1M"))
	assert.Equal(t, "5", resp.Header.Get("Retry-After"))
}

func TestCustomClient_NewRequest_failsWhenURLIsInvalid(t *testing.T) {
	client := dittohttp.NewClient(5 * time.Second)
	resp, err := client.NewRequest(context.Background(), http.MethodGet, "://invalid-url", nil)

	assert.ErrorIs(t, err, dittohttp.FailedToCreateRequest)
	assert.Equal(t, dittohttp.Response{}, resp)
}

func TestCustomClient_NewRequest_failsWhenContextIsCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := dittohttp.NewClient(5 * time.Second)
	resp, err := client.NewRequest(ctx, http.MethodGet, server.URL, nil)

	assert.ErrorIs(t, err, dittohttp.FailedToExecuteRequest)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, dittohttp.Response{}, resp)
}

func readAll(t *testing.T, r interface{ Read([]byte) (int, error) }) string {
	t.Helper()
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

func TestSystemRoutes_WithMigrationRunner_success(t *testing.T) {
	mockRunner := dittohttp.MockMigrationRunner(nil)
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux).WithMigrationRunner(mockRunner)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestSystemRoutes_WithMigrationRunner_failsWhenRunnerFails(t *testing.T) {
	mockRunner := dittohttp.MockMigrationRunner(errors.New("migration failed"))
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux).WithMigrationRunner(mockRunner)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusInternalServerError
	got := w.Code

	assert.Equal(t, want, got)
}

func TestSystemRoutes_WithMigrationRunner_failsWhenRunnerFailsWithSensitiveError(t *testing.T) {
	underlyingErrText := "pq: relation \"trades\" already exists"
	mockRunner := dittohttp.MockMigrationRunner(errors.New(underlyingErrText))
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux).WithMigrationRunner(mockRunner)

	req := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	got := w.Body.String()

	assert.NotContains(t, got, underlyingErrText)
}

func TestSystemRoutes_WithDatabasePing_success(t *testing.T) {
	mockPing := dittohttp.MockDatabasePing(nil)
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusOK
	got := w.Code

	assert.Equal(t, want, got)
}

func TestSystemRoutes_WithDatabasePing_failsWhenPingFails(t *testing.T) {
	mockPing := dittohttp.MockDatabasePing(errors.New("database unreachable"))
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := http.StatusServiceUnavailable
	got := w.Code

	assert.Equal(t, want, got)
}

func TestSystemRoutes_WithDatabasePing_failsWhenPingFailsWithSensitiveError(t *testing.T) {
	underlyingErrText := "pgx: connection refused on host db-internal:5432"
	mockPing := dittohttp.MockDatabasePing(errors.New(underlyingErrText))
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux).WithDatabasePing(mockPing)

	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	got := w.Body.String()

	assert.NotContains(t, got, underlyingErrText)
}

func TestSystemRoutes_WithDatabasePing_successWhenChainedAfterMigrationRunner(t *testing.T) {
	mockRunner := dittohttp.MockMigrationRunner(nil)
	mockPing := dittohttp.MockDatabasePing(nil)
	mux := http.NewServeMux()
	dittohttp.RegisterSystemRoutes(mux).WithMigrationRunner(mockRunner).WithDatabasePing(mockPing)

	migrationsReq := httptest.NewRequest(http.MethodPost, "/migrations/run/v1", nil)
	migrationsW := httptest.NewRecorder()
	mux.ServeHTTP(migrationsW, migrationsReq)

	pingReq := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	pingW := httptest.NewRecorder()
	mux.ServeHTTP(pingW, pingReq)

	assert.Equal(t, http.StatusOK, migrationsW.Code)
	assert.Equal(t, http.StatusOK, pingW.Code)
}
