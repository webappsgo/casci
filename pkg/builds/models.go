package builds

import (
	"encoding/json"
	"time"
)

// BuildStatus represents the status of a build
type BuildStatus string

const (
	StatusQueued   BuildStatus = "queued"
	StatusRunning  BuildStatus = "running"
	StatusSuccess  BuildStatus = "success"
	StatusFailed   BuildStatus = "failed"
	StatusCanceled BuildStatus = "canceled"
)

// BuildTrigger represents what triggered a build
type BuildTrigger string

const (
	TriggerPush     BuildTrigger = "push"
	TriggerPR       BuildTrigger = "pr"
	TriggerManual   BuildTrigger = "manual"
	TriggerSchedule BuildTrigger = "schedule"
	TriggerAPI      BuildTrigger = "api"
)

// Build represents a build execution
type Build struct {
	ID             int             `json:"id"`
	ProjectID      int             `json:"project_id"`
	BuildNumber    int             `json:"build_number"`
	Status         BuildStatus     `json:"status"`
	Trigger        BuildTrigger    `json:"trigger"`
	CommitSHA      string          `json:"commit_sha,omitempty"`
	CommitMessage  string          `json:"commit_message,omitempty"`
	CommitAuthor   string          `json:"commit_author,omitempty"`
	Branch         string          `json:"branch,omitempty"`
	RepositoryURL  string          `json:"repository_url,omitempty"`
	ContainerImage string          `json:"container_image,omitempty"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	FinishedAt     *time.Time      `json:"finished_at,omitempty"`
	Duration       int64           `json:"duration,omitempty"` // Duration in seconds
	LogPath        string          `json:"log_path,omitempty"`
	Artifacts      json.RawMessage `json:"artifacts,omitempty"`
}

// TriggerBuildRequest represents a request to trigger a build
type TriggerBuildRequest struct {
	Branch    string       `json:"branch,omitempty"`
	CommitSHA string       `json:"commit_sha,omitempty"`
	Trigger   BuildTrigger `json:"trigger,omitempty"`
}

// BuildWithProject includes build with project information
type BuildWithProject struct {
	Build
	ProjectName   string `json:"project_name"`
	RepositoryURL string `json:"repository_url"`
}

// BuildStats represents build statistics
type BuildStats struct {
	TotalBuilds      int     `json:"total_builds"`
	SuccessfulBuilds int     `json:"successful_builds"`
	FailedBuilds     int     `json:"failed_builds"`
	QueuedBuilds     int     `json:"queued_builds"`
	RunningBuilds    int     `json:"running_builds"`
	SuccessRate      float64 `json:"success_rate"`
	AverageDuration  float64 `json:"average_duration"`
}

// Validate validates a trigger build request
func (r *TriggerBuildRequest) Validate() error {
	if r.Trigger == "" {
		r.Trigger = TriggerManual // Default to manual
	}
	return nil
}

// IsRunning checks if a build is currently running
func (b *Build) IsRunning() bool {
	return b.Status == StatusQueued || b.Status == StatusRunning
}

// IsCompleted checks if a build is completed
func (b *Build) IsCompleted() bool {
	return b.Status == StatusSuccess || b.Status == StatusFailed || b.Status == StatusCanceled
}

// CalculateDuration calculates the build duration
func (b *Build) CalculateDuration() {
	if b.StartedAt != nil && b.FinishedAt != nil {
		b.Duration = int64(b.FinishedAt.Sub(*b.StartedAt).Seconds())
	}
}
