package metrics

import (
	"runtime"
	"sync"
	"time"
)

// Collector collects metrics for the CASCI system
type Collector struct {
	// System metrics
	systemMetrics *SystemMetrics

	// Build metrics
	buildMetrics *BuildMetrics

	// Node metrics
	nodeMetrics *NodeMetrics

	// User metrics
	userMetrics *UserMetrics

	// Security metrics
	securityMetrics *SecurityMetrics

	// API metrics
	apiMetrics *APIMetrics

	mu sync.RWMutex
}

// SystemMetrics tracks system-level metrics
type SystemMetrics struct {
	CPUUsagePercent    float64
	MemoryUsageBytes   uint64
	MemoryTotalBytes   uint64
	DiskUsageBytes     uint64
	DiskTotalBytes     uint64
	NetworkBytesIn     uint64
	NetworkBytesOut    uint64
	ContainerCount     int
	GoroutineCount     int
	UptimeSeconds      int64
	LastUpdated        time.Time
}

// BuildMetrics tracks build-related metrics
type BuildMetrics struct {
	BuildsTotal       int64
	BuildsQueued      int64
	BuildsRunning     int64
	BuildsSuccess     int64
	BuildsFailed      int64
	BuildsCancelled   int64
	BuildDurationSum  float64
	BuildDurationCount int64
	QueueTimeSum      float64
	QueueTimeCount    int64
	LastUpdated       time.Time
}

// NodeMetrics tracks node-related metrics
type NodeMetrics struct {
	NodesTotal          int64
	NodesOnline         int64
	NodesOffline        int64
	NodesDraining       int64
	NodeCapacityCPU     float64
	NodeCapacityMemory  float64
	NodeUsageCPU        float64
	NodeUsageMemory     float64
	LastUpdated         time.Time
}

// UserMetrics tracks user-related metrics
type UserMetrics struct {
	UsersTotal         int64
	UsersActive        int64
	APIRequests        int64
	APIRequestsByUser  map[int]int64
	ResourceUsageByUser map[int]*ResourceUsage
	LastUpdated        time.Time
}

// ResourceUsage tracks resource usage per user
type ResourceUsage struct {
	BuildsTotal    int64
	CPUTimeSeconds float64
	StorageBytes   uint64
}

// SecurityMetrics tracks security-related metrics
type SecurityMetrics struct {
	ScansTotal              int64
	VulnerabilitiesTotal    int64
	VulnerabilitiesCritical int64
	VulnerabilitiesHigh     int64
	VulnerabilitiesMedium   int64
	VulnerabilitiesLow      int64
	SecretsFound            int64
	LicenseIssues           int64
	CodeIssues              int64
	LastUpdated             time.Time
}

// APIMetrics tracks API request metrics
type APIMetrics struct {
	RequestsTotal       int64
	RequestsByEndpoint  map[string]int64
	RequestsByStatus    map[int]int64
	RequestDurationSum  float64
	RequestDurationCount int64
	ErrorsTotal         int64
	LastUpdated         time.Time
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	c := &Collector{
		systemMetrics: &SystemMetrics{
			LastUpdated: time.Now(),
		},
		buildMetrics: &BuildMetrics{
			LastUpdated: time.Now(),
		},
		nodeMetrics: &NodeMetrics{
			LastUpdated: time.Now(),
		},
		userMetrics: &UserMetrics{
			APIRequestsByUser:   make(map[int]int64),
			ResourceUsageByUser: make(map[int]*ResourceUsage),
			LastUpdated:         time.Now(),
		},
		securityMetrics: &SecurityMetrics{
			LastUpdated: time.Now(),
		},
		apiMetrics: &APIMetrics{
			RequestsByEndpoint: make(map[string]int64),
			RequestsByStatus:   make(map[int]int64),
			LastUpdated:        time.Now(),
		},
	}

	// Start background collection
	go c.collectSystemMetrics()

	return c
}

// RecordBuildStarted records a build start event
func (c *Collector) RecordBuildStarted() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buildMetrics.BuildsTotal++
	c.buildMetrics.BuildsQueued--
	c.buildMetrics.BuildsRunning++
	c.buildMetrics.LastUpdated = time.Now()
}

