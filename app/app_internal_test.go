package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeouts_withDefaults_success(t *testing.T) {
	want := Timeouts{Startup: defaultStartupTimeout, Ping: defaultPingTimeout, Shutdown: defaultShutdownTimeout}
	got := Timeouts{}.withDefaults()

	assert.Equal(t, want, got)
}

func TestTimeouts_withDefaults_successWhenValuesAreSet(t *testing.T) {
	want := Timeouts{Startup: time.Second, Ping: 3 * time.Second, Shutdown: 4 * time.Second}
	got := want.withDefaults()

	assert.Equal(t, want, got)
}
