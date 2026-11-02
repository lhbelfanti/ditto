package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/log"
)

// MockSelectUserIDByToken returns a SelectUserIDByToken that always returns userID and err.
func MockSelectUserIDByToken(userID int, err error) SelectUserIDByToken {
	return func(context.Context, string) (int, error) { return userID, err }
}

// MockGetSessionByToken returns a GetSessionByToken that always returns session and err.
func MockGetSessionByToken(session Session, err error) GetSessionByToken {
	return func(context.Context, string) (Session, error) { return session, err }
}

// MockGetSessionByTokenExpecting returns a GetSessionByToken that returns session when it receives
// token, and an error otherwise.
func MockGetSessionByTokenExpecting(token string, session Session) GetSessionByToken {
	return func(_ context.Context, received string) (Session, error) {
		if received != token {
			return Session{}, fmt.Errorf("unexpected token: %q", received)
		}

		return session, nil
	}
}

// MockOKHandler returns a handler that answers 200.
func MockOKHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// MockNoopHandler returns a handler that does nothing.
func MockNoopHandler() http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
}

// MockUserIDHandler returns a handler that answers 200 and stores the user ID from the request context in gotUserID.
func MockUserIDHandler(gotUserID *int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
}

// MockUserIDResponseHandler returns a handler that answers 200 with the user ID from the request context as its body.
func MockUserIDResponseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, UserIDFromContext(r.Context()))
	})
}

// MockRequestIDLoggingHandler returns a handler that logs a message with the request context.
func MockRequestIDLoggingHandler() http.Handler {
	return http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		log.Info(r.Context(), "test message")
	})
}

// MockCountingHandler returns a handler that counts its calls in *calls and answers 200.
func MockCountingHandler(calls *int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*calls++
		w.WriteHeader(http.StatusOK)
	})
}

// MockCORSHeaders returns the headers CORS sets on a response, with origin as the allowed origin, or without one when origin is empty.
func MockCORSHeaders(origin string) http.Header {
	headers := http.Header{
		"Vary":                             {"Origin"},
		"Access-Control-Allow-Methods":     {"GET, POST, PUT, DELETE, OPTIONS"},
		"Access-Control-Allow-Headers":     {"Authorization, Content-Type"},
		"Access-Control-Allow-Credentials": {"true"},
		"Access-Control-Max-Age":           {"86400"},
	}
	if origin != "" {
		headers.Set("Access-Control-Allow-Origin", origin)
	}

	return headers
}
