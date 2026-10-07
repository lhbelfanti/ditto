package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/log"
)

const requestIDKey string = "request_id"

// RequestID is an HTTP middleware that honors an inbound X-Request-ID header, or generates one
// if absent, injects it into the request context via log.With, and sets the X-Request-ID response
// header — so a chain of ditto-based services shares one ID across the whole call instead of each
// hop minting its own.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = generateID()
		}
		ctx := log.With(r.Context(), log.BaseParam(requestIDKey, id))
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
