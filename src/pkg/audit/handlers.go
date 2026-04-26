package audit

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	filter := &AuditFilter{
		Limit: 100,
	}
	
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			filter.UserID = &userID
		}
	}
	
	if action := r.URL.Query().Get("action"); action != "" {
		filter.Action = action
	}
	
	if resource := r.URL.Query().Get("resource"); resource != "" {
		filter.Resource = resource
	}
	
	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}
	
	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}
	
	if successStr := r.URL.Query().Get("success"); successStr != "" {
		if success, err := strconv.ParseBool(successStr); err == nil {
			filter.Success = &success
		}
	}
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}
	
	events, err := h.service.Query(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	count, _ := h.service.Count(ctx, filter)
	
	response := map[string]interface{}{
		"events": events,
		"total":  count,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) GetStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"enabled":   h.service.IsEnabled(),
		"retention": h.service.GetRetention().String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled   *bool   `json:"enabled"`
		Retention *string `json:"retention"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Enabled != nil {
		h.service.SetEnabled(*req.Enabled)
	}
	
	if req.Retention != nil {
		if duration, err := time.ParseDuration(*req.Retention); err == nil {
			h.service.SetRetention(duration)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *Handlers) TriggerCleanup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	count, err := h.service.Cleanup(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"deleted": count,
		"message": "Cleanup completed successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