// RecordBuildQueued records a build being queued
func (c *Collector) RecordBuildQueued() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buildMetrics.BuildsQueued++
	c.buildMetrics.LastUpdated = time.Now()
}

// RecordBuildCompleted records a build completion
func (c *Collector) RecordBuildCompleted(success bool, duration time.Duration, queueTime time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buildMetrics.BuildsRunning--

	if success {
		c.buildMetrics.BuildsSuccess++
	} else {
		c.buildMetrics.BuildsFailed++
	}

	// Record duration
	c.buildMetrics.BuildDurationSum += duration.Seconds()
	c.buildMetrics.BuildDurationCount++

	// Record queue time
	c.buildMetrics.QueueTimeSum += queueTime.Seconds()
	c.buildMetrics.QueueTimeCount++

	c.buildMetrics.LastUpdated = time.Now()
}

// RecordBuildCancelled records a build cancellation
func (c *Collector) RecordBuildCancelled() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buildMetrics.BuildsRunning--
	c.buildMetrics.BuildsCancelled++
	c.buildMetrics.LastUpdated = time.Now()
}

// RecordNodeStatus records node status
func (c *Collector) RecordNodeStatus(total, online, offline, draining int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nodeMetrics.NodesTotal = total
	c.nodeMetrics.NodesOnline = online
	c.nodeMetrics.NodesOffline = offline
	c.nodeMetrics.NodesDraining = draining
	c.nodeMetrics.LastUpdated = time.Now()
}

// RecordNodeCapacity records node capacity
func (c *Collector) RecordNodeCapacity(cpu, memory float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nodeMetrics.NodeCapacityCPU = cpu
	c.nodeMetrics.NodeCapacityMemory = memory
	c.nodeMetrics.LastUpdated = time.Now()
}

// RecordNodeUsage records node usage
func (c *Collector) RecordNodeUsage(cpu, memory float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nodeMetrics.NodeUsageCPU = cpu
	c.nodeMetrics.NodeUsageMemory = memory
	c.nodeMetrics.LastUpdated = time.Now()
}

// RecordAPIRequest records an API request
func (c *Collector) RecordAPIRequest(endpoint string, statusCode int, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.apiMetrics.RequestsTotal++
	c.apiMetrics.RequestsByEndpoint[endpoint]++
	c.apiMetrics.RequestsByStatus[statusCode]++
	c.apiMetrics.RequestDurationSum += duration.Seconds()
	c.apiMetrics.RequestDurationCount++

	if statusCode >= 400 {
		c.apiMetrics.ErrorsTotal++
	}

	c.apiMetrics.LastUpdated = time.Now()
}

// RecordUserActivity records user activity
func (c *Collector) RecordUserActivity(userID int, builds int64, cpuTime float64, storage uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.userMetrics.ResourceUsageByUser[userID]; !exists {
		c.userMetrics.ResourceUsageByUser[userID] = &ResourceUsage{}
	}

	usage := c.userMetrics.ResourceUsageByUser[userID]
	usage.BuildsTotal = builds
	usage.CPUTimeSeconds = cpuTime
	usage.StorageBytes = storage

	c.userMetrics.LastUpdated = time.Now()
}

// RecordSecurityScan records security scan results
func (c *Collector) RecordSecurityScan(critical, high, medium, low int64, secrets, licenses, codeIssues int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.securityMetrics.ScansTotal++
	c.securityMetrics.VulnerabilitiesCritical += critical
	c.securityMetrics.VulnerabilitiesHigh += high
	c.securityMetrics.VulnerabilitiesMedium += medium
	c.securityMetrics.VulnerabilitiesLow += low
	c.securityMetrics.VulnerabilitiesTotal += critical + high + medium + low
	c.securityMetrics.SecretsFound += secrets
	c.securityMetrics.LicenseIssues += licenses
	c.securityMetrics.CodeIssues += codeIssues
	c.securityMetrics.LastUpdated = time.Now()
}

// GetSystemMetrics returns current system metrics
func (c *Collector) GetSystemMetrics() SystemMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return *c.systemMetrics
}

