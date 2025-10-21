package metrics

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// PrometheusExporter exports metrics in Prometheus format
type PrometheusExporter struct {
	collector *Collector
	mu        sync.RWMutex
}

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(collector *Collector) *PrometheusExporter {
	return &PrometheusExporter{
		collector: collector,
	}
}

// Export returns metrics in Prometheus text format
func (pe *PrometheusExporter) Export() string {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	var sb strings.Builder

	// System metrics
	pe.writeSystemMetrics(&sb)

	// Build metrics
	pe.writeBuildMetrics(&sb)

	// Node metrics
	pe.writeNodeMetrics(&sb)

	// User metrics
	pe.writeUserMetrics(&sb)

	// Security metrics
	pe.writeSecurityMetrics(&sb)

	// API metrics
	pe.writeAPIMetrics(&sb)

	return sb.String()
}

func (pe *PrometheusExporter) writeSystemMetrics(sb *strings.Builder) {
	metrics := pe.collector.GetSystemMetrics()

	writeMetric(sb, "casci_system_cpu_usage_percent", "gauge", "CPU usage percentage", metrics.CPUUsagePercent)
	writeMetric(sb, "casci_system_memory_usage_bytes", "gauge", "Memory usage in bytes", float64(metrics.MemoryUsageBytes))
	writeMetric(sb, "casci_system_memory_total_bytes", "gauge", "Total memory in bytes", float64(metrics.MemoryTotalBytes))
	writeMetric(sb, "casci_system_disk_usage_bytes", "gauge", "Disk usage in bytes", float64(metrics.DiskUsageBytes))
	writeMetric(sb, "casci_system_disk_total_bytes", "gauge", "Total disk space in bytes", float64(metrics.DiskTotalBytes))
	writeMetric(sb, "casci_system_network_bytes_in", "counter", "Network bytes received", float64(metrics.NetworkBytesIn))
	writeMetric(sb, "casci_system_network_bytes_out", "counter", "Network bytes sent", float64(metrics.NetworkBytesOut))
	writeMetric(sb, "casci_system_container_count", "gauge", "Number of containers", float64(metrics.ContainerCount))
	writeMetric(sb, "casci_system_goroutine_count", "gauge", "Number of goroutines", float64(metrics.GoroutineCount))
	writeMetric(sb, "casci_system_uptime_seconds", "counter", "System uptime in seconds", float64(metrics.UptimeSeconds))
}

func (pe *PrometheusExporter) writeBuildMetrics(sb *strings.Builder) {
	metrics := pe.collector.GetBuildMetrics()

	writeMetric(sb, "casci_builds_total", "counter", "Total number of builds", float64(metrics.BuildsTotal))
	writeMetric(sb, "casci_builds_queued", "gauge", "Number of queued builds", float64(metrics.BuildsQueued))
	writeMetric(sb, "casci_builds_running", "gauge", "Number of running builds", float64(metrics.BuildsRunning))
	writeMetric(sb, "casci_builds_success", "counter", "Number of successful builds", float64(metrics.BuildsSuccess))
	writeMetric(sb, "casci_builds_failed", "counter", "Number of failed builds", float64(metrics.BuildsFailed))
	writeMetric(sb, "casci_builds_cancelled", "counter", "Number of cancelled builds", float64(metrics.BuildsCancelled))

	// Average build duration
	avgDuration := pe.collector.GetAverageBuildDuration()
	writeMetric(sb, "casci_build_duration_seconds_avg", "gauge", "Average build duration in seconds", avgDuration)

	// Average queue time
	avgQueue := pe.collector.GetAverageQueueTime()
	writeMetric(sb, "casci_build_queue_time_seconds_avg", "gauge", "Average queue time in seconds", avgQueue)

	// Success rate
	successRate := pe.collector.GetSuccessRate()
	writeMetric(sb, "casci_build_success_rate", "gauge", "Build success rate percentage", successRate)
}

func (pe *PrometheusExporter) writeNodeMetrics(sb *strings.Builder) {
	metrics := pe.collector.GetNodeMetrics()

	writeMetric(sb, "casci_nodes_total", "gauge", "Total number of nodes", float64(metrics.NodesTotal))
	writeMetric(sb, "casci_nodes_online", "gauge", "Number of online nodes", float64(metrics.NodesOnline))
	writeMetric(sb, "casci_nodes_offline", "gauge", "Number of offline nodes", float64(metrics.NodesOffline))
	writeMetric(sb, "casci_nodes_draining", "gauge", "Number of draining nodes", float64(metrics.NodesDraining))
	writeMetric(sb, "casci_node_capacity_cpu", "gauge", "Total node CPU capacity", metrics.NodeCapacityCPU)
	writeMetric(sb, "casci_node_capacity_memory", "gauge", "Total node memory capacity", metrics.NodeCapacityMemory)
	writeMetric(sb, "casci_node_usage_cpu", "gauge", "Total node CPU usage", metrics.NodeUsageCPU)
	writeMetric(sb, "casci_node_usage_memory", "gauge", "Total node memory usage", metrics.NodeUsageMemory)
}

