package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeouts_orDefaults_success(t *testing.T) {
	want := Timeouts{Startup: defaultStartupTimeout, Ping: defaultPingTimeout, Shutdown: defaultShutdownTimeout}
	got := Timeouts{}.orDefaults()

	assert.Equal(t, want, got)
}

func TestTimeouts_orDefaults_successWhenValuesAreSet(t *testing.T) {
	want := Timeouts{Startup: time.Second, Ping: 3 * time.Second, Shutdown: 4 * time.Second}
	got := want.orDefaults()

	assert.Equal(t, want, got)
}

func TestOptions_orDefaults_success(t *testing.T) {
	want := Options{PortEnv: defaultPortEnv}
	got := Options{}.orDefaults()

	assert.Equal(t, want, got)
}

func TestOptions_orDefaults_successWhenValuesAreSet(t *testing.T) {
	want := Options{PortEnv: "PORT"}
	got := want.orDefaults()

	assert.Equal(t, want, got)
}

func TestOptions_decorate_success(t *testing.T) {
	mockFirst := MockMiddleware("first")
	mockSecond := MockMiddleware("second")
	opts := Options{}.WithMiddleware(mockFirst).WithMiddleware(mockSecond)
	handler := opts.decorate(http.NewServeMux())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{"first", "second"}
	got := w.Header().Values("X-Order")

	assert.Equal(t, want, got)
}
