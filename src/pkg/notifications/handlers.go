package notifications

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Handler handles HTTP requests for notification operations
type Handler struct {
	service *Service
}

// NewHandler creates a new notification handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ListConfigs lists all notification configurations for a user
func (h *Handler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	configs, err := h.service.GetConfigsByUser(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get notification configs: %v", err)
		http.Error(w, "Failed to retrieve notification configs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}

// GetConfig retrieves a specific notification configuration
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	// Extract config ID from path
	configID, err := extractIDFromPath(r.URL.Path, "/api/v1/notifications/")
	if err != nil {
		http.Error(w, "Invalid config ID", http.StatusBadRequest)
		return
	}

	config, err := h.service.GetConfig(r.Context(), configID)
	if err != nil {
		log.Printf("Failed to get notification config %d: %v", configID, err)
		http.Error(w, "Notification config not found", http.StatusNotFound)
		return
	}

	// Verify user owns this config
	userID := getUserIDFromContext(r)
	if config.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// CreateConfig creates a new notification configuration
func (h *Handler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var config NotificationConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID
	config.UserID = userID

	// Validate configuration
	if config.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if config.Type == "" {
		http.Error(w, "Type is required", http.StatusBadRequest)
		return
	}

	if config.Config == nil {
		config.Config = make(map[string]string)
	}

	if config.Events == nil {
		config.Events = []EventType{EventBuildSuccess, EventBuildFailure}
	}

	// Create configuration
	if err := h.service.CreateConfig(r.Context(), &config); err != nil {
		log.Printf("Failed to create notification config: %v", err)
		http.Error(w, "Failed to create notification config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(config)
}

// UpdateConfig updates a notification configuration
func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract config ID from path
	configID, err := extractIDFromPath(r.URL.Path, "/api/v1/notifications/")
	if err != nil {
		http.Error(w, "Invalid config ID", http.StatusBadRequest)
		return
	}

	// Get existing config to verify ownership
	existing, err := h.service.GetConfig(r.Context(), configID)
	if err != nil {
		http.Error(w, "Notification config not found", http.StatusNotFound)
		return
	}

	if existing.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse update
	var config NotificationConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Preserve immutable fields
	config.ID = configID
	config.UserID = userID

	// Update configuration
	if err := h.service.UpdateConfig(r.Context(), &config); err != nil {
		log.Printf("Failed to update notification config: %v", err)
		http.Error(w, "Failed to update notification config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// DeleteConfig deletes a notification configuration
func (h *Handler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract config ID from path
	configID, err := extractIDFromPath(r.URL.Path, "/api/v1/notifications/")
	if err != nil {
		http.Error(w, "Invalid config ID", http.StatusBadRequest)
		return
	}

	// Get existing config to verify ownership
	existing, err := h.service.GetConfig(r.Context(), configID)
	if err != nil {
		http.Error(w, "Notification config not found", http.StatusNotFound)
		return
	}

	if existing.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delete configuration
	if err := h.service.DeleteConfig(r.Context(), configID); err != nil {
		log.Printf("Failed to delete notification config: %v", err)
		http.Error(w, "Failed to delete notification config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestConfig sends a test notification
func (h *Handler) TestConfig(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse config from request body
	var config NotificationConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID
	config.UserID = userID

	// Send test notification
	result := h.service.TestNotification(r.Context(), &config)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetProjectConfigs lists notification configurations for a specific project
func (h *Handler) GetProjectConfigs(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract project ID from path
	projectID, err := extractIDFromPath(r.URL.Path, "/api/v1/projects/")
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// TODO: Verify user owns this project

	configs, err := h.service.GetConfigsByProject(r.Context(), projectID)
	if err != nil {
		log.Printf("Failed to get project notification configs: %v", err)
		http.Error(w, "Failed to retrieve notification configs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}

// GetBuildLogs retrieves notification logs for a build
func (h *Handler) GetBuildLogs(w http.ResponseWriter, r *http.Request) {
	// Extract build ID from path
	buildID, err := extractIDFromPath(r.URL.Path, "/api/v1/builds/")
	if err != nil {
		http.Error(w, "Invalid build ID", http.StatusBadRequest)
		return
	}

	// TODO: Verify user has access to this build

	logs, err := h.service.GetNotificationLogs(r.Context(), buildID)
	if err != nil {
		log.Printf("Failed to get notification logs for build %d: %v", buildID, err)
		http.Error(w, "Failed to retrieve notification logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// GetSupportedTypes returns a list of supported notification types
func (h *Handler) GetSupportedTypes(w http.ResponseWriter, r *http.Request) {
	types := map[string]interface{}{
		"types": []map[string]string{
			{"id": string(NotificationTypeSlack), "name": "Slack", "category": "chat"},
			{"id": string(NotificationTypeDiscord), "name": "Discord", "category": "chat"},
			{"id": string(NotificationTypeTeams), "name": "Microsoft Teams", "category": "chat"},
			{"id": string(NotificationTypeTelegram), "name": "Telegram", "category": "chat"},
			{"id": string(NotificationTypeMatrix), "name": "Matrix", "category": "chat"},
			{"id": string(NotificationTypeIRC), "name": "IRC", "category": "chat"},
			{"id": string(NotificationTypeMattermost), "name": "Mattermost", "category": "chat"},
			{"id": string(NotificationTypeEmail), "name": "Email", "category": "email"},
			{"id": string(NotificationTypeJira), "name": "JIRA", "category": "issue_tracking"},
			{"id": string(NotificationTypeGithub), "name": "GitHub Issues", "category": "issue_tracking"},
			{"id": string(NotificationTypeGitlab), "name": "GitLab Issues", "category": "issue_tracking"},
			{"id": string(NotificationTypeLinear), "name": "Linear", "category": "issue_tracking"},
			{"id": string(NotificationTypeAsana), "name": "Asana", "category": "issue_tracking"},
			{"id": string(NotificationTypeGithubStatus), "name": "GitHub Status", "category": "git_status"},
			{"id": string(NotificationTypeGitlabStatus), "name": "GitLab Status", "category": "git_status"},
			{"id": string(NotificationTypeBitbucketStatus), "name": "Bitbucket Status", "category": "git_status"},
			{"id": string(NotificationTypeGiteaStatus), "name": "Gitea Status", "category": "git_status"},
			{"id": string(NotificationTypePagerDuty), "name": "PagerDuty", "category": "incident"},
			{"id": string(NotificationTypeOpsGenie), "name": "OpsGenie", "category": "incident"},
			{"id": string(NotificationTypeVictorOps), "name": "VictorOps", "category": "incident"},
			{"id": string(NotificationTypeWebhook), "name": "Webhook", "category": "custom"},
			{"id": string(NotificationTypeScript), "name": "Script", "category": "custom"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types)
}

// GetEventTypes returns a list of supported event types
func (h *Handler) GetEventTypes(w http.ResponseWriter, r *http.Request) {
	events := map[string]interface{}{
		"events": []map[string]string{
			{"id": string(EventBuildStarted), "name": "Build Started", "category": "build"},
			{"id": string(EventBuildSuccess), "name": "Build Success", "category": "build"},
			{"id": string(EventBuildFailure), "name": "Build Failure", "category": "build"},
			{"id": string(EventBuildCancelled), "name": "Build Cancelled", "category": "build"},
			{"id": string(EventDeployStarted), "name": "Deploy Started", "category": "deploy"},
			{"id": string(EventDeploySuccess), "name": "Deploy Success", "category": "deploy"},
			{"id": string(EventDeployFailure), "name": "Deploy Failure", "category": "deploy"},
			{"id": string(EventSecurityAlert), "name": "Security Alert", "category": "security"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// Helper functions

// extractIDFromPath extracts ID from URL path
func extractIDFromPath(path, prefix string) (int, error) {
	path = strings.TrimPrefix(path, prefix)
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 {
		return 0, strconv.ErrSyntax
	}

	return strconv.Atoi(parts[0])
}

// getUserIDFromContext extracts user ID from request context
// In production, this would be set by the auth middleware
func getUserIDFromContext(r *http.Request) int {
	// Check for user in context (set by auth middleware)
	if user := r.Context().Value("user"); user != nil {
		if u, ok := user.(map[string]interface{}); ok {
			if id, ok := u["id"].(int); ok {
				return id
			}
		}
	}

	// Fallback: check header (for testing)
	if userIDStr := r.Header.Get("X-User-ID"); userIDStr != "" {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			return userID
		}
	}

	return 0
}
