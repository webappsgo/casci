package server

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/webappsgo/casci/src/internal/config"
	"github.com/webappsgo/casci/src/pkg/users"
)

// negotiateHealthzFormat picks the healthz response format from the
// request's Accept header / extension, per PART 14 content negotiation:
// .txt and text/plain are always priority 1
func negotiateHealthzFormat(r *http.Request) string {
	if strings.HasSuffix(r.URL.Path, ".txt") {
		return "text"
	}
	if strings.HasSuffix(r.URL.Path, ".json") {
		return "json"
	}
	accept := r.Header.Get("Accept")
	switch {
	case strings.Contains(accept, "text/plain"):
		return "text"
	case strings.Contains(accept, "application/json"):
		return "json"
	default:
		return "html"
	}
}

// PageData is the base data structure for all pages
type PageData struct {
	Title                  string
	SiteName               string
	Tagline                string
	Version                string
	Mode                   string
	AdminPath              string
	APIVersion             string
	User                   *UserContext
	Active                 string
	ExtraScripts           []string
	CSRFToken              string
	CSPNonce               string
	CurrentYear            int
	FooterCustomHTML       string
	TrackingID             string
	CookieConsentEnabled   bool
	CookieConsentMessage   string
	CookieConsentPolicyURL string
}

// UserContext holds user info for templates
type UserContext struct {
	ID       int64
	Username string
	Email    string
	IsAdmin  bool
}

// HealthData holds health check data
type HealthData struct {
	Status     string            `json:"status"`
	Version    string            `json:"version"`
	Mode       string            `json:"mode"`
	Uptime     string            `json:"uptime"`
	Timestamp  string            `json:"timestamp"`
	Checks     map[string]string `json:"checks"`
	SystemInfo *SystemInfo       `json:"system_info,omitempty"`
}

// SystemInfo holds system metrics
type SystemInfo struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	Goroutines  int     `json:"goroutines"`
}

// HomeData holds home page data
type HomeData struct {
	PageData
	Stats struct {
		TotalProjects int64
		TotalBuilds   int64
		ActiveNodes   int64
		SuccessRate   float64
	}
}

// DashboardData holds admin dashboard data
type DashboardData struct {
	PageData
	Stats struct {
		TotalProjects int64
		TotalBuilds   int64
		TotalUsers    int64
		ActiveNodes   int64
	}
	RecentBuilds []struct {
		ProjectName string
		BuildNumber int64
		Status      string
		Duration    string
	}
	SystemMetrics struct {
		CPUUsage    float64
		MemoryUsage float64
		DiskUsage   float64
	}
}

// SettingsData holds admin settings data
type SettingsData struct {
	PageData
	Config SettingsView
}

// SettingsView is the typed view-model behind the admin settings page. It
// mirrors the server.yml shape the template renders, so a field the template
// asks for that the configuration does not actually carry is a compile-time
// error here rather than a 500 at render time.
type SettingsView struct {
	Server SettingsServerView
	Web    SettingsWebView
}

// SettingsServerView is the `server:` block as shown in the admin panel.
type SettingsServerView struct {
	Title      string
	Tagline    string
	FQDN       string
	Port       int
	Address    string
	Mode       string
	AdminPath  string
	APIVersion string
	SSL        SettingsSSLView
	Database   SettingsDatabaseView
}

// SettingsSSLView is the `server.ssl:` block as shown in the admin panel.
type SettingsSSLView struct {
	Enabled    bool
	MinVersion string
}

// SettingsDatabaseView is the `server.database:` block as shown in the admin
// panel. The DSN is deliberately absent — it carries credentials.
type SettingsDatabaseView struct {
	Driver      string
	MaxConns    int
	MaxIdleConn int
}

// SettingsWebView is the `web:` block as shown in the admin panel.
type SettingsWebView struct {
	CORS                 string
	CookieConsentEnabled bool
}

// ErrorData holds error page data
type ErrorData struct {
	PageData
	ErrorCode    int
	ErrorTitle   string
	ErrorMessage string
	ErrorDetails string
}

// percentOf returns used as a percentage of total. A zero total means the
// metric was never collected, so it reports 0 rather than dividing by zero.
func percentOf(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) / float64(total) * 100
}

