package projects

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/casapps/casci/src/pkg/database"
)

// Repository handles project data operations
type Repository struct {
	db *database.Database
}

// NewRepository creates a new project repository
func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

// Create creates a new project
func (r *Repository) Create(ctx context.Context, project *Project) error {
	query := `
		INSERT INTO projects (user_id, name, repository_url, branch, pipeline_config, auto_detect, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	// Adjust query for PostgreSQL
	if r.db.GetType() == "postgres" {
		query = `
			INSERT INTO projects (user_id, name, repository_url, branch, pipeline_config, auto_detect, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`
	}

	project.CreatedAt = time.Now()

	// Convert pipeline config to string for storage
	var pipelineStr interface{}
	if len(project.PipelineConfig) > 0 {
		pipelineStr = string(project.PipelineConfig)
	}

	if r.db.GetType() == "postgres" {
		err := r.db.QueryRow(ctx, query,
			project.UserID,
			project.Name,
			project.RepositoryURL,
			project.Branch,
			pipelineStr,
			project.AutoDetect,
			project.CreatedAt,
		).Scan(&project.ID)
		return err
	}

	result, err := r.db.Exec(ctx, query,
		project.UserID,
		project.Name,
		project.RepositoryURL,
		project.Branch,
		pipelineStr,
		project.AutoDetect,
		project.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get project ID: %w", err)
	}

	project.ID = int(id)
	return nil
}

// GetByID retrieves a project by ID
func (r *Repository) GetByID(ctx context.Context, id int) (*Project, error) {
	query := `
		SELECT id, user_id, name, repository_url, branch, pipeline_config, auto_detect, created_at
		FROM projects
		WHERE id = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, user_id, name, repository_url, branch, pipeline_config, auto_detect, created_at
			FROM projects
			WHERE id = $1
		`
	}

	project := &Project{}
	var pipelineStr sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.RepositoryURL,
		&project.Branch,
		&pipelineStr,
		&project.AutoDetect,
		&project.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if pipelineStr.Valid {
		project.PipelineConfig = json.RawMessage(pipelineStr.String)
	}

	return project, nil
}

// GetByUserAndName retrieves a project by user ID and name
func (r *Repository) GetByUserAndName(ctx context.Context, userID int, name string) (*Project, error) {
	query := `
		SELECT id, user_id, name, repository_url, branch, pipeline_config, auto_detect, created_at
		FROM projects
		WHERE user_id = ? AND name = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, user_id, name, repository_url, branch, pipeline_config, auto_detect, created_at
			FROM projects
			WHERE user_id = $1 AND name = $2
		`
	}

	project := &Project{}
	var pipelineStr sql.NullString

	err := r.db.QueryRow(ctx, query, userID, name).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.RepositoryURL,
		&project.Branch,
		&pipelineStr,
		&project.AutoDetect,
		&project.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if pipelineStr.Valid {
		project.PipelineConfig = json.RawMessage(pipelineStr.String)
	}

	return project, nil
}

// ListByUser retrieves all projects for a user
func (r *Repository) ListByUser(ctx context.Context, userID int) ([]*Project, error) {
	query := `
		SELECT id, user_id, name, repository_url, branch, pipeline_config, auto_detect, created_at
		FROM projects
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, user_id, name, repository_url, branch, pipeline_config, auto_detect, created_at
			FROM projects
			WHERE user_id = $1
			ORDER BY created_at DESC
		`
	}

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		var pipelineStr sql.NullString

		err := rows.Scan(
			&project.ID,
			&project.UserID,
			&project.Name,
			&project.RepositoryURL,
			&project.Branch,
			&pipelineStr,
			&project.AutoDetect,
			&project.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		if pipelineStr.Valid {
			project.PipelineConfig = json.RawMessage(pipelineStr.String)
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// ListAll retrieves all projects (admin only)
func (r *Repository) ListAll(ctx context.Context) ([]*Project, error) {
	query := `
		SELECT id, user_id, name, repository_url, branch, pipeline_config, auto_detect, created_at
		FROM projects
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		var pipelineStr sql.NullString

		err := rows.Scan(
			&project.ID,
			&project.UserID,
			&project.Name,
			&project.RepositoryURL,
			&project.Branch,
			&pipelineStr,
			&project.AutoDetect,
			&project.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		if pipelineStr.Valid {
			project.PipelineConfig = json.RawMessage(pipelineStr.String)
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// Update updates a project
func (r *Repository) Update(ctx context.Context, project *Project) error {
	query := `
		UPDATE projects
		SET name = ?, repository_url = ?, branch = ?, pipeline_config = ?, auto_detect = ?
		WHERE id = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			UPDATE projects
			SET name = $1, repository_url = $2, branch = $3, pipeline_config = $4, auto_detect = $5
			WHERE id = $6
		`
	}

	var pipelineStr interface{}
	if len(project.PipelineConfig) > 0 {
		pipelineStr = string(project.PipelineConfig)
	}

	_, err := r.db.Exec(ctx, query,
		project.Name,
		project.RepositoryURL,
		project.Branch,
		pipelineStr,
		project.AutoDetect,
		project.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// Delete deletes a project
func (r *Repository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM projects WHERE id = ?`

	if r.db.GetType() == "postgres" {
		query = `DELETE FROM projects WHERE id = $1`
	}

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// Exists checks if a project with the given name exists for a user
func (r *Repository) Exists(ctx context.Context, userID int, name string) (bool, error) {
	query := `SELECT COUNT(*) FROM projects WHERE user_id = ? AND name = ?`

	if r.db.GetType() == "postgres" {
		query = `SELECT COUNT(*) FROM projects WHERE user_id = $1 AND name = $2`
	}

	var count int
	err := r.db.QueryRow(ctx, query, userID, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}

	return count > 0, nil
}

// Count returns the total number of projects for a user
func (r *Repository) Count(ctx context.Context, userID int) (int, error) {
	query := `SELECT COUNT(*) FROM projects WHERE user_id = ?`

	if r.db.GetType() == "postgres" {
		query = `SELECT COUNT(*) FROM projects WHERE user_id = $1`
	}

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count projects: %w", err)
	}

	return count, nil
}
