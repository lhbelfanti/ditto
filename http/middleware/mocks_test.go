package middleware_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/middleware"
)

func TestMockGetSessionByTokenExpecting_success(t *testing.T) {
	mockGetSession := middleware.MockGetSessionByTokenExpecting("valid-token", middleware.Session{UserID: 42})

	want := middleware.Session{UserID: 42}
	got, _ := mockGetSession(context.Background(), "valid-token")

	assert.Equal(t, want, got)
}

func TestMockGetSessionByTokenExpecting_failsWhenTokenIsUnexpected(t *testing.T) {
	mockGetSession := middleware.MockGetSessionByTokenExpecting("valid-token", middleware.Session{UserID: 42})

	_, got := mockGetSession(context.Background(), "other-token")

	assert.Error(t, got)
}
