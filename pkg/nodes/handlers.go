package nodes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// Handler handles HTTP requests for nodes
type Handler struct {
	service *Service
}

// NewHandler creates a new node handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register handles node registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reg NodeRegistration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if reg.Hostname == "" || reg.Token == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	node, err := h.service.Register(r.Context(), &reg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(node)
}

// List handles listing all nodes
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodes, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

// Get handles getting a specific node
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id, err := extractIDFromPath(r.URL.Path, "/api/v1/nodes/")
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	node, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

// Update handles updating a node
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id, err := extractIDFromPath(r.URL.Path, "/api/v1/nodes/")
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	// Get existing node
	node, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Decode update
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Apply updates
	if hostname, ok := updates["hostname"].(string); ok {
		node.Hostname = hostname
	}
	if labels, ok := updates["labels"].(map[string]interface{}); ok {
		node.Labels = make(map[string]string)
		for k, v := range labels {
			if str, ok := v.(string); ok {
				node.Labels[k] = str
			}
		}
	}

	if err := h.service.Update(r.Context(), node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

// Delete handles deleting a node
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id, err := extractIDFromPath(r.URL.Path, "/api/v1/nodes/")
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Heartbeat handles node heartbeat
func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id, err := extractIDFromPath(r.URL.Path, "/api/v1/nodes/")
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateHeartbeat(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Drain handles draining a node
func (h *Handler) Drain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id, err := extractIDFromPath(r.URL.Path, "/api/v1/nodes/")
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Drain(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "draining"})
}

// GenerateToken handles generating a new node join token
func (h *Handler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse expiry from request (default 30 minutes)
	expiry := 30
	if expiryStr := r.URL.Query().Get("expiry"); expiryStr != "" {
		if e, err := strconv.Atoi(expiryStr); err == nil {
			expiry = e
		}
	}

	token, err := h.service.GenerateToken(r.Context(), expiry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      token.Token,
		"expires_at": token.ExpiresAt,
	})
}

// Helper function to extract ID from path
func extractIDFromPath(path, prefix string) (int, error) {
	// Remove prefix and any trailing slashes
	idStr := strings.TrimPrefix(path, prefix)
	idStr = strings.TrimSuffix(idStr, "/")

	// Remove any sub-paths (e.g., /nodes/123/heartbeat -> 123)
	parts := strings.Split(idStr, "/")
	if len(parts) > 0 {
		idStr = parts[0]
	}

	return strconv.Atoi(idStr)
}
