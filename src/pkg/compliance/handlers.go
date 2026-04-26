package compliance

import (
	"encoding/json"
	"net/http"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.service.GetConfig()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (h *Handlers) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var config ComplianceConfig
	
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.service.SetConfig(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *Handlers) SetMode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mode Mode `json:"mode"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.service.SetMode(req.Mode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	config := h.service.GetConfig()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (h *Handlers) RunCompliance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	report, err := h.service.RunCompliance(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *Handlers) ListModes(w http.ResponseWriter, r *http.Request) {
	modes := []struct {
		Mode        Mode   `json:"mode"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}{
		{
			Mode:        ModeNone,
			Name:        "None",
			Description: "No compliance mode enabled",
		},
		{
			Mode:        ModeHIPAA,
			Name:        "HIPAA",
			Description: "Health Insurance Portability and Accountability Act",
		},
		{
			Mode:        ModeSOX,
			Name:        "SOX",
			Description: "Sarbanes-Oxley Act",
		},
		{
			Mode:        ModePCIDSS,
			Name:        "PCI-DSS",
			Description: "Payment Card Industry Data Security Standard",
		},
		{
			Mode:        ModeGDPR,
			Name:        "GDPR",
			Description: "General Data Protection Regulation",
		},
		{
			Mode:        ModeFedRAMP,
			Name:        "FedRAMP",
			Description: "Federal Risk and Authorization Management Program",
		},
		{
			Mode:        ModeISO27001,
			Name:        "ISO 27001",
			Description: "Information Security Management System",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modes)
}

func (h *Handlers) GetPreset(w http.ResponseWriter, r *http.Request) {
	modeStr := r.URL.Query().Get("mode")
	if modeStr == "" {
		http.Error(w, "mode parameter required", http.StatusBadRequest)
		return
	}
	
	mode := Mode(modeStr)
	config := GetPresetConfig(mode)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}
