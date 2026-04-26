package builds

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/casapps/casci/src/pkg/projects"
	"github.com/casapps/casci/src/pkg/users"
)

// Handler handles HTTP requests for builds
type Handler struct {
	service        *Service
	projectService *projects.Service
}

// NewHandler creates a new build handler
func NewHandler(service *Service, projectService *projects.Service) *Handler {
	return &Handler{
		service:        service,
		projectService: projectService,
	}
}

// Trigger handles triggering a new build
func (h *Handler) Trigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract project ID from path: /api/v1/projects/{id}/builds
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid URL"})
		return
	}

	projectID, err := strconv.Atoi(parts[3])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
		return
	}

	// Check project access
	if err := h.projectService.CheckAccess(r.Context(), projectID, user.ID); err != nil {
		if err == projects.ErrProjectNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Project not found"})
			return
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get project to populate build details
	project, err := h.projectService.GetByID(r.Context(), projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get project"})
		return
	}

	var req TriggerBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is OK, use defaults
		req = TriggerBuildRequest{
			Trigger: TriggerManual,
		}
	}

	// Use project branch if not specified
	if req.Branch == "" {
		req.Branch = project.Branch
	}

	build, err := h.service.Trigger(r.Context(), projectID, &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Populate build with project details
	build.RepositoryURL = project.RepositoryURL

	// Update the build with the repository URL
	if err := h.service.UpdateCommit(r.Context(), build.ID, build.CommitSHA, build.CommitMessage, build.CommitAuthor); err != nil {
		// Non-fatal error, just log it
		fmt.Printf("Warning: failed to update build: %v\n", err)
	}

	writeJSON(w, http.StatusCreated, build)
}

// Get handles getting a build by ID
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract build ID from path: /api/v1/builds/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/builds/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid build ID"})
		return
	}

	build, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == ErrBuildNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Build not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Check project access
	if err := h.projectService.CheckAccess(r.Context(), build.ProjectID, user.ID); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	writeJSON(w, http.StatusOK, build)
}

// ListByProject handles listing builds for a project
func (h *Handler) ListByProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract project ID from path: /api/v1/projects/{id}/builds
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid URL"})
		return
	}

	projectID, err := strconv.Atoi(parts[3])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
		return
	}

	// Check project access
	if err := h.projectService.CheckAccess(r.Context(), projectID, user.ID); err != nil {
		if err == projects.ErrProjectNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Project not found"})
			return
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get limit from query parameter
	limit := 100 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	builds, err := h.service.ListByProject(r.Context(), projectID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, builds)
}

// GetLog handles getting build logs
func (h *Handler) GetLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract build ID from path: /api/v1/builds/{id}/log
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid URL"})
		return
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid build ID"})
		return
	}

	build, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == ErrBuildNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Build not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Check project access
	if err := h.projectService.CheckAccess(r.Context(), build.ProjectID, user.ID); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Read log file
	logPath, err := h.service.GetLogPath(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	logContent, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, "No logs available yet")
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to read log file"})
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(logContent)
}

// Cancel handles canceling a build
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract build ID from path: /api/v1/builds/{id}/cancel
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid URL"})
		return
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid build ID"})
		return
	}

	build, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == ErrBuildNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Build not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Check project access
	if err := h.projectService.CheckAccess(r.Context(), build.ProjectID, user.ID); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.service.Cancel(r.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if err == ErrCannotCancelBuild {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Build canceled"})
}

// Restart handles restarting a build
func (h *Handler) Restart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract build ID from path: /api/v1/builds/{id}/restart
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid URL"})
		return
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid build ID"})
		return
	}

	build, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == ErrBuildNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Build not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Check project access
	if err := h.projectService.CheckAccess(r.Context(), build.ProjectID, user.ID); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	newBuild, err := h.service.Restart(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, newBuild)
}

// GetStats handles getting build statistics for a project
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract project ID from path: /api/v1/projects/{id}/builds/stats
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid URL"})
		return
	}

	projectID, err := strconv.Atoi(parts[3])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
		return
	}

	// Check project access
	if err := h.projectService.CheckAccess(r.Context(), projectID, user.ID); err != nil {
		if err == projects.ErrProjectNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Project not found"})
			return
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	stats, err := h.service.GetStats(r.Context(), projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// Helper function to write JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
