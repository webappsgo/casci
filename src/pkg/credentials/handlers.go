package credentials

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Handler handles HTTP requests for credential operations
type Handler struct {
	service *Service
}

// NewHandler creates a new credential handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// User Credentials

// ListUserCredentials lists all credentials for the authenticated user
func (h *Handler) ListUserCredentials(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	creds, err := h.service.ListUserCredentials(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to list user credentials: %v", err)
		http.Error(w, "Failed to retrieve credentials", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creds)
}

// GetUserCredential retrieves a specific user credential
func (h *Handler) GetUserCredential(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	credID, err := extractIDFromPath(r.URL.Path, "/api/v1/credentials/user/")
	if err != nil {
		http.Error(w, "Invalid credential ID", http.StatusBadRequest)
		return
	}

	cred, err := h.service.GetUserCredential(r.Context(), credID, userID)
	if err != nil {
		if err == ErrCredentialNotFound {
			http.Error(w, "Credential not found", http.StatusNotFound)
			return
		}
		if err == ErrUnauthorizedAccess {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		log.Printf("Failed to get user credential: %v", err)
		http.Error(w, "Failed to retrieve credential", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cred)
}

// CreateUserCredential creates a new user credential
func (h *Handler) CreateUserCredential(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateUserCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cred, err := h.service.CreateUserCredential(r.Context(), userID, &req)
	if err != nil {
		log.Printf("Failed to create user credential: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cred)
}

// UpdateUserCredential updates a user credential
func (h *Handler) UpdateUserCredential(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	credID, err := extractIDFromPath(r.URL.Path, "/api/v1/credentials/user/")
	if err != nil {
		http.Error(w, "Invalid credential ID", http.StatusBadRequest)
		return
	}

	var req UpdateUserCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cred, err := h.service.UpdateUserCredential(r.Context(), credID, userID, &req)
	if err != nil {
		if err == ErrCredentialNotFound {
			http.Error(w, "Credential not found", http.StatusNotFound)
			return
		}
		if err == ErrUnauthorizedAccess {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		log.Printf("Failed to update user credential: %v", err)
		http.Error(w, "Failed to update credential", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cred)
}

// DeleteUserCredential deletes a user credential
func (h *Handler) DeleteUserCredential(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	credID, err := extractIDFromPath(r.URL.Path, "/api/v1/credentials/user/")
	if err != nil {
		http.Error(w, "Invalid credential ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteUserCredential(r.Context(), credID, userID)
	if err != nil {
		if err == ErrCredentialNotFound {
			http.Error(w, "Credential not found", http.StatusNotFound)
			return
		}
		if err == ErrUnauthorizedAccess {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		log.Printf("Failed to delete user credential: %v", err)
		http.Error(w, "Failed to delete credential", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Project Credentials

// ListProjectCredentials lists all credentials for a project
func (h *Handler) ListProjectCredentials(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	projectID, err := extractIDFromPath(r.URL.Path, "/api/v1/projects/")
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	creds, err := h.service.ListProjectCredentials(r.Context(), projectID, userID)
	if err != nil {
		log.Printf("Failed to list project credentials: %v", err)
		http.Error(w, "Failed to retrieve credentials", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creds)
}

// GetProjectCredential retrieves a specific project credential
func (h *Handler) GetProjectCredential(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	credID, err := extractIDFromPath(r.URL.Path, "/api/v1/credentials/project/")
	if err != nil {
		http.Error(w, "Invalid credential ID", http.StatusBadRequest)
		return
	}

	cred, err := h.service.GetProjectCredential(r.Context(), credID, userID)
	if err != nil {
		if err == ErrCredentialNotFound {
			http.Error(w, "Credential not found", http.StatusNotFound)
			return
		}
		if err == ErrUnauthorizedAccess {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		log.Printf("Failed to get project credential: %v", err)
		http.Error(w, "Failed to retrieve credential", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cred)
}

// CreateProjectCredential creates a new project credential
func (h *Handler) CreateProjectCredential(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	projectID, err := extractIDFromPath(r.URL.Path, "/api/v1/projects/")
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req CreateProjectCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cred, err := h.service.CreateProjectCredential(r.Context(), projectID, userID, &req)
	if err != nil {
		log.Printf("Failed to create project credential: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cred)
}

// UpdateProjectCredential updates a project credential
func (h *Handler) UpdateProjectCredential(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	credID, err := extractIDFromPath(r.URL.Path, "/api/v1/credentials/project/")
	if err != nil {
		http.Error(w, "Invalid credential ID", http.StatusBadRequest)
		return
	}

	var req UpdateProjectCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cred, err := h.service.UpdateProjectCredential(r.Context(), credID, userID, &req)
	if err != nil {
		if err == ErrCredentialNotFound {
			http.Error(w, "Credential not found", http.StatusNotFound)
			return
		}
		if err == ErrUnauthorizedAccess {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		log.Printf("Failed to update project credential: %v", err)
		http.Error(w, "Failed to update credential", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cred)
}

// DeleteProjectCredential deletes a project credential
func (h *Handler) DeleteProjectCredential(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	credID, err := extractIDFromPath(r.URL.Path, "/api/v1/credentials/project/")
	if err != nil {
		http.Error(w, "Invalid credential ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteProjectCredential(r.Context(), credID, userID)
	if err != nil {
		if err == ErrCredentialNotFound {
			http.Error(w, "Credential not found", http.StatusNotFound)
			return
		}
		if err == ErrUnauthorizedAccess {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		log.Printf("Failed to delete project credential: %v", err)
		http.Error(w, "Failed to delete credential", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
