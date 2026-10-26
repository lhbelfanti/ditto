package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBearerToken_success(t *testing.T) {
	want := "valid-token"
	got, _ := bearerToken("Bearer valid-token")

	assert.Equal(t, want, got)
}

func TestSessionReadMethod_success(t *testing.T) {
	got := sessionReadMethod(http.MethodGet)

	assert.True(t, got)
}
