package server

import (
	"net/http"
	"strings"
)

// CORSMiddleware adds CORS headers based on configuration.
// The secure default is same-origin only: with no configured allow-list no
// Access-Control-Allow-Origin header is emitted at all, so browsers refuse
// cross-origin reads. The "*" wildcard is reachable only when an operator
// configures it explicitly.
func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// The response varies per Origin, so caches must key on it
		w.Header().Add("Vary", "Origin")

		corsOrigins := strings.TrimSpace(s.config.Server.CORS)

		switch {
		case corsOrigins == "":
			// Same-origin only - no cross-origin access granted

		case corsOrigins == "*":
			// Explicitly configured wildcard. Credentials are never sent with
			// "*" - the combination is rejected by browsers anyway
			w.Header().Set("Access-Control-Allow-Origin", "*")

		default:
			// Check if origin is in the allowed list
			for _, o := range strings.Split(corsOrigins, ",") {
				if strings.TrimSpace(o) == origin && origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					break
				}
			}
		}

		// Common CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-CSRF-Token")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