// handleIndex serves the home page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := HomeData{
		PageData: PageData{
			Title:  "Home",
			Active: "home",
		},
	}

	s.populatePageData(&data.PageData, r)

	// Live counts. A stats query failing must not take the home page down, so
	// each count falls back to zero and the page still renders.
	ctx := r.Context()
	if count, err := s.projectService.Count(ctx, 0); err == nil {
		data.Stats.TotalProjects = int64(count)
	}
	if count, err := s.buildService.Count(ctx, 0); err == nil {
		data.Stats.TotalBuilds = int64(count)
	}
	if online, err := s.nodeService.ListOnline(ctx); err == nil {
		data.Stats.ActiveNodes = int64(len(online))
	}
	data.Stats.SuccessRate = s.metricsCollector.GetSuccessRate()

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", "pages/index.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleHealthz serves the health check via content negotiation, per PART
// 13: JSON and plain text respond directly, everything else gets the HTML
// dashboard page
func (s *Server) handleHealthzHTML(w http.ResponseWriter, r *http.Request) {
	system := s.metricsCollector.GetSystemMetrics()

	// The database check is the only dependency whose failure changes the
	// reported status, so it is probed rather than asserted
	status := "healthy"
	dbCheck := "ok"
	if err := s.db.DB().PingContext(r.Context()); err != nil {
		status = "degraded"
		dbCheck = "error"
	}

	switch negotiateHealthzFormat(r) {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		if status != "healthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		fmt.Fprintf(w, "{\"status\": \"%s\", \"checks\": {\"database\": \"%s\"}}\n", status, dbCheck)
		return
	case "text":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if status != "healthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		fmt.Fprintf(w, "status: %s\ndatabase: %s\n", status, dbCheck)
		return
	}

	data := struct {
		PageData
		Health HealthData
	}{
		PageData: PageData{
			Title:  "Health Check",
			Active: "health",
		},
		Health: HealthData{
			Status:    status,
			Version:   config.Version(),
			Mode:      s.config.Server.Mode,
			Uptime:    time.Since(s.startedAt).Truncate(time.Second).String(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Checks: map[string]string{
				"database": dbCheck,
			},
			SystemInfo: &SystemInfo{
				CPUUsage:    system.CPUUsagePercent,
				MemoryUsage: percentOf(system.MemoryUsageBytes, system.MemoryTotalBytes),
				DiskUsage:   percentOf(system.DiskUsageBytes, system.DiskTotalBytes),
				Goroutines:  runtime.NumGoroutine(),
			},
		},
	}

	s.populatePageData(&data.PageData, r)

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", "pages/healthz.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleAdminDashboard serves the admin dashboard
func (s *Server) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	admin := s.requireAdmin(w, r)
	if admin == nil {
		return
	}

	data := DashboardData{
		PageData: PageData{
			Title:  "Dashboard",
			Active: "dashboard",
			User:   admin,
		},
	}
	s.populatePageData(&data.PageData, r)

	// Live counts. A stats query failing must not take the dashboard down, so
	// each count falls back to zero and the page still renders.
	ctx := r.Context()
	if count, err := s.projectService.Count(ctx, 0); err == nil {
		data.Stats.TotalProjects = int64(count)
	}
	if count, err := s.buildService.Count(ctx, 0); err == nil {
		data.Stats.TotalBuilds = int64(count)
	}
	if count, err := s.userService.Count(ctx); err == nil {
		data.Stats.TotalUsers = int64(count)
	}
	if online, err := s.nodeService.ListOnline(ctx); err == nil {
		data.Stats.ActiveNodes = int64(len(online))
	}

	system := s.metricsCollector.GetSystemMetrics()
	data.SystemMetrics.CPUUsage = system.CPUUsagePercent
	data.SystemMetrics.MemoryUsage = percentOf(system.MemoryUsageBytes, system.MemoryTotalBytes)
	data.SystemMetrics.DiskUsage = percentOf(system.DiskUsageBytes, system.DiskTotalBytes)

	if err := s.templateRenderer.Render(w, "layouts/admin.tmpl", "admin/dashboard.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleAdminSettings serves the admin settings page
func (s *Server) handleAdminSettings(w http.ResponseWriter, r *http.Request) {
	admin := s.requireAdmin(w, r)
	if admin == nil {
		return
	}

	data := SettingsData{
		PageData: PageData{
			Title:  "Settings",
			Active: "settings",
			User:   admin,
		},
		// Only non-secret settings — the encryption key and database DSN are
		// never handed to a template
		Config: sanitizeConfigForDisplay(s.config),
	}
	s.populatePageData(&data.PageData, r)

	if err := s.templateRenderer.Render(w, "layouts/admin.tmpl", "admin/settings.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleServerAbout renders the about page
func (s *Server) handleServerAbout(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "About",
	}
	s.populatePageData(&data, r)

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", "pages/about.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleServerPrivacy renders the privacy policy page
func (s *Server) handleServerPrivacy(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Privacy Policy",
	}
	s.populatePageData(&data, r)

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", "pages/privacy.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleServerContact renders and processes the contact form
func (s *Server) handleServerContact(w http.ResponseWriter, r *http.Request) {
	type ContactData struct {
		PageData
		Success bool
		Error   string
	}

	// Generate CSRF token
	csrfToken := s.GetCSRFToken(r)
	s.SetCSRFCookie(w, csrfToken)

	data := ContactData{
		PageData: PageData{
			Title:     "Contact Us",
			CSRFToken: csrfToken,
		},
	}
	s.populatePageData(&data.PageData, r)

	if r.Method == http.MethodPost {
		// Parse form
		if err := r.ParseForm(); err != nil {
			data.Error = "Invalid form data"
		} else {
			// Validate CSRF token
			formToken := r.FormValue("csrf_token")
			if !s.csrfManager.ValidateToken(sessionID(r), formToken) {
				data.Error = "Invalid security token. Please refresh the page and try again."
			} else {
				name := r.FormValue("name")
				email := r.FormValue("email")
				subject := r.FormValue("subject")
				message := r.FormValue("message")
				captcha := r.FormValue("captcha")

				// Simple captcha check (2 + 3 = 5)
				if captcha != "5" {
					data.Error = "Incorrect security check answer"
				} else if name == "" || email == "" || subject == "" || message == "" {
					data.Error = "All fields are required"
				} else {
					// TODO: Send email using notification service
					// For now, just log and show success
					data.Success = true
					// Generate new token after successful submission
					csrfToken = s.GetCSRFToken(r)
					s.SetCSRFCookie(w, csrfToken)
					data.CSRFToken = csrfToken
				}
			}
		}
	}

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", "pages/contact.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleServerHelp renders the help/documentation page
func (s *Server) handleServerHelp(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Help & Documentation",
	}
	s.populatePageData(&data, r)

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", "pages/help.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleError renders an error page
func (s *Server) handleError(w http.ResponseWriter, r *http.Request, statusCode int, message string, err error) {
	w.WriteHeader(statusCode)

	errorTitles := map[int]string{
		http.StatusNotFound:            "Page Not Found",
		http.StatusInternalServerError: "Internal Server Error",
		http.StatusForbidden:           "Forbidden",
		http.StatusUnauthorized:        "Unauthorized",
		http.StatusBadRequest:          "Bad Request",
	}

	title, ok := errorTitles[statusCode]
	if !ok {
		title = "Error"
	}

	data := ErrorData{
		PageData: PageData{
			Title: title,
		},
		ErrorCode:    statusCode,
		ErrorTitle:   title,
		ErrorMessage: message,
	}

	// Only show error details in development mode
	if s.config != nil && err != nil {
		// TODO: Check if development mode
		// data.ErrorDetails = err.Error()
	}

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", "pages/error.tmpl", data); err != nil {
		// Fallback to plain text if template rendering fails
		http.Error(w, message, statusCode)
	}
}

// handle404 handles 404 errors
func (s *Server) handle404(w http.ResponseWriter, r *http.Request) {
	s.handleError(w, r, http.StatusNotFound, "The page you are looking for does not exist.", nil)
}

// requireAdmin resolves the authenticated administrator for an admin page. It
// writes the appropriate error page and returns nil when the caller is not an
// authenticated administrator
func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) *UserContext {
	user := users.GetUserFromContext(r.Context())
	if user == nil {
		s.handleError(w, r, http.StatusUnauthorized, "You must sign in to access this page.", nil)
		return nil
	}

	if !user.IsAdmin {
		s.handleError(w, r, http.StatusForbidden, "Administrator access is required.", nil)
		return nil
	}

	return &UserContext{
		ID:       int64(user.ID),
		Username: user.Username,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
	}
}

// sanitizeConfigForDisplay returns the operator-visible subset of the running
// configuration. Secrets (encryption key, database DSN) are excluded entirely
func sanitizeConfigForDisplay(cfg *config.Config) SettingsView {
	return SettingsView{
		Server: SettingsServerView{
			Title:      cfg.Server.ApplicationName,
			Tagline:    cfg.Server.Tagline,
			FQDN:       cfg.Server.Domain,
			Port:       cfg.Server.Port,
			Address:    cfg.Server.Host,
			Mode:       cfg.Server.Mode,
			AdminPath:  cfg.Server.AdminPath,
			APIVersion: cfg.Server.APIVersion,
			SSL: SettingsSSLView{
				Enabled:    cfg.Server.TLSEnabled,
				MinVersion: cfg.Server.TLSMinVersion,
			},
			Database: SettingsDatabaseView{
				Driver:      cfg.Database.Type,
				MaxConns:    cfg.Database.MaxConns,
				MaxIdleConn: cfg.Database.MaxIdleConn,
			},
		},
		Web: SettingsWebView{
			CORS:                 cfg.Server.CORS,
			CookieConsentEnabled: cfg.Server.Footer.CookieConsent.Enabled,
		},
	}
}

// populatePageData fills the fields every page shares: build metadata, the
// operator-configured admin path and API version, the per-request CSP nonce,
// and the footer configuration. Every handler that renders a template must
// call it, otherwise the layout emits an empty nonce and its inline scripts
// are blocked by the CSP
func (s *Server) populatePageData(data *PageData, r *http.Request) {
	data.Version = config.Version()
	data.Mode = s.config.Server.Mode
	data.AdminPath = s.config.Server.AdminPath
	data.APIVersion = s.config.Server.APIVersion
	data.SiteName = s.config.Server.ApplicationName
	data.Tagline = s.config.Server.Tagline
	data.CSPNonce = CSPNonceFromRequest(r)
	data.CurrentYear = time.Now().Year()
	data.FooterCustomHTML = s.config.Server.Footer.CustomHTML
	data.TrackingID = s.config.Server.Footer.TrackingID
	data.CookieConsentEnabled = s.config.Server.Footer.CookieConsent.Enabled
	data.CookieConsentMessage = s.config.Server.Footer.CookieConsent.Message
	data.CookieConsentPolicyURL = s.config.Server.Footer.CookieConsent.PolicyURL
}
