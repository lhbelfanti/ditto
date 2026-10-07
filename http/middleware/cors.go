package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/lhbelfanti/ditto/v2/env"
)

// CORS returns an HTTP middleware that sets CORS headers on every response. CORS_ALLOWED_ORIGIN
// may list more than one origin separated by commas — the request's own Origin header is echoed
// back only when it matches one of them, since Access-Control-Allow-Origin accepts exactly one
// value and Access-Control-Allow-Credentials being "true" below rules out a "*" wildcard.
func CORS() func(http.Handler) http.Handler {
	allowed := strings.Split(env.Get("CORS_ALLOWED_ORIGIN", "http://localhost:3000"), ",")
	for i := range allowed {
		allowed[i] = strings.TrimSpace(allowed[i])
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")
			origin := r.Header.Get("Origin")
			if isAllowedOrigin(origin, allowed) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	if origin == "" {
		return false
	}

	return slices.Contains(allowed, origin)
}
