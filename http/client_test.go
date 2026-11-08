package http_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestNewClient_success(t *testing.T) {
	want := &_http.CustomClient{HTTPClient: &http.Client{Timeout: 5 * time.Second}}
	got := _http.NewClient(5 * time.Second)

	assert.Equal(t, want, got)
}
