package server

import (
	"net/http"
	"runtime"
	"time"
)

// PageData is the base data structure for all pages
type PageData struct {
	Title                 string
	Version               string
	Mode                  string
	User                  *UserContext
	Active                string
	ExtraScripts          []string
	CSRFToken             string
	CurrentYear           int
	FooterCustomHTML      string
	TrackingID            string
	CookieConsentEnabled  bool
	CookieConsentMessage  string
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
	Status     string                 `json:"status"`
	Version    string                 `json:"version"`
	Mode       string                 `json:"mode"`
	Uptime     string                 `json:"uptime"`
	Timestamp  string                 `json:"timestamp"`
	Checks     map[string]string      `json:"checks"`
	SystemInfo *SystemInfo            `json:"system_info,omitempty"`
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
	Config interface{}
}

// ErrorData holds error page data
type ErrorData struct {
	PageData
	ErrorCode    int
	ErrorTitle   string
	ErrorMessage string
	ErrorDetails string
}

// handleIndex serves the home page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := HomeData{
		PageData: PageData{
			Title:   "Home",
			Version: "0.1.0-dev",
			Mode:    "production", // TODO: Get from config
			Active:  "home",
		},
	}
	
	// Populate footer configuration
	s.populateFooterConfig(&data.PageData)

	// Get stats from services
	// TODO: Implement actual stats
	data.Stats.TotalProjects = 0
	data.Stats.TotalBuilds = 0
	data.Stats.ActiveNodes = 0
	data.Stats.SuccessRate = 0.0

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleHealthz serves the health check HTML page
func (s *Server) handleHealthzHTML(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now() // TODO: Store actual server start time

	data := struct {
		PageData
		Health HealthData
	}{
		PageData: PageData{
			Title:   "Health Check",
			Version: "0.1.0-dev",
			Mode:    "production",
			Active:  "health",
		},
		Health: HealthData{
			Status:    "healthy",
			Version:   "0.1.0-dev",
			Mode:      "production",
			Uptime:    time.Since(startTime).String(),
			Timestamp: time.Now().Format(time.RFC3339),
			Checks: map[string]string{
				"database": "ok",
				"cache":    "ok",
				"disk":     "ok",
			},
			SystemInfo: &SystemInfo{
				CPUUsage:    0.0,    // TODO: Get actual metrics
				MemoryUsage: 0.0,
				DiskUsage:   0.0,
				Goroutines:  runtime.NumGoroutine(),
			},
		},
	}

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleAdminDashboard serves the admin dashboard
func (s *Server) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	// TODO: Check authentication
	
	data := DashboardData{
		PageData: PageData{
			Title:   "Dashboard",
			Version: "0.1.0-dev",
			Mode:    "production",
			Active:  "dashboard",
		},
	}

	// TODO: Get actual stats from services
	data.Stats.TotalProjects = 0
	data.Stats.TotalBuilds = 0
	data.Stats.TotalUsers = 1
	data.Stats.ActiveNodes = 0

	data.SystemMetrics.CPUUsage = 0.0
	data.SystemMetrics.MemoryUsage = 0.0
	data.SystemMetrics.DiskUsage = 0.0

	if err := s.templateRenderer.Render(w, "layouts/admin.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleAdminSettings serves the admin settings page
func (s *Server) handleAdminSettings(w http.ResponseWriter, r *http.Request) {
	// TODO: Check authentication

	data := SettingsData{
		PageData: PageData{
			Title:   "Settings",
			Version: "0.1.0-dev",
			Mode:    "production",
			Active:  "settings",
		},
		Config: s.config, // TODO: Sanitize sensitive data
	}

	if err := s.templateRenderer.Render(w, "layouts/admin.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleServerAbout renders the about page
func (s *Server) handleServerAbout(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "About - CASCI",
		Version: "0.1.0",
		Mode:    "production",
	}
	s.populateFooterConfig(&data)

	if err := s.templateRenderer.Render(w, "pages/about.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleServerPrivacy renders the privacy policy page
func (s *Server) handleServerPrivacy(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Privacy Policy - CASCI",
		Version: "0.1.0",
		Mode:    "production",
	}
	s.populateFooterConfig(&data)

	if err := s.templateRenderer.Render(w, "pages/privacy.tmpl", data); err != nil {
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
			Title:     "Contact Us - CASCI",
			Version:   "0.1.0",
			Mode:      "production",
			CSRFToken: csrfToken,
		},
	}
	s.populateFooterConfig(&data.PageData)

	if r.Method == http.MethodPost {
		// Parse form
		if err := r.ParseForm(); err != nil {
			data.Error = "Invalid form data"
		} else {
			// Validate CSRF token
			formToken := r.FormValue("csrf_token")
			if !s.csrfManager.ValidateToken(formToken) {
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

	if err := s.templateRenderer.Render(w, "pages/contact.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleServerHelp renders the help/documentation page
func (s *Server) handleServerHelp(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Help & Documentation - CASCI",
		Version: "0.1.0",
		Mode:    "production",
	}
	s.populateFooterConfig(&data)

	if err := s.templateRenderer.Render(w, "pages/help.tmpl", data); err != nil {
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
			Title:   title,
			Version: "0.1.0-dev",
			Mode:    "production",
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

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", data); err != nil {
		// Fallback to plain text if template rendering fails
		http.Error(w, message, statusCode)
	}
}

// handle404 handles 404 errors
func (s *Server) handle404(w http.ResponseWriter, r *http.Request) {
	s.handleError(w, r, http.StatusNotFound, "The page you are looking for does not exist.", nil)
}

// populateFooterConfig populates footer configuration in PageData
func (s *Server) populateFooterConfig(data *PageData) {
	data.CurrentYear = time.Now().Year()
	data.FooterCustomHTML = s.config.Server.Footer.CustomHTML
	data.TrackingID = s.config.Server.Footer.TrackingID
	data.CookieConsentEnabled = s.config.Server.Footer.CookieConsent.Enabled
	data.CookieConsentMessage = s.config.Server.Footer.CookieConsent.Message
	data.CookieConsentPolicyURL = s.config.Server.Footer.CookieConsent.PolicyURL
}
