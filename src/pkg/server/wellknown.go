package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// requestScheme reports the scheme the client actually used. Reverse-proxy
// forwarding headers are deliberately ignored: honoring them requires a
// trusted_proxies allowlist, and trusting them without one lets any client
// rewrite the URLs published in security.txt.
func (s *Server) requestScheme(r *http.Request) string {
	if r.TLS != nil || s.config.Server.TLSEnabled {
		return "https"
	}
	return "http"
}

// requestHost returns the host the client addressed, falling back to the
// configured domain when the request carries no Host header.
func (s *Server) requestHost(r *http.Request) string {
	if r.Host != "" {
		return r.Host
	}
	return s.config.Server.Domain
}

// buildURL resolves an absolute URL for the current request. Every URL
// published in a generated well-known file is built per request so it always
// matches the host the client used.
func (s *Server) buildURL(r *http.Request, path string) string {
	return s.requestScheme(r) + "://" + s.requestHost(r) + path
}

// securityContactEmail returns the CC address for security reports, deriving
// security@{host} when no address is configured.
func (s *Server) securityContactEmail(r *http.Request) string {
	if configured := s.config.Server.Security.Contact; configured != "" {
		return configured
	}

	host := s.requestHost(r)
	if host == "" {
		return ""
	}
	// Strip any port before building the address
	if idx := strings.LastIndex(host, ":"); idx != -1 && !strings.Contains(host[idx:], "]") {
		host = host[:idx]
	}

	return "security@" + host
}

// handleSecurityTxt serves the RFC 9116 security.txt document. Contact lines
// are emitted in order of preference: private vulnerability reporting first,
// then the CC mailbox.
func (s *Server) handleSecurityTxt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var b strings.Builder
	if reportURL := s.config.Server.Security.ReportURL; reportURL != "" {
		fmt.Fprintf(&b, "Contact: %s\n", reportURL)
	}
	if email := s.securityContactEmail(r); email != "" {
		fmt.Fprintf(&b, "Contact: mailto:%s\n", email)
	}
	fmt.Fprintf(&b, "Policy: %s\n", s.buildURL(r, "/server/security"))
	fmt.Fprintf(&b, "Expires: %s\n", s.config.Server.Security.Expires)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte(b.String())); err != nil {
		log.Printf("failed to write security.txt: %v", err)
	}
}

// handleSecurityPolicy renders the human-readable counterpart of security.txt
// required by AI.md PART 11. It carries the same contact, policy and expiry
// data, resolved through the same per-request URL builder so the page and the
// file a given client fetches can never disagree behind a reverse proxy.
func (s *Server) handleSecurityPolicy(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageData
		ReportURL    string
		ContactEmail string
		PolicyURL    string
		SecurityTxt  string
		Expires      string
	}{
		PageData: PageData{
			Title:  "Security Policy",
			Active: "security",
		},
		ReportURL:    s.config.Server.Security.ReportURL,
		ContactEmail: s.securityContactEmail(r),
		PolicyURL:    s.buildURL(r, "/server/security"),
		SecurityTxt:  s.buildURL(r, "/.well-known/security.txt"),
		Expires:      s.config.Server.Security.Expires,
	}
	s.populatePageData(&data.PageData, r)

	if err := s.templateRenderer.Render(w, "layouts/base.tmpl", "pages/security.tmpl", data); err != nil {
		s.handleError(w, r, http.StatusInternalServerError, "Failed to render page", err)
	}
}

// handleLLMsTxt serves the AI-agent discovery document required by AI.md
// PART 11. Only public and authenticated API surface is advertised — admin,
// internal and metrics routes are deliberately omitted.
func (s *Server) handleLLMsTxt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	api := "/api/" + s.config.Server.APIVersion

	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", s.config.Server.ApplicationName)
	if tagline := s.config.Server.Tagline; tagline != "" {
		fmt.Fprintf(&b, "> %s\n", tagline)
	}
	fmt.Fprintf(&b, "\n## API\nBase URL: %s\n", s.buildURL(r, api))
	fmt.Fprintf(&b, "Authentication: Bearer token (obtain via POST %s/server/auth/login)\n", api)
	fmt.Fprintf(&b, "\n## Endpoints\n")
	fmt.Fprintf(&b, "- GET /server/healthz - Health check (no auth)\n")
	fmt.Fprintf(&b, "- GET /server/about - Server information (no auth)\n")
	fmt.Fprintf(&b, "- POST %s/server/auth/login - User login\n", api)
	fmt.Fprintf(&b, "- POST %s/server/auth/register - User registration\n", api)
	fmt.Fprintf(&b, "- GET %s/projects - List projects (authenticated)\n", api)
	fmt.Fprintf(&b, "- GET %s/builds - List builds (authenticated)\n", api)
	fmt.Fprintf(&b, "- GET %s/nodes - List build nodes (authenticated)\n", api)
	fmt.Fprintf(&b, "- GET %s/server/swagger - OpenAPI description (no auth)\n", api)
	fmt.Fprintf(&b, "- POST %s/server/graphql - GraphQL endpoint (authenticated)\n", api)
	fmt.Fprintf(&b, "\n## Capabilities\n")
	fmt.Fprintf(&b, "- Continuous integration pipelines and build orchestration\n")
	fmt.Fprintf(&b, "- Distributed build nodes\n")
	fmt.Fprintf(&b, "- REST and GraphQL APIs\n")
	fmt.Fprintf(&b, "\n## Contact\n")
	fmt.Fprintf(&b, "API issues: %s\n", s.buildURL(r, "/server/contact"))
	if email := s.securityContactEmail(r); email != "" {
		fmt.Fprintf(&b, "Security: %s\n", email)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte(b.String())); err != nil {
		log.Printf("failed to write llms.txt: %v", err)
	}
}

// handleWellKnown answers every unhandled /.well-known/ request. Unknown
// entries return 404 and are never redirected.
func (s *Server) handleWellKnown(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
