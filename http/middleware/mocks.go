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

// MockNextHandler returns a handler that sets *called to true.
func MockNextHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		*called = true
	})
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
