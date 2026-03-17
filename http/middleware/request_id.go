package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/lhbelfanti/ditto/log"
)

const requestIDKey = "request_id"

// RequestID is an HTTP middleware that generates a unique request ID,
// injects it into the request context via log.With, and sets the
// X-Request-ID response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := generateID()
		ctx := log.With(r.Context(), log.Param(requestIDKey, id))
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
