package server

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"sync"
)

//go:embed all:templates
var templatesFS embed.FS

//go:embed all:static
var staticFS embed.FS

// staticRoot is staticFS re-rooted at its "static" directory so that a request
// path, once its /static/ prefix has been stripped, maps directly onto an entry
// in the filesystem.
var staticRoot = mustSubFS(staticFS, "static")

// mustSubFS re-roots an embedded filesystem at dir. The embed is resolved at
// compile time, so a failure here means the directory was renamed without the
// embed directive being updated — a build defect, not a runtime condition.
func mustSubFS(fsys embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic("embedded assets are missing directory " + dir + ": " + err.Error())
	}
	return sub
}

// TemplateRenderer handles HTML template rendering.
//
// Every page defines a template block named "content", so all pages cannot
// share one namespace — parsing them together makes the last page parsed win
// for every request. Each page therefore gets its own template set containing
// the layouts, partials and components plus exactly that one page.
type TemplateRenderer struct {
	sets    map[string]*template.Template
	mu      sync.RWMutex
	devMode bool
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(devMode bool) (*TemplateRenderer, error) {
	tr := &TemplateRenderer{
		devMode: devMode,
	}

	if err := tr.loadTemplates(); err != nil {
		return nil, err
	}

	return tr, nil
}

// loadTemplates builds one template set per page from the embedded FS
func (tr *TemplateRenderer) loadTemplates() error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	pages, err := pageTemplatePaths()
	if err != nil {
		return err
	}

	sets := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		set := template.New("").Funcs(templateFuncs())

		set, err = set.ParseFS(templatesFS,
			"templates/layouts/*.tmpl",
			"templates/partials/*.tmpl",
			"templates/components/*.tmpl",
			"templates/"+page,
		)
		if err != nil {
			return fmt.Errorf("parse %s: %w", page, err)
		}

		sets[page] = set
	}

	tr.sets = sets
	return nil
}

// templateFuncs returns the function map every template set shares
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"eq": func(a, b interface{}) bool {
			return a == b
		},
	}
}

// pageTemplatePaths lists every renderable page relative to the templates root
func pageTemplatePaths() ([]string, error) {
	var pages []string
	for _, dir := range []string{"pages", "admin"} {
		entries, err := fs.Glob(templatesFS, "templates/"+dir+"/*.tmpl")
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			pages = append(pages, strings.TrimPrefix(entry, "templates/"))
		}
	}
	return pages, nil
}

// Render renders one page inside the named layout. layout is a path such as
// "layouts/base.tmpl" and page is a path such as "pages/index.tmpl".
func (tr *TemplateRenderer) Render(w http.ResponseWriter, layout, page string, data interface{}) error {
	// In development mode, reload templates on each request
	if tr.devMode {
		if err := tr.loadTemplates(); err != nil {
			return err
		}
	}

	tr.mu.RLock()
	set, ok := tr.sets[page]
	tr.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown page template %q", page)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	return set.ExecuteTemplate(w, path.Base(layout), data)
}

type securityContextKey string

const (
	// cspNonceContextKey carries the per-request CSP nonce
	cspNonceContextKey securityContextKey = "csp_nonce"
	// requestIDContextKey carries the per-request identifier
	requestIDContextKey securityContextKey = "request_id"
)

// permissionsPolicy is the default Permissions-Policy: every browser feature the
// application does not use is disabled, per AI.md PART 11
const permissionsPolicy = "accelerometer=(), ambient-light-sensor=(), autoplay=(self), browsing-topics=(), camera=(), display-capture=(), encrypted-media=(self), fullscreen=(self), geolocation=(), gyroscope=(), interest-cohort=(), magnetometer=(), microphone=(), midi=(), payment=(self), picture-in-picture=(self), publickey-credentials-get=(self), screen-wake-lock=(), serial=(), storage-access=(self), usb=(), web-share=(self), xr-spatial-tracking=()"

// randomHex returns a hex string built from n random bytes
func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// newCSPNonce returns a fresh base64 CSP nonce
func newCSPNonce() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

// CSPNonceFromRequest returns the per-request CSP nonce assigned by the
// security headers middleware
func CSPNonceFromRequest(r *http.Request) string {
	if nonce, ok := r.Context().Value(cspNonceContextKey).(string); ok {
		return nonce
	}
	return ""
}

// RequestIDFromRequest returns the per-request identifier assigned by the
// security headers middleware
func RequestIDFromRequest(r *http.Request) string {
	if id, ok := r.Context().Value(requestIDContextKey).(string); ok {
		return id
	}
	return ""
}

// Origins the Google Analytics tag loads from and reports to. They are only
// ever added to the CSP when a tracking ID is configured.
const (
	analyticsScriptOrigin  = "https://www.googletagmanager.com"
	analyticsCollectOrigin = "https://www.google-analytics.com https://region1.google-analytics.com"
)

// buildCSP assembles the Content-Security-Policy for a request. script-src is
// restricted to 'self' plus the per-request nonce so injected inline script can
// never execute
func (s *Server) buildCSP(nonce string) string {
	scriptSrc := "'self'"
	if nonce != "" {
		scriptSrc += " 'nonce-" + nonce + "'"
	}
	connectSrc := "'self'"

	// The analytics tag is only reachable when the operator has configured a
	// tracking ID, so its origins are added only in that case — a default
	// install never widens script-src beyond 'self' plus the nonce
	if s.config.Server.Footer.TrackingID != "" {
		scriptSrc += " " + analyticsScriptOrigin
		connectSrc += " " + analyticsCollectOrigin
	}

	directives := []string{
		"default-src 'self'",
		"script-src " + scriptSrc,
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data: blob: https:",
		"font-src 'self' https:",
		"connect-src " + connectSrc,
		"media-src 'self' blob:",
		"worker-src 'self' blob:",
		"manifest-src 'self'",
		"frame-src 'self'",
		"frame-ancestors 'self'",
		"base-uri 'self'",
		"form-action 'self'",
		"object-src 'none'",
	}

	if s.config.Server.TLSEnabled {
		directives = append(directives, "upgrade-insecure-requests")
	}

	return strings.Join(directives, "; ")
}

// SecurityHeadersMiddleware applies the AI.md PART 11 security headers to every
// response — API, static, GraphQL and HTML alike
func (s *Server) SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = randomHex(16)
		}
		nonce := newCSPNonce()

		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("X-XSS-Protection", "1; mode=block")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Permitted-Cross-Domain-Policies", "none")
		h.Set("Origin-Agent-Cluster", "?1")
		h.Set("Cross-Origin-Opener-Policy", "unsafe-none")
		h.Set("Cross-Origin-Embedder-Policy", "unsafe-none")
		h.Set("Cross-Origin-Resource-Policy", "cross-origin")
		h.Set("Content-Security-Policy", s.buildCSP(nonce))
		h.Set("Permissions-Policy", permissionsPolicy)
		h.Set("X-Request-ID", requestID)

		// HSTS is only meaningful — and only safe — once TLS is serving
		if s.config.Server.TLSEnabled {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}

		ctx := context.WithValue(r.Context(), cspNonceContextKey, nonce)
		ctx = context.WithValue(ctx, requestIDContextKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
