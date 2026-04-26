package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

// CSRFToken represents a CSRF token with expiration
type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
}

// CSRFManager manages CSRF tokens
type CSRFManager struct {
	tokens map[string]*CSRFToken
	mu     sync.RWMutex
}

// NewCSRFManager creates a new CSRF manager
func NewCSRFManager() *CSRFManager {
	cm := &CSRFManager{
		tokens: make(map[string]*CSRFToken),
	}
	
	// Start cleanup goroutine
	go cm.cleanup()
	
	return cm
}

// GenerateToken generates a new CSRF token
func (cm *CSRFManager) GenerateToken() (string, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	
	token := base64.URLEncoding.EncodeToString(b)
	
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	// Store token with 24h expiration
	cm.tokens[token] = &CSRFToken{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	
	return token, nil
}

// ValidateToken validates a CSRF token
func (cm *CSRFManager) ValidateToken(token string) bool {
	if token == "" {
		return false
	}
	
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	t, exists := cm.tokens[token]
	if !exists {
		return false
	}
	
	// Check if expired
	if time.Now().After(t.ExpiresAt) {
		return false
	}
	
	return true
}

// InvalidateToken removes a token (used after validation)
func (cm *CSRFManager) InvalidateToken(token string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.tokens, token)
}

// cleanup removes expired tokens every hour
func (cm *CSRFManager) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		cm.mu.Lock()
		now := time.Now()
		for token, t := range cm.tokens {
			if now.After(t.ExpiresAt) {
				delete(cm.tokens, token)
			}
		}
		cm.mu.Unlock()
	}
}

// CSRFMiddleware provides CSRF protection
func (s *Server) CSRFMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for GET, HEAD, OPTIONS
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		
		// Get token from form or header
		token := r.FormValue("csrf_token")
		if token == "" {
			token = r.Header.Get("X-CSRF-Token")
		}
		
		// Validate token
		if !s.csrfManager.ValidateToken(token) {
			http.Error(w, "Invalid or missing CSRF token", http.StatusForbidden)
			return
		}
		
		next(w, r)
	}
}

// GetCSRFToken retrieves or generates a CSRF token for the current session
func (s *Server) GetCSRFToken(r *http.Request) string {
	// Try to get existing token from cookie
	cookie, err := r.Cookie("csrf_token")
	if err == nil && cookie.Value != "" {
		// Validate existing token
		if s.csrfManager.ValidateToken(cookie.Value) {
			return cookie.Value
		}
	}
	
	// Generate new token
	token, err := s.csrfManager.GenerateToken()
	if err != nil {
		return ""
	}
	
	return token
}

// SetCSRFCookie sets the CSRF token cookie
func (s *Server) SetCSRFCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
	})
}
