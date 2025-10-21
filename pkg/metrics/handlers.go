package metrics

import (
	"encoding/json"
	"net/http"
)

// Handler handles HTTP requests for metrics
type Handler struct {
	exporter *PrometheusExporter
}

// NewHandler creates a new metrics handler
func NewHandler(exporter *PrometheusExporter) *Handler {
	return &Handler{
		exporter: exporter,
	}
}

// ServeMetrics serves Prometheus-formatted metrics
func (h *Handler) ServeMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	metrics := h.exporter.Export()
	w.Write([]byte(metrics))
}

// ServeMetricsJSON serves metrics in JSON format
func (h *Handler) ServeMetricsJSON(w http.ResponseWriter, r *http.Request) {
	snapshot := h.exporter.GetSnapshot()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

// ServeHealth serves health check endpoint
func (h *Handler) ServeHealth(w http.ResponseWriter, r *http.Request) {
	health := h.getHealthStatus()
	w.Header().Set("Content-Type", "application/json")

	if health["status"] == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(health)
}

// ServeReadiness serves readiness check endpoint
func (h *Handler) ServeReadiness(w http.ResponseWriter, r *http.Request) {
	ready := h.getReadinessStatus()
	w.Header().Set("Content-Type", "application/json")

	if ready["ready"] == true {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(ready)
}

// ServeLiveness serves liveness check endpoint
func (h *Handler) ServeLiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alive": true,
	})
}

func (h *Handler) getHealthStatus() map[string]interface{} {
	snapshot := h.exporter.GetSnapshot()

	// Check system health
	systemHealthy := true
	if snapshot.SystemMetrics.CPUUsagePercent > 95 {
		systemHealthy = false
	}
	if snapshot.SystemMetrics.MemoryUsageBytes > 0 &&
		snapshot.SystemMetrics.MemoryTotalBytes > 0 {
		memPercent := float64(snapshot.SystemMetrics.MemoryUsageBytes) /
			float64(snapshot.SystemMetrics.MemoryTotalBytes) * 100
		if memPercent > 95 {
			systemHealthy = false
		}
	}
	if snapshot.SystemMetrics.DiskUsageBytes > 0 &&
		snapshot.SystemMetrics.DiskTotalBytes > 0 {
		diskPercent := float64(snapshot.SystemMetrics.DiskUsageBytes) /
			float64(snapshot.SystemMetrics.DiskTotalBytes) * 100
		if diskPercent > 95 {
			systemHealthy = false
		}
	}

	// Check if nodes are available
	nodesHealthy := snapshot.NodeMetrics.NodesOnline > 0

	// Overall health
	status := "healthy"
	if !systemHealthy || !nodesHealthy {
		status = "degraded"
	}

	return map[string]interface{}{
		"status": status,
		"checks": map[string]interface{}{
			"system": systemHealthy,
			"nodes":  nodesHealthy,
		},
		"metrics": map[string]interface{}{
			"cpu_usage_percent":  snapshot.SystemMetrics.CPUUsagePercent,
			"memory_usage_bytes": snapshot.SystemMetrics.MemoryUsageBytes,
			"disk_usage_bytes":   snapshot.SystemMetrics.DiskUsageBytes,
			"nodes_online":       snapshot.NodeMetrics.NodesOnline,
			"builds_running":     snapshot.BuildMetrics.BuildsRunning,
			"builds_queued":      snapshot.BuildMetrics.BuildsQueued,
		},
	}
}

func (h *Handler) getReadinessStatus() map[string]interface{} {
	snapshot := h.exporter.GetSnapshot()

	// System is ready if:
	// 1. At least one node is online
	// 2. System is not overloaded
	// 3. Can accept new builds

	ready := true
	reasons := []string{}

	if snapshot.NodeMetrics.NodesOnline == 0 {
		ready = false
		reasons = append(reasons, "no nodes online")
	}

	if snapshot.SystemMetrics.CPUUsagePercent > 98 {
		ready = false
		reasons = append(reasons, "cpu overloaded")
	}

	if snapshot.SystemMetrics.MemoryUsageBytes > 0 &&
		snapshot.SystemMetrics.MemoryTotalBytes > 0 {
		memPercent := float64(snapshot.SystemMetrics.MemoryUsageBytes) /
			float64(snapshot.SystemMetrics.MemoryTotalBytes) * 100
		if memPercent > 98 {
			ready = false
			reasons = append(reasons, "memory overloaded")
		}
	}

	result := map[string]interface{}{
		"ready": ready,
	}

	if !ready {
		result["reasons"] = reasons
	}

	return result
}

// ServeSystemMetrics serves detailed system metrics
func (h *Handler) ServeSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.exporter.collector.GetSystemMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// ServeBuildMetrics serves detailed build metrics
func (h *Handler) ServeBuildMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.exporter.collector.GetBuildMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"builds":                metrics,
		"average_duration":      h.exporter.collector.GetAverageBuildDuration(),
		"average_queue_time":    h.exporter.collector.GetAverageQueueTime(),
		"success_rate_percent":  h.exporter.collector.GetSuccessRate(),
	})
}

// ServeNodeMetrics serves detailed node metrics
func (h *Handler) ServeNodeMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.exporter.collector.GetNodeMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// ServeSecurityMetrics serves detailed security metrics
func (h *Handler) ServeSecurityMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.exporter.collector.GetSecurityMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// ServeAPIMetrics serves detailed API metrics
func (h *Handler) ServeAPIMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.exporter.collector.GetAPIMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"api":             metrics,
		"average_latency": h.exporter.collector.GetAverageAPILatency(),
	})
}
