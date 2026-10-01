package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/lhbelfanti/ditto/v2/http/response"
)

type contextKey string

// UserIDKey is the context key used to store and retrieve the authenticated user ID.
const UserIDKey contextKey = "userID"

// SelectUserIDByToken is a function that retrieves the user ID associated with a session token.
type SelectUserIDByToken func(ctx context.Context, token string) (int, error)

// Auth returns an HTTP middleware that validates Bearer tokens and injects the user ID into the request context.
func Auth(selectUserIDByToken SelectUserIDByToken) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Send(r.Context(), w, http.StatusUnauthorized, "authorization header required", nil, ErrMissingAuthHeader)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.Send(r.Context(), w, http.StatusUnauthorized, ErrMsgInvalidToken, nil, ErrInvalidToken)
				return
			}

			token := parts[1]
			userID, err := selectUserIDByToken(r.Context(), token)
			if err != nil {
				response.Send(r.Context(), w, http.StatusUnauthorized, ErrMsgInvalidToken, nil, ErrInvalidToken)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
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
