package server

import (
	"net/http"
	"strings"
)

// CORSMiddleware adds CORS headers based on configuration
func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Get configured CORS origins (default: "*")
		corsOrigins := s.config.Server.CORS
		if corsOrigins == "" {
			corsOrigins = "*"
		}
		
		// Handle CORS
		if corsOrigins == "*" {
			// Allow all origins
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if corsOrigins != "" {
			// Check if origin is in allowed list
			allowed := false
			origins := strings.Split(corsOrigins, ",")
			for _, o := range origins {
				o = strings.TrimSpace(o)
				if o == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
		
		// Common CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-CSRF-Token")
		w.Header().Set("Access-Control-Max-Age", "86400")
		
		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}
