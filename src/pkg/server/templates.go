package server

import (
	"embed"
	"html/template"
	"net/http"
	"sync"
)

//go:embed all:templates
var templatesFS embed.FS

//go:embed all:static
var staticFS embed.FS

// TemplateRenderer handles HTML template rendering
type TemplateRenderer struct {
	templates *template.Template
	mu        sync.RWMutex
	devMode   bool
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

// loadTemplates loads all templates from embedded FS
func (tr *TemplateRenderer) loadTemplates() error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	tmpl := template.New("").Funcs(template.FuncMap{
		"eq": func(a, b interface{}) bool {
			return a == b
		},
	})

	// Parse all template files
	var err error
	tmpl, err = tmpl.ParseFS(templatesFS,
		"templates/layouts/*.tmpl",
		"templates/partials/*.tmpl",
		"templates/pages/*.tmpl",
		"templates/admin/*.tmpl",
		"templates/components/*.tmpl",
	)
	if err != nil {
		return err
	}

	tr.templates = tmpl
	return nil
}

// Render renders a template with data
func (tr *TemplateRenderer) Render(w http.ResponseWriter, name string, data interface{}) error {
	// In development mode, reload templates on each request
	if tr.devMode {
		if err := tr.loadTemplates(); err != nil {
			return err
		}
	}

	tr.mu.RLock()
	defer tr.mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Add security headers (TEMPLATE.md compliant)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

	return tr.templates.ExecuteTemplate(w, name, data)
}

// RenderWithLayout renders a template with a specific layout
func (tr *TemplateRenderer) RenderWithLayout(w http.ResponseWriter, layout, content string, data interface{}) error {
	// In development mode, reload templates
	if tr.devMode {
		if err := tr.loadTemplates(); err != nil {
			return err
		}
	}

	tr.mu.RLock()
	defer tr.mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Add security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

	// First define the content template
	if err := tr.templates.ExecuteTemplate(w, content, data); err != nil {
		return err
	}

	// Then wrap with layout
	return tr.templates.ExecuteTemplate(w, layout, data)
}
