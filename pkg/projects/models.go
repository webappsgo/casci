package projects

import (
	"encoding/json"
	"time"
)

// Project represents a CI/CD project
type Project struct {
	ID             int             `json:"id"`
	UserID         int             `json:"user_id"`
	Name           string          `json:"name"`
	RepositoryURL  string          `json:"repository_url"`
	Branch         string          `json:"branch"`
	PipelineConfig json.RawMessage `json:"pipeline_config,omitempty"`
	AutoDetect     bool            `json:"auto_detect"`
	CreatedAt      time.Time       `json:"created_at"`
}

// CreateProjectRequest represents a project creation request
type CreateProjectRequest struct {
	Name           string          `json:"name"`
	RepositoryURL  string          `json:"repository_url"`
	Branch         string          `json:"branch"`
	PipelineConfig json.RawMessage `json:"pipeline_config,omitempty"`
	AutoDetect     bool            `json:"auto_detect"`
}

// UpdateProjectRequest represents a project update request
type UpdateProjectRequest struct {
	Name           string          `json:"name,omitempty"`
	RepositoryURL  string          `json:"repository_url,omitempty"`
	Branch         string          `json:"branch,omitempty"`
	PipelineConfig json.RawMessage `json:"pipeline_config,omitempty"`
	AutoDetect     *bool           `json:"auto_detect,omitempty"`
}

// ProjectWithStats includes project with build statistics
type ProjectWithStats struct {
	Project
	TotalBuilds      int        `json:"total_builds"`
	SuccessfulBuilds int        `json:"successful_builds"`
	FailedBuilds     int        `json:"failed_builds"`
	LastBuildTime    *time.Time `json:"last_build_time,omitempty"`
	LastBuildStatus  string     `json:"last_build_status,omitempty"`
}

// Validate validates a create project request
func (r *CreateProjectRequest) Validate() error {
	if r.Name == "" {
		return ErrInvalidProjectName
	}
	if len(r.Name) < 3 {
		return ErrProjectNameTooShort
	}
	if r.RepositoryURL == "" {
		return ErrInvalidRepositoryURL
	}
	if r.Branch == "" {
		r.Branch = "main" // Default branch
	}
	return nil
}

// Validate validates an update project request
func (r *UpdateProjectRequest) Validate() error {
	if r.Name != "" && len(r.Name) < 3 {
		return ErrProjectNameTooShort
	}
	return nil
}
