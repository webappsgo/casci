package users

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	// UserContextKey is the key for user in context
	UserContextKey contextKey = "user"
	// ClaimsContextKey is the key for claims in context
	ClaimsContextKey contextKey = "claims"
)

// Middleware provides authentication middleware
type Middleware struct {
	authManager *AuthManager
	service     *Service
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(authManager *AuthManager, service *Service) *Middleware {
	return &Middleware{
		authManager: authManager,
		service:     service,
	}
}

// Authenticate is middleware that validates JWT token or API token
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try JWT token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Bearer token
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := m.authManager.ValidateToken(token)
				if err == nil {
					// Get full user from database
					user, err := m.service.GetByID(r.Context(), claims.UserID)
					if err == nil {
						ctx := context.WithValue(r.Context(), UserContextKey, user)
						ctx = context.WithValue(ctx, ClaimsContextKey, claims)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}
		}

		// Try API token from header
		apiToken := r.Header.Get("X-API-Token")
		if apiToken != "" {
			user, err := m.service.GetByAPIToken(r.Context(), apiToken)
			if err == nil {
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Try API token from query parameter (for webhooks, etc.)
		apiToken = r.URL.Query().Get("token")
		if apiToken != "" {
			user, err := m.service.GetByAPIToken(r.Context(), apiToken)
			if err == nil {
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// No valid authentication found
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// RequireAuth is middleware that requires authentication
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return m.Authenticate(next)
}

// RequireAdmin is middleware that requires admin privileges
func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !user.IsAdmin {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}))
}

// OptionalAuth is middleware that optionally authenticates if token is present
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to authenticate, but don't fail if not authenticated
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := m.authManager.ValidateToken(token)
			if err == nil {
				user, err := m.service.GetByID(r.Context(), claims.UserID)
				if err == nil {
					ctx := context.WithValue(r.Context(), UserContextKey, user)
					ctx = context.WithValue(ctx, ClaimsContextKey, claims)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// Continue without authentication
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves the authenticated user from context
func GetUserFromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(UserContextKey).(*User); ok {
		return user
	}
	return nil
}

// GetClaimsFromContext retrieves the JWT claims from context
func GetClaimsFromContext(ctx context.Context) *Claims {
	if claims, ok := ctx.Value(ClaimsContextKey).(*Claims); ok {
		return claims
	}
	return nil
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	return GetUserFromContext(ctx) != nil
}

// IsAdmin checks if the authenticated user is an admin
func IsAdmin(ctx context.Context) bool {
	user := GetUserFromContext(ctx)
	return user != nil && user.IsAdmin
}
