package middleware

import (
	"context"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/log"
)

// MockSelectUserIDByToken returns a SelectUserIDByToken that always returns userID and err.
func MockSelectUserIDByToken(userID int, err error) SelectUserIDByToken {
	return func(context.Context, string) (int, error) { return userID, err }
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

// MockUserIDHandler returns a handler that stores the user ID from the request context in *gotUserID and answers 200.
func MockUserIDHandler(gotUserID *int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
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
