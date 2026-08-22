package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"
)

// csrfCookieName is the cookie holding the double-submit CSRF token
const csrfCookieName = "csrf_token"

// sessionCookieName is the browser session cookie CSRF tokens are bound to
const sessionCookieName = "casci_session"

// csrfExemptPaths are endpoints that cannot supply a CSRF token: they are
// public or authenticated by a provider signature / registration token.
// The API entries are matched as suffixes because the /api/{api_version}/
// prefix is operator-configurable.
var csrfExemptPaths = []string{
	"/webhook",
}

// csrfExemptAPISuffixes are the API paths exempt from CSRF regardless of the
// configured api_version prefix
var csrfExemptAPISuffixes = []string{
	"/auth/login",
	"/auth/register",
	"/nodes/register",
}

// CSRFToken represents a CSRF token bound to a session with expiration
type CSRFToken struct {
	Token     string
	SessionID string
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

// GenerateToken generates a new CSRF token bound to the given session
func (cm *CSRFManager) GenerateToken(sessionID string) (string, error) {
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
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return token, nil
}

// ValidateToken reports whether the token is live and bound to the given session
func (cm *CSRFManager) ValidateToken(sessionID, token string) bool {
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

	// A token issued for one session must never authorize another
	return subtle.ConstantTimeCompare([]byte(t.SessionID), []byte(sessionID)) == 1
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

// sessionID returns the session identifier a CSRF token is bound to. Browsers
// that have no session cookie yet are bound to the empty (anonymous) session
func sessionID(r *http.Request) string {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		return cookie.Value
	}
	return ""
}

// csrfExempt reports whether the path is exempt from CSRF validation
func csrfExempt(path string) bool {
	for _, p := range csrfExemptPaths {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}

	if !strings.HasPrefix(path, "/api/") {
		return false
	}

	for _, suffix := range csrfExemptAPISuffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}

	// Node agents authenticate with their own registration token
	return strings.Contains(path, "/nodes/") && strings.HasSuffix(path, "/heartbeat")
}

// CSRFMiddleware validates the CSRF token on every state-changing request that
// could carry browser-attached cookie credentials (AI.md PART 16)
func (s *Server) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read-only methods cannot change state (RFC 9110)
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		// Bearer credentials are never auto-attached by the browser, so there is
		// no cross-site forgery vector to defend against
		if r.Header.Get("Authorization") != "" || r.Header.Get("X-API-Token") != "" {
			next.ServeHTTP(w, r)
			return
		}

		// WebSocket upgrades authenticate at the connection level
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			next.ServeHTTP(w, r)
			return
		}

		if csrfExempt(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Double-submit: the submitted token must match the cookie the browser
		// sent, and the cookie must be a live token bound to this session
		cookie, err := r.Cookie(csrfCookieName)
		if err != nil || cookie.Value == "" {
			s.rejectCSRF(w)
			return
		}

		token := r.FormValue("csrf_token")
		if token == "" {
			token = r.Header.Get("X-CSRF-Token")
		}

		if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(token)) != 1 {
			s.rejectCSRF(w)
			return
		}

		if !s.csrfManager.ValidateToken(sessionID(r), token) {
			s.rejectCSRF(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// rejectCSRF writes the canonical CSRF failure response
func (s *Server) rejectCSRF(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"ok": false, "error": "CSRF_FAILED", "message": "CSRF token validation failed"}`))
}

// GetCSRFToken retrieves or generates a CSRF token for the current session
func (s *Server) GetCSRFToken(r *http.Request) string {
	session := sessionID(r)

	// Try to get existing token from cookie
	cookie, err := r.Cookie(csrfCookieName)
	if err == nil && cookie.Value != "" {
		// Validate existing token
		if s.csrfManager.ValidateToken(session, cookie.Value) {
			return cookie.Value
		}
	}

	// Generate new token
	token, err := s.csrfManager.GenerateToken(session)
	if err != nil {
		return ""
	}

	return token
}

// SetCSRFCookie sets the CSRF token cookie
func (s *Server) SetCSRFCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:  csrfCookieName,
		Value: token,
		Path:  "/",
		// 24 hours
		MaxAge: 86400,
		// The form must be able to read the token back for double-submit
		HttpOnly: false,
		Secure:   s.config.Server.TLSEnabled,
		SameSite: http.SameSiteStrictMode,
	})
}
