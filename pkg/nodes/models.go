package nodes

import (
	"time"
)

// Node represents a build node in the cluster
type Node struct {
	ID            int                    `json:"id"`
	Hostname      string                 `json:"hostname"`
	IPAddress     string                 `json:"ip_address"`
	Port          int                    `json:"port"`
	Architecture  string                 `json:"architecture"` // amd64, arm64
	OS            string                 `json:"os"`           // linux, darwin, windows
	Role          string                 `json:"role"`         // orchestrator, builder, hybrid
	Status        string                 `json:"status"`       // online, offline, draining
	Capacity      map[string]interface{} `json:"capacity"`     // CPU, RAM, disk
	Labels        map[string]string      `json:"labels"`
	LastHeartbeat time.Time              `json:"last_heartbeat"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// NodeCapacity represents the capacity of a node
type NodeCapacity struct {
	CPUCores      int   `json:"cpu_cores"`
	CPUAvailable  int   `json:"cpu_available"`
	MemoryTotal   int64 `json:"memory_total"`    // bytes
	MemoryFree    int64 `json:"memory_free"`     // bytes
	DiskTotal     int64 `json:"disk_total"`      // bytes
	DiskFree      int64 `json:"disk_free"`       // bytes
	MaxConcurrent int   `json:"max_concurrent"`  // max concurrent builds
	CurrentBuilds int   `json:"current_builds"`  // current running builds
}

// NodeRegistration represents a node registration request
type NodeRegistration struct {
	Hostname     string            `json:"hostname"`
	IPAddress    string            `json:"ip_address"`
	Port         int               `json:"port"`
	Architecture string            `json:"architecture"`
	OS           string            `json:"os"`
	Role         string            `json:"role"`
	Token        string            `json:"token"`
	Capacity     NodeCapacity      `json:"capacity"`
	Labels       map[string]string `json:"labels"`
}

// NodeHealth represents the health status of a node
type NodeHealth struct {
	NodeID        int       `json:"node_id"`
	Status        string    `json:"status"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsage   float64   `json:"memory_usage"`
	DiskUsage     float64   `json:"disk_usage"`
	CurrentBuilds int       `json:"current_builds"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// NodeToken represents a node join token
type NodeToken struct {
	ID        int       `json:"id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
