package nodes

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/casapps/casci/src/pkg/database"
)

// Service handles node business logic
type Service struct {
	repo            *Repository
	nodes           map[int]*Node
	mu              sync.RWMutex
	healthCheckStop chan struct{}
}

// NewService creates a new node service
func NewService(db *database.Database) *Service {
	return &Service{
		repo:            NewRepository(db),
		nodes:           make(map[int]*Node),
		healthCheckStop: make(chan struct{}),
	}
}

// Register registers a new node
func (s *Service) Register(ctx context.Context, reg *NodeRegistration) (*Node, error) {
	// Validate token
	token, err := s.repo.GetToken(ctx, reg.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	// Check if token was already used
	if token.UsedAt != nil {
		return nil, fmt.Errorf("token already used")
	}

	// Check if node already exists
	existing, _ := s.repo.GetByHostname(ctx, reg.Hostname)
	if existing != nil {
		return nil, fmt.Errorf("node with hostname %s already exists", reg.Hostname)
	}

	// Create node
	node := &Node{
		Hostname:     reg.Hostname,
		IPAddress:    reg.IPAddress,
		Port:         reg.Port,
		Architecture: reg.Architecture,
		OS:           reg.OS,
		Role:         reg.Role,
		Status:       "online",
		Capacity: map[string]interface{}{
			"cpu_cores":      reg.Capacity.CPUCores,
			"cpu_available":  reg.Capacity.CPUAvailable,
			"memory_total":   reg.Capacity.MemoryTotal,
			"memory_free":    reg.Capacity.MemoryFree,
			"disk_total":     reg.Capacity.DiskTotal,
			"disk_free":      reg.Capacity.DiskFree,
			"max_concurrent": reg.Capacity.MaxConcurrent,
			"current_builds": 0,
		},
		Labels: reg.Labels,
	}

	if err := s.repo.Create(ctx, node); err != nil {
		return nil, err
	}

	// Mark token as used
	if err := s.repo.MarkTokenUsed(ctx, token.ID); err != nil {
		log.Printf("Warning: failed to mark token used: %v", err)
	}

	// Add to cache
	s.mu.Lock()
	s.nodes[node.ID] = node
	s.mu.Unlock()

	log.Printf("Node registered: %s (%s/%s)", node.Hostname, node.OS, node.Architecture)

	return node, nil
}

// GetByID retrieves a node by ID
func (s *Service) GetByID(ctx context.Context, id int) (*Node, error) {
	// Check cache first
	s.mu.RLock()
	if node, ok := s.nodes[id]; ok {
		s.mu.RUnlock()
		return node, nil
	}
	s.mu.RUnlock()

	// Load from database
	node, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update cache
	s.mu.Lock()
	s.nodes[id] = node
	s.mu.Unlock()

	return node, nil
}

// List retrieves all nodes
func (s *Service) List(ctx context.Context) ([]*Node, error) {
	return s.repo.List(ctx)
}

// ListOnline retrieves all online nodes
func (s *Service) ListOnline(ctx context.Context) ([]*Node, error) {
	return s.repo.ListByStatus(ctx, "online")
}

// Update updates a node
func (s *Service) Update(ctx context.Context, node *Node) error {
	if err := s.repo.Update(ctx, node); err != nil {
		return err
	}

	// Update cache
	s.mu.Lock()
	s.nodes[node.ID] = node
	s.mu.Unlock()

	return nil
}

// UpdateHeartbeat updates the last heartbeat for a node
func (s *Service) UpdateHeartbeat(ctx context.Context, id int) error {
	if err := s.repo.UpdateHeartbeat(ctx, id); err != nil {
		return err
	}

	// Update cache
	s.mu.Lock()
	if node, ok := s.nodes[id]; ok {
		node.LastHeartbeat = time.Now()
		node.Status = "online"
	}
	s.mu.Unlock()

	return nil
}

// Drain marks a node as draining (no new builds)
func (s *Service) Drain(ctx context.Context, id int) error {
	if err := s.repo.UpdateStatus(ctx, id, "draining"); err != nil {
		return err
	}

	// Update cache
	s.mu.Lock()
	if node, ok := s.nodes[id]; ok {
		node.Status = "draining"
	}
	s.mu.Unlock()

	return nil
}

// Delete deletes a node
func (s *Service) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Remove from cache
	s.mu.Lock()
	delete(s.nodes, id)
	s.mu.Unlock()

	return nil
}

