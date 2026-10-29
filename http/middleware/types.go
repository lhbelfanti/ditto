package middleware

import (
	"context"
)

type (
	// contextKey is the type of the keys this package stores in a request context.
	contextKey string

	// SelectUserIDByToken is a function that retrieves the user ID associated with a session token.
	SelectUserIDByToken func(ctx context.Context, token string) (int, error)

	// Session identifies the user resolved from a token and whether writes are allowed.
	Session struct {
		UserID   int
		ReadOnly bool
	}

	// GetSessionByToken resolves a bearer token on each request.
	GetSessionByToken func(ctx context.Context, token string) (Session, error)
)
