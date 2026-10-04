package app

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/env"
)

func TestRun_failsWhenPortIsMissing(t *testing.T) {
	err := Run(Options{Name: "svc", PortEnv: "APP_TEST_MISSING_PORT"})

	var keyErr *env.KeyError
	require.ErrorAs(t, err, &keyErr)
}

func TestRun_failsWhenDatabaseEnvIsIncomplete(t *testing.T) {
	t.Setenv("APP_TEST_PORT", "4000")
	t.Setenv("POSTGRES_DB_PORT", "")

	err := Run(Options{Name: "svc", PortEnv: "APP_TEST_PORT", MigrationsDir: "./migrations"})

	assert.Error(t, err)
}

// Run installs its signal handler before it starts listening, so once the liveness route answers,
// sending SIGINT to this process exercises the real graceful shutdown path.
func TestRun_successWhenServerStartsAndShutsDown(t *testing.T) {
	port := freePort(t)
	t.Setenv("APP_TEST_PORT", strconv.Itoa(port))
	opts := Options{
		Name:    "svc",
		PortEnv: "APP_TEST_PORT",
		Routes: func(mux *http.ServeMux, _ *database.Postgres) {
			mux.HandleFunc("GET /items/v1", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})
		},
		Wrap: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Wrapped", "true")
				next.ServeHTTP(w, r)
			})
		},
	}
	done := make(chan error, 1)
	go func() { done <- Run(opts) }()
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	waitForServer(t, baseURL+"/ping/v1")

	ping := get(t, baseURL+"/ping/v1")
	items := get(t, baseURL+"/items/v1")
	databasePing := get(t, baseURL+"/database/ping/v1")

	assert.Equal(t, http.StatusOK, ping.StatusCode)
	assert.Equal(t, http.StatusTeapot, items.StatusCode)
	assert.Equal(t, "true", items.Header.Get("X-Wrapped"))
	assert.Equal(t, http.StatusNotFound, databasePing.StatusCode)

	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not shut down after SIGINT")
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = listener.Close() }()
	return listener.Addr().(*net.TCPAddr).Port
}

// The client never reuses connections, so no idle or unread connection can keep Shutdown waiting.
var httpClient = &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}

func waitForServer(t *testing.T, url string) {
	t.Helper()
	for i := 0; i < 100; i++ {
		resp, err := httpClient.Get(url)
		if err == nil {
			drainAndClose(resp)
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server at %s did not start", url)
}

func get(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := httpClient.Get(url)
	require.NoError(t, err)
	drainAndClose(resp)
	return resp
}

func drainAndClose(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