// GenerateToken generates a new node join token
func (s *Service) GenerateToken(ctx context.Context, expiryMinutes int) (*NodeToken, error) {
	// Generate random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	tokenStr := "ci_" + hex.EncodeToString(bytes)

	token := &NodeToken{
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(time.Duration(expiryMinutes) * time.Minute),
	}

	if err := s.repo.CreateToken(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// ValidateToken validates a node join token
func (s *Service) ValidateToken(ctx context.Context, tokenStr string) error {
	token, err := s.repo.GetToken(ctx, tokenStr)
	if err != nil {
		return fmt.Errorf("invalid token")
	}

	if time.Now().After(token.ExpiresAt) {
		return fmt.Errorf("token expired")
	}

	if token.UsedAt != nil {
		return fmt.Errorf("token already used")
	}

	return nil
}

// SelectNodeForBuild selects the best node for a build based on requirements
func (s *Service) SelectNodeForBuild(ctx context.Context, requirements map[string]interface{}) (*Node, error) {
	nodes, err := s.ListOnline(ctx)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no online nodes available")
	}

	// Filter by architecture if specified
	arch, _ := requirements["architecture"].(string)
	if arch != "" {
		var filtered []*Node
		for _, node := range nodes {
			if node.Architecture == arch {
				filtered = append(filtered, node)
			}
		}
		nodes = filtered
	}

	// Filter by OS if specified
	osReq, _ := requirements["os"].(string)
	if osReq != "" {
		var filtered []*Node
		for _, node := range nodes {
			if node.OS == osReq {
				filtered = append(filtered, node)
			}
		}
		nodes = filtered
	}

	// Filter by labels if specified
	labels, _ := requirements["labels"].(map[string]string)
	if len(labels) > 0 {
		var filtered []*Node
		for _, node := range nodes {
			matches := true
			for key, value := range labels {
				if node.Labels[key] != value {
					matches = false
					break
				}
			}
			if matches {
				filtered = append(filtered, node)
			}
		}
		nodes = filtered
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes match requirements")
	}

	// Exclude draining nodes
	var available []*Node
	for _, node := range nodes {
		if node.Status != "draining" {
			available = append(available, node)
		}
	}

	if len(available) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	// Select node with least load
	var selected *Node
	minLoad := float64(1000000)

	for _, node := range available {
		capacity := node.Capacity
		currentBuilds, _ := capacity["current_builds"].(float64)
		maxConcurrent, _ := capacity["max_concurrent"].(float64)

		if maxConcurrent == 0 {
			maxConcurrent = 5 // Default
		}

		load := currentBuilds / maxConcurrent
		if load < minLoad {
			minLoad = load
			selected = node
		}
	}

	return selected, nil
}

// StartHealthCheck starts the health check goroutine
func (s *Service) StartHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkNodeHealth(ctx)
		case <-s.healthCheckStop:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkNodeHealth checks the health of all nodes
func (s *Service) checkNodeHealth(ctx context.Context) {
	nodes, err := s.List(ctx)
	if err != nil {
		log.Printf("Failed to list nodes for health check: %v", err)
		return
	}

	threshold := 30 * time.Second // Consider offline if no heartbeat for 30s

	for _, node := range nodes {
		if time.Since(node.LastHeartbeat) > threshold && node.Status != "offline" {
			log.Printf("Node %s is offline (no heartbeat for %v)", node.Hostname, time.Since(node.LastHeartbeat))
			if err := s.repo.UpdateStatus(ctx, node.ID, "offline"); err != nil {
				log.Printf("Failed to mark node %s offline: %v", node.Hostname, err)
			}

			// Update cache
			s.mu.Lock()
			if n, ok := s.nodes[node.ID]; ok {
				n.Status = "offline"
			}
			s.mu.Unlock()
		}
	}
}

// StopHealthCheck stops the health check goroutine
func (s *Service) StopHealthCheck() {
	close(s.healthCheckStop)
}