// GetBuildMetrics returns current build metrics
func (c *Collector) GetBuildMetrics() BuildMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return *c.buildMetrics
}

// GetNodeMetrics returns current node metrics
func (c *Collector) GetNodeMetrics() NodeMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return *c.nodeMetrics
}

// GetUserMetrics returns current user metrics
func (c *Collector) GetUserMetrics() UserMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Deep copy maps
	metrics := *c.userMetrics
	metrics.APIRequestsByUser = make(map[int]int64)
	metrics.ResourceUsageByUser = make(map[int]*ResourceUsage)

	for k, v := range c.userMetrics.APIRequestsByUser {
		metrics.APIRequestsByUser[k] = v
	}

	for k, v := range c.userMetrics.ResourceUsageByUser {
		usage := *v
		metrics.ResourceUsageByUser[k] = &usage
	}

	return metrics
}

// GetSecurityMetrics returns current security metrics
func (c *Collector) GetSecurityMetrics() SecurityMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return *c.securityMetrics
}

// GetAPIMetrics returns current API metrics
func (c *Collector) GetAPIMetrics() APIMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Deep copy maps
	metrics := *c.apiMetrics
	metrics.RequestsByEndpoint = make(map[string]int64)
	metrics.RequestsByStatus = make(map[int]int64)

	for k, v := range c.apiMetrics.RequestsByEndpoint {
		metrics.RequestsByEndpoint[k] = v
	}

	for k, v := range c.apiMetrics.RequestsByStatus {
		metrics.RequestsByStatus[k] = v
	}

	return metrics
}

// collectSystemMetrics collects system metrics in background
func (c *Collector) collectSystemMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for range ticker.C {
		c.mu.Lock()

		// Collect runtime metrics
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		c.systemMetrics.MemoryUsageBytes = m.Alloc
		c.systemMetrics.MemoryTotalBytes = m.Sys
		c.systemMetrics.GoroutineCount = runtime.NumGoroutine()
		c.systemMetrics.UptimeSeconds = int64(time.Since(startTime).Seconds())
		c.systemMetrics.LastUpdated = time.Now()

		c.mu.Unlock()
	}
}

// Reset resets all metrics (useful for testing)
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buildMetrics = &BuildMetrics{LastUpdated: time.Now()}
	c.nodeMetrics = &NodeMetrics{LastUpdated: time.Now()}
	c.userMetrics = &UserMetrics{
		APIRequestsByUser:   make(map[int]int64),
		ResourceUsageByUser: make(map[int]*ResourceUsage),
		LastUpdated:         time.Now(),
	}
	c.securityMetrics = &SecurityMetrics{LastUpdated: time.Now()}
	c.apiMetrics = &APIMetrics{
		RequestsByEndpoint: make(map[string]int64),
		RequestsByStatus:   make(map[int]int64),
		LastUpdated:        time.Now(),
	}
}

// GetAverageBuildDuration returns average build duration in seconds
func (c *Collector) GetAverageBuildDuration() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.buildMetrics.BuildDurationCount == 0 {
		return 0
	}

	return c.buildMetrics.BuildDurationSum / float64(c.buildMetrics.BuildDurationCount)
}

// GetAverageQueueTime returns average queue time in seconds
func (c *Collector) GetAverageQueueTime() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.buildMetrics.QueueTimeCount == 0 {
		return 0
	}

	return c.buildMetrics.QueueTimeSum / float64(c.buildMetrics.QueueTimeCount)
}

// GetAverageAPILatency returns average API latency in seconds
func (c *Collector) GetAverageAPILatency() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.apiMetrics.RequestDurationCount == 0 {
		return 0
	}

	return c.apiMetrics.RequestDurationSum / float64(c.apiMetrics.RequestDurationCount)
}

// GetSuccessRate returns build success rate as percentage
func (c *Collector) GetSuccessRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.buildMetrics.BuildsSuccess + c.buildMetrics.BuildsFailed
	if total == 0 {
		return 0
	}

	return (float64(c.buildMetrics.BuildsSuccess) / float64(total)) * 100
}
