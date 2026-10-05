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

	"github.com/lhbelfanti/ditto/v2/database"
)

// MockService is a Run started in the background by MockRunningService, reachable at BaseURL.
type MockService struct {
	BaseURL string
	done    chan error
	client  *http.Client
	stopped bool
}

// MockRunningService starts Run on a free port. Run installs its signal handler before it starts
// listening, so Stop sending SIGINT to this process exercises the real graceful shutdown path.
func MockRunningService(t *testing.T) *MockService {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	t.Setenv("APP_MOCK_PORT", strconv.Itoa(port))

	opts := Options{
		Name:    "svc",
		PortEnv: "APP_MOCK_PORT",
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
	service := &MockService{
		BaseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
		done:    make(chan error, 1),
		client:  &http.Client{Transport: &http.Transport{DisableKeepAlives: true}},
	}
	go func() { service.done <- Run(opts) }()
	t.Cleanup(func() { _ = service.Stop() })

	for range 100 {
		resp, err := service.client.Get(service.BaseURL + "/ping/v1")
		if err == nil {
			_ = resp.Body.Close()
			return service
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("service at %s did not start", service.BaseURL)
	return service
}

// Get requests path on the service and returns the response with its body drained, or an empty response if the request fails.
func (m *MockService) Get(path string) *http.Response {
	resp, err := m.client.Get(m.BaseURL + path)
	if err != nil {
		return &http.Response{}
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp
}

// Stop sends SIGINT to this process and returns the error Run returned after its graceful shutdown. Calling it again returns nil.
func (m *MockService) Stop() error {
	if m.stopped {
		return nil
	}
	m.stopped = true
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	select {
	case err := <-m.done:
		return err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("service did not shut down after SIGINT")
	}
}
