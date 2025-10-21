package projects

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/casapps/casci/pkg/users"
)

// Handler handles HTTP requests for projects
type Handler struct {
	service *Service
}

// NewHandler creates a new project handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Create handles project creation
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	project, err := h.service.Create(r.Context(), user.ID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrProjectNameExists {
			status = http.StatusConflict
		} else if err == ErrInvalidProjectName || err == ErrProjectNameTooShort || err == ErrInvalidRepositoryURL {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

// Get handles getting a project by ID
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

	// Extract ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/projects/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
		return
	}

	project, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == ErrProjectNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Project not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Check authorization
	if project.UserID != user.ID && !user.IsAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	writeJSON(w, http.StatusOK, project)
}

// List handles listing projects
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Admins can see all projects with ?all=true
	if user.IsAdmin && r.URL.Query().Get("all") == "true" {
		projects, err := h.service.ListAll(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, projects)
		return
	}

	// Regular users see only their projects
	projects, err := h.service.ListByUser(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, projects)
}

// Update handles updating a project
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/projects/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
		return
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	project, err := h.service.Update(r.Context(), id, user.ID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrProjectNotFound {
			status = http.StatusNotFound
		} else if err == ErrNotProjectOwner || err == ErrUnauthorizedAccess {
			status = http.StatusForbidden
		} else if err == ErrProjectNameExists || err == ErrProjectNameTooShort {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, project)
}

// Delete handles deleting a project
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := users.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/projects/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
		return
	}

	if err := h.service.Delete(r.Context(), id, user.ID); err != nil {
		status := http.StatusInternalServerError
		if err == ErrProjectNotFound {
			status = http.StatusNotFound
		} else if err == ErrNotProjectOwner || err == ErrUnauthorizedAccess {
			status = http.StatusForbidden
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to write JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
