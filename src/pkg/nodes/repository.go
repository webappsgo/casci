package nodes

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/casapps/casci/src/pkg/database"
)

// Repository handles node data persistence
type Repository struct {
	db *database.Database
}

// NewRepository creates a new node repository
func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

// Create creates a new node
func (r *Repository) Create(ctx context.Context, node *Node) error {
	capacityJSON, err := json.Marshal(node.Capacity)
	if err != nil {
		return fmt.Errorf("failed to marshal capacity: %w", err)
	}

	labelsJSON, err := json.Marshal(node.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	query := `
		INSERT INTO nodes (hostname, ip_address, port, architecture, os, role,
			status, capacity, labels, last_heartbeat, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query,
		node.Hostname,
		node.IPAddress,
		node.Port,
		node.Architecture,
		node.OS,
		node.Role,
		node.Status,
		string(capacityJSON),
		string(labelsJSON),
		now,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get node id: %w", err)
	}

	node.ID = int(id)
	node.CreatedAt = now
	node.UpdatedAt = now
	node.LastHeartbeat = now

	return nil
}

// GetByID retrieves a node by ID
func (r *Repository) GetByID(ctx context.Context, id int) (*Node, error) {
	query := `
		SELECT id, hostname, ip_address, port, architecture, os, role,
			status, capacity, labels, last_heartbeat, created_at, updated_at
		FROM nodes
		WHERE id = ?
	`

	var node Node
	var capacityJSON, labelsJSON string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&node.ID,
		&node.Hostname,
		&node.IPAddress,
		&node.Port,
		&node.Architecture,
		&node.OS,
		&node.Role,
		&node.Status,
		&capacityJSON,
		&labelsJSON,
		&node.LastHeartbeat,
		&node.CreatedAt,
		&node.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("node not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	if err := json.Unmarshal([]byte(capacityJSON), &node.Capacity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capacity: %w", err)
	}

	if err := json.Unmarshal([]byte(labelsJSON), &node.Labels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
	}

	return &node, nil
}

// GetByHostname retrieves a node by hostname
func (r *Repository) GetByHostname(ctx context.Context, hostname string) (*Node, error) {
	query := `
		SELECT id, hostname, ip_address, port, architecture, os, role,
			status, capacity, labels, last_heartbeat, created_at, updated_at
		FROM nodes
		WHERE hostname = ?
	`

	var node Node
	var capacityJSON, labelsJSON string

	err := r.db.QueryRow(ctx, query, hostname).Scan(
		&node.ID,
		&node.Hostname,
		&node.IPAddress,
		&node.Port,
		&node.Architecture,
		&node.OS,
		&node.Role,
		&node.Status,
		&capacityJSON,
		&labelsJSON,
		&node.LastHeartbeat,
		&node.CreatedAt,
		&node.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("node not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	if err := json.Unmarshal([]byte(capacityJSON), &node.Capacity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capacity: %w", err)
	}

	if err := json.Unmarshal([]byte(labelsJSON), &node.Labels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
	}

	return &node, nil
}

// List retrieves all nodes
func (r *Repository) List(ctx context.Context) ([]*Node, error) {
	query := `
		SELECT id, hostname, ip_address, port, architecture, os, role,
			status, capacity, labels, last_heartbeat, created_at, updated_at
		FROM nodes
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		var capacityJSON, labelsJSON string

		err := rows.Scan(
			&node.ID,
			&node.Hostname,
			&node.IPAddress,
			&node.Port,
			&node.Architecture,
			&node.OS,
			&node.Role,
			&node.Status,
			&capacityJSON,
			&labelsJSON,
			&node.LastHeartbeat,
			&node.CreatedAt,
			&node.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}

		if err := json.Unmarshal([]byte(capacityJSON), &node.Capacity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal capacity: %w", err)
		}

		if err := json.Unmarshal([]byte(labelsJSON), &node.Labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
		}

		nodes = append(nodes, &node)
	}

	return nodes, nil
}

// ListByStatus retrieves nodes by status
func (r *Repository) ListByStatus(ctx context.Context, status string) ([]*Node, error) {
	query := `
		SELECT id, hostname, ip_address, port, architecture, os, role,
			status, capacity, labels, last_heartbeat, created_at, updated_at
		FROM nodes
		WHERE status = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes by status: %w", err)
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		var capacityJSON, labelsJSON string

		err := rows.Scan(
			&node.ID,
			&node.Hostname,
			&node.IPAddress,
			&node.Port,
			&node.Architecture,
			&node.OS,
			&node.Role,
			&node.Status,
			&capacityJSON,
			&labelsJSON,
			&node.LastHeartbeat,
			&node.CreatedAt,
			&node.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}

		if err := json.Unmarshal([]byte(capacityJSON), &node.Capacity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal capacity: %w", err)
		}

		if err := json.Unmarshal([]byte(labelsJSON), &node.Labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
		}

		nodes = append(nodes, &node)
	}

	return nodes, nil
}

// Update updates a node
func (r *Repository) Update(ctx context.Context, node *Node) error {
	capacityJSON, err := json.Marshal(node.Capacity)
	if err != nil {
		return fmt.Errorf("failed to marshal capacity: %w", err)
	}

	labelsJSON, err := json.Marshal(node.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	query := `
		UPDATE nodes
		SET hostname = ?, ip_address = ?, port = ?, architecture = ?, os = ?,
			role = ?, status = ?, capacity = ?, labels = ?, updated_at = ?
		WHERE id = ?
	`

	node.UpdatedAt = time.Now()

	_, err = r.db.Exec(ctx, query,
		node.Hostname,
		node.IPAddress,
		node.Port,
		node.Architecture,
		node.OS,
		node.Role,
		node.Status,
		string(capacityJSON),
		string(labelsJSON),
		node.UpdatedAt,
		node.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	return nil
}

// UpdateHeartbeat updates the last heartbeat time for a node
func (r *Repository) UpdateHeartbeat(ctx context.Context, id int) error {
	query := `
		UPDATE nodes
		SET last_heartbeat = ?, status = 'online', updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of a node
func (r *Repository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE nodes
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// Delete deletes a node
func (r *Repository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM nodes WHERE id = ?`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	return nil
}

// CreateToken creates a new node join token
func (r *Repository) CreateToken(ctx context.Context, token *NodeToken) error {
	query := `
		INSERT INTO node_tokens (token, expires_at, created_at)
		VALUES (?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, token.Token, token.ExpiresAt, now)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get token id: %w", err)
	}

	token.ID = int(id)
	token.CreatedAt = now

	return nil
}

// GetToken retrieves a token by token string
func (r *Repository) GetToken(ctx context.Context, tokenStr string) (*NodeToken, error) {
	query := `
		SELECT id, token, expires_at, used_at, created_at
		FROM node_tokens
		WHERE token = ?
	`

	var token NodeToken
	var usedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, tokenStr).Scan(
		&token.ID,
		&token.Token,
		&token.ExpiresAt,
		&usedAt,
		&token.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	if usedAt.Valid {
		token.UsedAt = &usedAt.Time
	}

	return &token, nil
}

// MarkTokenUsed marks a token as used
func (r *Repository) MarkTokenUsed(ctx context.Context, tokenID int) error {
	query := `
		UPDATE node_tokens
		SET used_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, now, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark token used: %w", err)
	}

	return nil
}
