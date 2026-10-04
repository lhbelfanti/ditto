package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllowedOrigin_success(t *testing.T) {
	allowed := []string{"https://app.example.com", "https://admin.example.com"}

	want := true
	got := isAllowedOrigin("https://admin.example.com", allowed)

	assert.Equal(t, want, got)
}

func TestIsAllowedOrigin_successWhenOriginIsNotListed(t *testing.T) {
	allowed := []string{"https://app.example.com"}

	want := false
	got := isAllowedOrigin("https://evil.example.com", allowed)

	assert.Equal(t, want, got)
}

func TestIsAllowedOrigin_successWhenOriginIsEmpty(t *testing.T) {
	allowed := []string{""}

	want := false
	got := isAllowedOrigin("", allowed)

	assert.Equal(t, want, got)
}
