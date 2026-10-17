package http_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http"
)

func TestNewClient_success(t *testing.T) {
	want := 5 * time.Second
	client := http.NewClient(want)

	got := client.HTTPClient.Timeout

	assert.Equal(t, want, got)
}
