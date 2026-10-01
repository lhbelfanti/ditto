package middleware

import (
	"context"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/log"
)

func MockSelectUserIDByToken(userID int, err error) SelectUserIDByToken {
	return func(context.Context, string) (int, error) { return userID, err }
}

func MockOKHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func MockNoopHandler() http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
}

func MockNextHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		*called = true
	})
}

func MockUserIDHandler(gotUserID *int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
}

func MockRequestIDLoggingHandler() http.Handler {
	return http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		log.Info(r.Context(), "test message")
	})
}
