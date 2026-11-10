package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/http/response"
)

// RequireToken rejects a request whose bearer token is not token, the secret another service shares with this one to
// call its routes. An empty token rejects every request.
func RequireToken(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			header := r.Header.Get("Authorization")
			if header == "" {
				response.Send(ctx, w, http.StatusUnauthorized, ErrMsgMissingAuthHeader, nil, ErrMissingAuthHeader)
				return
			}

			sent, isBearer := bearerToken(header)
			isToken := isBearer && subtle.ConstantTimeCompare([]byte(sent), []byte(token)) == 1
			if !isToken {
				response.Send(ctx, w, http.StatusUnauthorized, ErrMsgInvalidToken, nil, ErrInvalidToken)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
