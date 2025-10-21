package security

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Handler handles HTTP requests for security operations
type Handler struct {
	service *Service
}

// NewHandler creates a new security handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// extractIDFromPath extracts ID from URL path like /api/v1/builds/{id}/security
func extractIDFromPath(path, prefix string) (int, error) {
	// Remove prefix and split
	path = strings.TrimPrefix(path, prefix)
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 {
		return 0, strconv.ErrSyntax
	}

	// First part should be the ID
	return strconv.Atoi(parts[0])
}

// GetBuildSecurityReports retrieves all security reports for a build
func (h *Handler) GetBuildSecurityReports(w http.ResponseWriter, r *http.Request) {
	// Extract build ID from path: /api/v1/builds/{id}/security
	buildID, err := extractIDFromPath(r.URL.Path, "/api/v1/builds/")
	if err != nil {
		http.Error(w, "Invalid build ID", http.StatusBadRequest)
		return
	}

	reports, err := h.service.GetReportsByBuild(r.Context(), buildID)
	if err != nil {
		log.Printf("Failed to get security reports for build %d: %v", buildID, err)
		http.Error(w, "Failed to retrieve security reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

// GetSecurityReport retrieves a specific security report
func (h *Handler) GetSecurityReport(w http.ResponseWriter, r *http.Request) {
	// Extract report ID from path: /api/v1/security/reports/{id}
	reportID, err := extractIDFromPath(r.URL.Path, "/api/v1/security/reports/")
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	report, err := h.service.GetReport(r.Context(), reportID)
	if err != nil {
		log.Printf("Failed to get security report %d: %v", reportID, err)
		http.Error(w, "Security report not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// ListSecurityReports lists security reports with optional filters
func (h *Handler) ListSecurityReports(w http.ResponseWriter, r *http.Request) {
	filters := make(map[string]interface{})

	// Parse query parameters
	if buildID := r.URL.Query().Get("build_id"); buildID != "" {
		if id, err := strconv.Atoi(buildID); err == nil {
			filters["build_id"] = id
		}
	}

	if scanType := r.URL.Query().Get("scan_type"); scanType != "" {
		filters["scan_type"] = scanType
	}

	if tool := r.URL.Query().Get("tool"); tool != "" {
		filters["tool"] = tool
	}

	if passed := r.URL.Query().Get("passed"); passed != "" {
		filters["passed"] = passed == "true"
	}

	if hasCritical := r.URL.Query().Get("has_critical"); hasCritical == "true" {
		filters["has_critical"] = true
	}

	if hasHigh := r.URL.Query().Get("has_high"); hasHigh == "true" {
		filters["has_high"] = true
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters["limit"] = l
		}
	}

	reports, err := h.service.repo.ListReports(r.Context(), filters)
	if err != nil {
		log.Printf("Failed to list security reports: %v", err)
		http.Error(w, "Failed to list security reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

// GetStatistics retrieves security statistics
func (h *Handler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	repo, ok := h.service.repo.(*SQLRepository)
	if !ok {
		http.Error(w, "Statistics not available", http.StatusNotImplemented)
		return
	}

	stats, err := repo.GetStatistics(r.Context())
	if err != nil {
		log.Printf("Failed to get security statistics: %v", err)
		http.Error(w, "Failed to retrieve statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetConfig retrieves the current security scan configuration
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.service.getEffectiveConfig()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// UpdateConfig updates the security scan configuration
func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var config ScanConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.service.UpdateConfig(&config)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Security configuration updated",
	})
}

// TriggerScan manually triggers a security scan for a build
func (h *Handler) TriggerScan(w http.ResponseWriter, r *http.Request) {
	// Extract build ID from path: /api/v1/builds/{id}/security/scan
	buildID, err := extractIDFromPath(r.URL.Path, "/api/v1/builds/")
	if err != nil {
		http.Error(w, "Invalid build ID", http.StatusBadRequest)
		return
	}

	// Parse optional scan configuration
	var config *ScanConfig
	if r.Body != nil {
		var reqConfig ScanConfig
		if err := json.NewDecoder(r.Body).Decode(&reqConfig); err == nil {
			config = &reqConfig
		}
	}

	// Get build workspace path (in a real implementation, this would come from the build)
	// For now, we'll use a placeholder
	target := "/var/lib/casci/workspaces/" + strconv.Itoa(buildID)

	// Queue the scan
	h.service.ScanBuildAsync(buildID, target, config)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Security scan queued",
		"build_id": buildID,
	})
}

// WebhookPayload represents a webhook notification payload
type WebhookPayload struct {
	BuildID    int            `json:"build_id"`
	Status     string         `json:"status"`
	Passed     bool           `json:"passed"`
	Summary    *ScanSummary   `json:"summary"`
	ReportURLs []string       `json:"report_urls"`
	Timestamp  string         `json:"timestamp"`
}

// SendWebhook sends security scan results to a webhook (helper function)
func SendWebhook(url string, result *CompleteScanResult) error {
	_ = WebhookPayload{
		BuildID: result.BuildID,
		Status:  "completed",
		Passed:  result.Summary.Passed,
		Summary: result.Summary,
	}

	// TODO: Send payload to webhook endpoint
	// For now, just skip marshaling

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}
