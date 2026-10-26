package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lhbelfanti/ditto/v2/http/response"
	"github.com/lhbelfanti/ditto/v2/log"
)

// SessionAuth validates a bearer token on each request, exposes its user ID, and blocks writes for read-only sessions.
func SessionAuth(getSession GetSessionByToken) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			header := r.Header.Get("Authorization")
			if header == "" {
				response.Send(ctx, w, http.StatusUnauthorized, ErrMsgMissingAuthHeader, nil, ErrMissingAuthHeader)
				return
			}

			token, ok := bearerToken(header)
			if !ok {
				response.Send(ctx, w, http.StatusUnauthorized, ErrMsgInvalidSession, nil, ErrInvalidSession)
				return
			}

			session, err := getSession(ctx, token)
			if err != nil {
				response.Send(ctx, w, http.StatusUnauthorized, ErrMsgInvalidSession, nil, fmt.Errorf("%w: %w", ErrInvalidSession, err))
				return
			}

			if session.UserID <= 0 {
				response.Send(ctx, w, http.StatusUnauthorized, ErrMsgInvalidSession, nil, ErrInvalidSession)
				return
			}

			ctx = ContextWithUserID(ctx, session.UserID)
			ctx = log.With(ctx, log.Param("user_id", session.UserID))

			if session.ReadOnly && !sessionReadMethod(r.Method) {
				response.Send(ctx, w, http.StatusForbidden, ErrMsgReadOnlySession, nil, ErrReadOnlySession)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(header string) (string, bool) {
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || token == "" || strings.ContainsAny(token, " \t\r\n") {
		return "", false
	}

	return token, true
}

func sessionReadMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
