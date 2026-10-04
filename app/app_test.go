package app_test

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

	"github.com/lhbelfanti/ditto/v2/app"
	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/env"
)

type runOutcome struct {
	pingCode         int
	itemsCode        int
	databasePingCode int
	wrapped          string
	shutdownErr      error
}

func TestRun_failsWhenPortIsMissing(t *testing.T) {
	got := app.Run(app.Options{Name: "svc", PortEnv: "APP_TEST_MISSING_PORT"})

	var keyErr *env.KeyError
	assert.ErrorAs(t, got, &keyErr)
}

func TestRun_failsWhenDatabaseEnvIsIncomplete(t *testing.T) {
	t.Setenv("APP_TEST_PORT", "4000")
	t.Setenv("POSTGRES_DB_PORT", "")

	got := app.Run(app.Options{Name: "svc", PortEnv: "APP_TEST_PORT", MigrationsDir: "./migrations"})

	assert.Error(t, got)
}

// Run installs its signal handler before it starts listening, so once the liveness route answers,
// sending SIGINT to this process exercises the real graceful shutdown path.
func TestRun_successWhenServerStartsAndShutsDown(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	t.Setenv("APP_TEST_PORT", strconv.Itoa(port))

	opts := app.Options{
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
	go func() { done <- app.Run(opts) }()

	// The client never reuses connections, so no idle or unread connection can keep Shutdown waiting.
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	fetch := func(path string) (*http.Response, error) {
		resp, err := client.Get(baseURL + path)
		if err != nil {
			return nil, err
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return resp, nil
	}

	for i := 0; i < 100; i++ {
		_, err = fetch("/ping/v1")
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	require.NoError(t, err)

	ping, err := fetch("/ping/v1")
	require.NoError(t, err)
	items, err := fetch("/items/v1")
	require.NoError(t, err)
	databasePing, err := fetch("/database/ping/v1")
	require.NoError(t, err)

	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
	var shutdownErr error
	select {
	case shutdownErr = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not shut down after SIGINT")
	}

	want := runOutcome{
		pingCode:         http.StatusOK,
		itemsCode:        http.StatusTeapot,
		databasePingCode: http.StatusNotFound,
		wrapped:          "true",
		shutdownErr:      nil,
	}
	got := runOutcome{
		pingCode:         ping.StatusCode,
		itemsCode:        items.StatusCode,
		databasePingCode: databasePing.StatusCode,
		wrapped:          items.Header.Get("X-Wrapped"),
		shutdownErr:      shutdownErr,
	}

	assert.Equal(t, want, got)
}
