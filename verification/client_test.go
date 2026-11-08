package verification_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/verification"
)

func TestNewClient_success(t *testing.T) {
	mockServer := verification.MockRedirectServer()
	defer mockServer.Close()

	client := verification.NewClient(time.Second)

	response, _ := client.NewRequest(http.MethodGet, mockServer.URL).Call(context.Background())

	want := http.StatusFound
	got := response.StatusCode

	assert.Equal(t, want, got)
}
