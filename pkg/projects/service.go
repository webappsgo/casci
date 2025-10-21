package projects

import (
	"context"

	"github.com/casapps/casci/pkg/database"
)

// Service handles project business logic
type Service struct {
	repo *Repository
}

// NewService creates a new project service
func NewService(db *database.Database) *Service {
	return &Service{
		repo: NewRepository(db),
	}
}

// Create creates a new project
func (s *Service) Create(ctx context.Context, userID int, req *CreateProjectRequest) (*Project, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if project name already exists for this user
	exists, err := s.repo.Exists(ctx, userID, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProjectNameExists
	}

	// Create project
	project := &Project{
		UserID:         userID,
		Name:           req.Name,
		RepositoryURL:  req.RepositoryURL,
		Branch:         req.Branch,
		PipelineConfig: req.PipelineConfig,
		AutoDetect:     req.AutoDetect,
	}

	// Set default branch if not provided
	if project.Branch == "" {
		project.Branch = "main"
	}

	// Enable auto-detect by default if no pipeline config provided
	if len(project.PipelineConfig) == 0 {
		project.AutoDetect = true
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

// GetByID retrieves a project by ID
func (s *Service) GetByID(ctx context.Context, id int) (*Project, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByUserAndName retrieves a project by user ID and name
func (s *Service) GetByUserAndName(ctx context.Context, userID int, name string) (*Project, error) {
	return s.repo.GetByUserAndName(ctx, userID, name)
}

// ListByUser retrieves all projects for a user
func (s *Service) ListByUser(ctx context.Context, userID int) ([]*Project, error) {
	return s.repo.ListByUser(ctx, userID)
}

// ListAll retrieves all projects (admin only)
func (s *Service) ListAll(ctx context.Context) ([]*Project, error) {
	return s.repo.ListAll(ctx)
}

// Update updates a project
func (s *Service) Update(ctx context.Context, id int, userID int, req *UpdateProjectRequest) (*Project, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get existing project
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if project.UserID != userID {
		return nil, ErrNotProjectOwner
	}

	// Update fields if provided
	if req.Name != "" {
		// Check if new name already exists for this user
		if req.Name != project.Name {
			exists, err := s.repo.Exists(ctx, userID, req.Name)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, ErrProjectNameExists
			}
		}
		project.Name = req.Name
	}

	if req.RepositoryURL != "" {
		project.RepositoryURL = req.RepositoryURL
	}

	if req.Branch != "" {
		project.Branch = req.Branch
	}

	if len(req.PipelineConfig) > 0 {
		project.PipelineConfig = req.PipelineConfig
	}

	if req.AutoDetect != nil {
		project.AutoDetect = *req.AutoDetect
	}

	if err := s.repo.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

// Delete deletes a project
func (s *Service) Delete(ctx context.Context, id int, userID int) error {
	// Get existing project
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check authorization
	if project.UserID != userID {
		return ErrNotProjectOwner
	}

	return s.repo.Delete(ctx, id)
}

// CheckAccess checks if a user has access to a project
func (s *Service) CheckAccess(ctx context.Context, projectID int, userID int) error {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	if project.UserID != userID {
		return ErrUnauthorizedAccess
	}

	return nil
}

// Count returns the total number of projects for a user
func (s *Service) Count(ctx context.Context, userID int) (int, error) {
	return s.repo.Count(ctx, userID)
}

// IsOwner checks if a user is the owner of a project
func (s *Service) IsOwner(ctx context.Context, projectID int, userID int) (bool, error) {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		if err == ErrProjectNotFound {
			return false, nil
		}
		return false, err
	}

	return project.UserID == userID, nil
}

// GetUserProjects returns all projects for a specific user with pagination
func (s *Service) GetUserProjects(ctx context.Context, userID int, limit, offset int) ([]*Project, error) {
	// For now, just return all projects (pagination can be added later)
	return s.repo.ListByUser(ctx, userID)
}

// Search searches projects by name for a user
func (s *Service) Search(ctx context.Context, userID int, query string) ([]*Project, error) {
	// Get all user projects
	projects, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Simple filtering (can be optimized with database query later)
	var results []*Project
	for _, project := range projects {
		if contains(project.Name, query) || contains(project.RepositoryURL, query) {
			results = append(results, project)
		}
	}

	return results, nil
}

// Helper function for simple string contains
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && (s == substr || len(s) > len(substr) && anyIndexOf(s, substr) >= 0)
}

func anyIndexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
