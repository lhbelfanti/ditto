package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllowedOrigin_success(t *testing.T) {
	allowed := []string{"https://app.example.com", "https://admin.example.com"}

	got := isAllowedOrigin("https://admin.example.com", allowed)

	assert.True(t, got)
}

func TestIsAllowedOrigin_successWhenOriginIsNotListed(t *testing.T) {
	allowed := []string{"https://app.example.com"}

	got := isAllowedOrigin("https://evil.example.com", allowed)

	assert.False(t, got)
}

func TestIsAllowedOrigin_successWhenOriginIsEmpty(t *testing.T) {
	allowed := []string{""}

	got := isAllowedOrigin("", allowed)

	assert.False(t, got)
}
