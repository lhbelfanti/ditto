package app

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// RunningService is a Run started in the background by StartService, reachable at BaseURL.
type RunningService struct {
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

// MockMiddleware returns a Middleware that adds name to the X-Order response header before calling the next handler.
func MockMiddleware(name string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Order", name)
			next.ServeHTTP(w, r)
		})
	}
}

// StartService starts Run on a free port with a teapot route and two ordered middleware.
func StartService(t *testing.T) *RunningService {
	t.Helper()
	opts := Options{Mux: MockMux("GET /items/v1", http.StatusTeapot)}.
		WithMiddleware(MockMiddleware("outer")).
		WithMiddleware(MockMiddleware("inner"))
	return StartServiceWith(t, opts)
}

// StartServiceWith starts Run with opts, without a database, on a free port. Run installs its signal handler
// before it starts listening, so Stop sending SIGINT to this process exercises the real graceful
// shutdown path. The free port is released before Run binds it, so another socket can take it in between: Run then
// fails with EADDRINUSE and the service starts again on a new port.
func StartServiceWith(t *testing.T, opts Options) *RunningService {
	t.Helper()
	const attempts int = 3
	opts.Name, opts.PortEnv, opts.NoDatabase = "svc", "APP_MOCK_PORT", true

	for attempt := 1; ; attempt++ {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}

		port := listener.Addr().(*net.TCPAddr).Port
		_ = listener.Close()
		t.Setenv("APP_MOCK_PORT", strconv.Itoa(port))

		// Run mounts its system routes on the mux it gets, so each attempt gets a new one in front of the caller's.
		attemptOpts := opts
		attemptOpts.Mux = http.NewServeMux()
		if opts.Mux != nil {
			attemptOpts.Mux.Handle("/", opts.Mux)
		}

		service := &RunningService{
			BaseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
			done:    make(chan error, 1),
			client:  &http.Client{Transport: &http.Transport{DisableKeepAlives: true}},
		}
		go func() { service.done <- Run(attemptOpts) }()
		t.Cleanup(func() { _ = service.Stop() })

		runErr := service.waitUntilStarted(t)
		if runErr == nil {
			return service
		}

		if !errors.Is(runErr, syscall.EADDRINUSE) || attempt == attempts {
			t.Fatalf("service at %s stopped before it started: %v", service.BaseURL, runErr)
		}
	}
}

// waitUntilStarted polls the service until it answers, and returns Run's error if Run returns first.
func (m *RunningService) waitUntilStarted(t *testing.T) error {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for {
		resp, err := m.client.Get(m.BaseURL + "/ping/v1")
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}

		select {
		case runErr := <-m.done:
			// Run already returned, so its signal handler is gone and a SIGINT from Stop would kill the test binary.
			m.stopped = true
			return runErr
		case <-deadline:
			t.Fatalf("service at %s did not start", m.BaseURL)
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// Get requests path on the service and returns the response with its body drained, or an empty response if the request fails.
func (m *RunningService) Get(path string) *http.Response {
	resp, err := m.client.Get(m.BaseURL + path)
	if err != nil {
		return &http.Response{}
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp
}

// Stop sends SIGINT to this process and returns the error Run returned after its graceful shutdown. Calling it again returns nil.
func (m *RunningService) Stop() error {
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
