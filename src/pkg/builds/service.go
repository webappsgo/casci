package builds

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/casapps/casci/src/pkg/database"
)

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	RecordBuildQueued()
	RecordBuildStarted()
	RecordBuildCompleted(success bool, duration time.Duration, queueTime time.Duration)
	RecordBuildCancelled()
}

// Service handles build business logic
type Service struct {
	repo    *Repository
	metrics MetricsCollector
}

// NewService creates a new build service
func NewService(db *database.Database) *Service {
	return &Service{
		repo:    NewRepository(db),
		metrics: nil, // Will be set via SetMetrics if available
	}
}

// SetMetrics sets the metrics collector for the service
func (s *Service) SetMetrics(metrics MetricsCollector) {
	s.metrics = metrics
}

// Create creates a new build
func (s *Service) Create(ctx context.Context, build *Build) error {
	// Set default status if not set
	if build.Status == "" {
		build.Status = StatusQueued
	}

	// Create the build in the repository
	if err := s.repo.Create(ctx, build); err != nil {
		return err
	}

	// Record metrics if available
	if s.metrics != nil {
		s.metrics.RecordBuildQueued()
	}

	return nil
}

// Trigger triggers a new build for a project
func (s *Service) Trigger(ctx context.Context, projectID int, req *TriggerBuildRequest) (*Build, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get next build number
	buildNumber, err := s.repo.GetNextBuildNumber(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Create build
	build := &Build{
		ProjectID:   projectID,
		BuildNumber: buildNumber,
		Status:      StatusQueued,
		Trigger:     req.Trigger,
		CommitSHA:   req.CommitSHA,
		Branch:      req.Branch,
	}

	// Generate log path
	build.LogPath = fmt.Sprintf("/var/log/casci/builds/project-%d/build-%d.log", projectID, buildNumber)

	if err := s.repo.Create(ctx, build); err != nil {
		return nil, err
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordBuildQueued()
	}

	return build, nil
}

// GetByID retrieves a build by ID
func (s *Service) GetByID(ctx context.Context, id int) (*Build, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByProjectAndNumber retrieves a build by project ID and build number
func (s *Service) GetByProjectAndNumber(ctx context.Context, projectID, buildNumber int) (*Build, error) {
	return s.repo.GetByProjectAndNumber(ctx, projectID, buildNumber)
}

// ListByProject retrieves builds for a project
func (s *Service) ListByProject(ctx context.Context, projectID int, limit int) ([]*Build, error) {
	return s.repo.ListByProject(ctx, projectID, limit)
}

// ListQueued retrieves all queued builds
func (s *Service) ListQueued(ctx context.Context) ([]*Build, error) {
	return s.repo.ListByStatus(ctx, StatusQueued)
}

// ListRunning retrieves all running builds
func (s *Service) ListRunning(ctx context.Context) ([]*Build, error) {
	return s.repo.ListByStatus(ctx, StatusRunning)
}

// Start marks a build as started
func (s *Service) Start(ctx context.Context, id int) error {
	build, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if build.Status != StatusQueued {
		return ErrBuildAlreadyRunning
	}

	now := time.Now()
	build.Status = StatusRunning
	build.StartedAt = &now

	if err := s.repo.Update(ctx, build); err != nil {
		return err
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordBuildStarted()
	}

	return nil
}

// Complete marks a build as completed with the given status
func (s *Service) Complete(ctx context.Context, id int, status BuildStatus) error {
	build, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !build.IsRunning() {
		return fmt.Errorf("build is not running")
	}

	now := time.Now()
	build.Status = status
	build.FinishedAt = &now
	build.CalculateDuration()

	// Calculate queue time (from creation to start)
	// Note: Using 0 for queue time since we don't track creation time separately
	var queueTime time.Duration

	if err := s.repo.Update(ctx, build); err != nil {
		return err
	}

	// Record metrics
	if s.metrics != nil {
		success := (status == StatusSuccess)
		duration := time.Duration(build.Duration) * time.Second
		s.metrics.RecordBuildCompleted(success, duration, queueTime)
	}

	return nil
}

// Cancel cancels a running or queued build
func (s *Service) Cancel(ctx context.Context, id int) error {
	build, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if build.IsCompleted() {
		return ErrCannotCancelBuild
	}

	now := time.Now()
	wasRunning := build.Status == StatusRunning
	build.Status = StatusCanceled

	// If the build was running, set finish time
	if wasRunning {
		build.FinishedAt = &now
		build.CalculateDuration()
	}

	if err := s.repo.Update(ctx, build); err != nil {
		return err
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordBuildCancelled()
	}

	return nil
}

// Delete deletes a build
func (s *Service) Delete(ctx context.Context, id int) error {
	build, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Don't allow deleting running builds
	if build.IsRunning() {
		return fmt.Errorf("cannot delete running build")
	}

	return s.repo.Delete(ctx, id)
}

// UpdateStatus updates the build status
func (s *Service) UpdateStatus(ctx context.Context, id int, status BuildStatus) error {
	build, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	build.Status = status

	// Set timestamps based on status
	now := time.Now()
	switch status {
	case StatusRunning:
		if build.StartedAt == nil {
			build.StartedAt = &now
		}
	case StatusSuccess, StatusFailed, StatusCanceled:
		if build.FinishedAt == nil {
			build.FinishedAt = &now
			build.CalculateDuration()
		}
	}

	return s.repo.Update(ctx, build)
}

// UpdateCommit updates the commit information for a build
func (s *Service) UpdateCommit(ctx context.Context, id int, commitSHA, commitMessage, commitAuthor string) error {
	build, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	build.CommitSHA = commitSHA
	build.CommitMessage = commitMessage
	build.CommitAuthor = commitAuthor

	return s.repo.Update(ctx, build)
}

// SetArtifacts sets the artifacts for a build
func (s *Service) SetArtifacts(ctx context.Context, id int, artifacts interface{}) error {
	build, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Convert artifacts to JSON
	artifactsJSON, err := json.Marshal(artifacts)
	if err != nil {
		return fmt.Errorf("failed to marshal artifacts: %w", err)
	}

	build.Artifacts = artifactsJSON
	return s.repo.Update(ctx, build)
}

// GetLogPath returns the log file path for a build
func (s *Service) GetLogPath(ctx context.Context, id int) (string, error) {
	build, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	if build.LogPath == "" {
		build.LogPath = fmt.Sprintf("/var/log/casci/builds/project-%d/build-%d.log",
			build.ProjectID, build.BuildNumber)
	}

	return build.LogPath, nil
}

// GetStats returns build statistics for a project
func (s *Service) GetStats(ctx context.Context, projectID int) (*BuildStats, error) {
	return s.repo.GetStats(ctx, projectID)
}

// Count returns the total number of builds for a project
func (s *Service) Count(ctx context.Context, projectID int) (int, error) {
	return s.repo.Count(ctx, projectID)
}

// GetLatest returns the latest build for a project
func (s *Service) GetLatest(ctx context.Context, projectID int) (*Build, error) {
	builds, err := s.repo.ListByProject(ctx, projectID, 1)
	if err != nil {
		return nil, err
	}

	if len(builds) == 0 {
		return nil, ErrBuildNotFound
	}

	return builds[0], nil
}

// GetLatestSuccessful returns the latest successful build for a project
func (s *Service) GetLatestSuccessful(ctx context.Context, projectID int) (*Build, error) {
	builds, err := s.repo.ListByProject(ctx, projectID, 100)
	if err != nil {
		return nil, err
	}

	for _, build := range builds {
		if build.Status == StatusSuccess {
			return build, nil
		}
	}

	return nil, ErrBuildNotFound
}

// Restart creates a new build with the same parameters as an existing build
func (s *Service) Restart(ctx context.Context, id int) (*Build, error) {
	// Get the existing build
	existingBuild, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Trigger a new build with the same parameters
	req := &TriggerBuildRequest{
		Branch:    existingBuild.Branch,
		CommitSHA: existingBuild.CommitSHA,
		Trigger:   TriggerManual, // Restart is always manual
	}

	return s.Trigger(ctx, existingBuild.ProjectID, req)
}
