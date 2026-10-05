package middleware

import (
	"context"
)

type (
	contextKey string

	// SelectUserIDByToken is a function that retrieves the user ID associated with a session token.
	SelectUserIDByToken func(ctx context.Context, token string) (int, error)
)