func (pe *PrometheusExporter) writeUserMetrics(sb *strings.Builder) {
	metrics := pe.collector.GetUserMetrics()

	writeMetric(sb, "casci_users_total", "gauge", "Total number of users", float64(metrics.UsersTotal))
	writeMetric(sb, "casci_users_active", "gauge", "Number of active users", float64(metrics.UsersActive))
	writeMetric(sb, "casci_api_requests_total", "counter", "Total API requests", float64(metrics.APIRequests))

	// Per-user metrics
	for userID, count := range metrics.APIRequestsByUser {
		writeMetricWithLabels(sb, "casci_api_requests_by_user", "counter", "API requests by user",
			map[string]string{"user_id": fmt.Sprintf("%d", userID)}, float64(count))
	}

	for userID, usage := range metrics.ResourceUsageByUser {
		labels := map[string]string{"user_id": fmt.Sprintf("%d", userID)}
		writeMetricWithLabels(sb, "casci_user_builds_total", "counter", "Total builds by user",
			labels, float64(usage.BuildsTotal))
		writeMetricWithLabels(sb, "casci_user_cpu_time_seconds", "counter", "CPU time used by user",
			labels, usage.CPUTimeSeconds)
		writeMetricWithLabels(sb, "casci_user_storage_bytes", "gauge", "Storage used by user",
			labels, float64(usage.StorageBytes))
	}
}

func (pe *PrometheusExporter) writeSecurityMetrics(sb *strings.Builder) {
	metrics := pe.collector.GetSecurityMetrics()

	writeMetric(sb, "casci_security_scans_total", "counter", "Total security scans", float64(metrics.ScansTotal))
	writeMetric(sb, "casci_security_vulnerabilities_total", "gauge", "Total vulnerabilities found", float64(metrics.VulnerabilitiesTotal))
	writeMetric(sb, "casci_security_vulnerabilities_critical", "gauge", "Critical vulnerabilities", float64(metrics.VulnerabilitiesCritical))
	writeMetric(sb, "casci_security_vulnerabilities_high", "gauge", "High vulnerabilities", float64(metrics.VulnerabilitiesHigh))
	writeMetric(sb, "casci_security_vulnerabilities_medium", "gauge", "Medium vulnerabilities", float64(metrics.VulnerabilitiesMedium))
	writeMetric(sb, "casci_security_vulnerabilities_low", "gauge", "Low vulnerabilities", float64(metrics.VulnerabilitiesLow))
	writeMetric(sb, "casci_security_secrets_found", "gauge", "Secrets found in scans", float64(metrics.SecretsFound))
	writeMetric(sb, "casci_security_license_issues", "gauge", "License compliance issues", float64(metrics.LicenseIssues))
	writeMetric(sb, "casci_security_code_issues", "gauge", "Code quality issues", float64(metrics.CodeIssues))
}

func (pe *PrometheusExporter) writeAPIMetrics(sb *strings.Builder) {
	metrics := pe.collector.GetAPIMetrics()

	writeMetric(sb, "casci_api_requests_total", "counter", "Total API requests", float64(metrics.RequestsTotal))
	writeMetric(sb, "casci_api_errors_total", "counter", "Total API errors", float64(metrics.ErrorsTotal))

	// Average API latency
	avgLatency := pe.collector.GetAverageAPILatency()
	writeMetric(sb, "casci_api_latency_seconds_avg", "gauge", "Average API latency in seconds", avgLatency)

	// Per-endpoint metrics
	for endpoint, count := range metrics.RequestsByEndpoint {
		writeMetricWithLabels(sb, "casci_api_requests_by_endpoint", "counter", "API requests by endpoint",
			map[string]string{"endpoint": endpoint}, float64(count))
	}

	// Per-status code metrics
	for status, count := range metrics.RequestsByStatus {
		writeMetricWithLabels(sb, "casci_api_requests_by_status", "counter", "API requests by status code",
			map[string]string{"status": fmt.Sprintf("%d", status)}, float64(count))
	}
}

// Helper functions

func writeMetric(sb *strings.Builder, name, metricType, help string, value float64) {
	sb.WriteString(fmt.Sprintf("# HELP %s %s\n", name, help))
	sb.WriteString(fmt.Sprintf("# TYPE %s %s\n", name, metricType))
	sb.WriteString(fmt.Sprintf("%s %v\n", name, value))
	sb.WriteString("\n")
}

func writeMetricWithLabels(sb *strings.Builder, name, metricType, help string, labels map[string]string, value float64) {
	// Write HELP and TYPE only once per metric name
	// In production, we'd track which metrics we've already written headers for
	labelStr := formatLabels(labels)
	sb.WriteString(fmt.Sprintf("%s{%s} %v\n", name, labelStr, value))
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=\"%s\"", k, escapeLabel(v)))
	}

	return strings.Join(parts, ",")
}

func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

// MetricsSnapshot represents a point-in-time snapshot of all metrics
type MetricsSnapshot struct {
	Timestamp       time.Time
	SystemMetrics   SystemMetrics
	BuildMetrics    BuildMetrics
	NodeMetrics     NodeMetrics
	UserMetrics     UserMetrics
	SecurityMetrics SecurityMetrics
	APIMetrics      APIMetrics
}

// GetSnapshot returns a snapshot of all current metrics
func (pe *PrometheusExporter) GetSnapshot() *MetricsSnapshot {
	return &MetricsSnapshot{
		Timestamp:       time.Now(),
		SystemMetrics:   pe.collector.GetSystemMetrics(),
		BuildMetrics:    pe.collector.GetBuildMetrics(),
		NodeMetrics:     pe.collector.GetNodeMetrics(),
		UserMetrics:     pe.collector.GetUserMetrics(),
		SecurityMetrics: pe.collector.GetSecurityMetrics(),
		APIMetrics:      pe.collector.GetAPIMetrics(),
	}
}
