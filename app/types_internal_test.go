package app

import (
	"net/http"
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

func TestOptions_orDefaults_successWhenPortEnvIsSet(t *testing.T) {
	want := Options{PortEnv: "PORT"}
	got := want.orDefaults()

	assert.Equal(t, want, got)
}

func TestOptions_decorate_successWhenWrapRunsOutsideMiddleware(t *testing.T) {
	opts := Options{Wrap: MockMiddleware("wrap")}.WithMiddleware(MockMiddleware("mw"))

	want := []string{"wrap", "mw"}
	got := MockServe(opts.decorate(http.NewServeMux()), "/").Header().Values("X-Order")

	assert.Equal(t, want, got)
}
