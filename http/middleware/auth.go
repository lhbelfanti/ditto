package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/http/response"
)

// UserIDKey is the context key used to store and retrieve the authenticated user ID.
const UserIDKey contextKey = "userID"

// Auth returns an HTTP middleware that validates Bearer tokens and injects the user ID into the request context.
func Auth(selectUserIDByToken SelectUserIDByToken) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Send(ctx, w, http.StatusUnauthorized, ErrMsgMissingAuthHeader, nil, ErrMissingAuthHeader)
				return
			}

			token, ok := bearerToken(authHeader)
			if !ok {
				response.Send(ctx, w, http.StatusUnauthorized, ErrMsgInvalidToken, nil, ErrInvalidToken)
				return
			}

			userID, err := selectUserIDByToken(ctx, token)
			if err != nil {
				response.Send(ctx, w, http.StatusUnauthorized, ErrMsgInvalidToken, nil, fmt.Errorf("%w: %w", ErrInvalidToken, err))
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithUserID(ctx, userID)))
		})
	}
}

// UserIDFromContext retrieves the user ID from the context. Returns 0 if not set.
func UserIDFromContext(ctx context.Context) int {
	id, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0
	}

	return id
}

// ContextWithUserID returns a new context with the given user ID stored under UserIDKey.
func ContextWithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
