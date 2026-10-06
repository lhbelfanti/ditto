package app

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
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

// MockMux returns a mux that answers GET pattern with status.
func MockMux(pattern string, status int) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	})
	return mux
}

// MockRoutes returns Routes that answer GET pattern with status.
func MockRoutes(pattern string, status int) Routes {
	return func(mux *http.ServeMux, _ *database.Postgres) {
		mux.HandleFunc(pattern, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		})
	}
}

// MockMiddleware returns a Middleware that adds name to the X-Order response header before calling the next handler.
func MockMiddleware(name string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Order", name)
			next.ServeHTTP(w, r)
		})
	}
}

// MockServe calls h on a GET to path and returns the recorded response.
func MockServe(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// MockRunningService starts Run on a free port with a teapot route and two ordered middleware.
func MockRunningService(t *testing.T) *MockService {
	t.Helper()
	opts := Options{Mux: MockMux("GET /items/v1", http.StatusTeapot)}.
		WithMiddleware(MockMiddleware("outer")).
		WithMiddleware(MockMiddleware("inner"))
	return MockRunningServiceWith(t, opts)
}

// MockRunningServiceWith starts Run with opts, without a database, on a free port. Run installs its signal handler
// before it starts listening, so Stop sending SIGINT to this process exercises the real graceful
// shutdown path.
func MockRunningServiceWith(t *testing.T, opts Options) *MockService {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	t.Setenv("APP_MOCK_PORT", strconv.Itoa(port))
	opts.Name, opts.PortEnv, opts.NoDatabase = "svc", "APP_MOCK_PORT", true

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
