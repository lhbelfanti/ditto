package middleware

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateID_success(t *testing.T) {
	want := regexp.MustCompile(`^[0-9a-f]{32}$`)
	got := generateID()

	assert.Regexp(t, want, got)
}

func TestGenerateID_successWhenCalledTwice(t *testing.T) {
	first := generateID()

	notWant := first
	got := generateID()

	assert.NotEqual(t, notWant, got)
}
