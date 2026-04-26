package builds

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/casapps/casci/src/pkg/database"
)

// Repository handles build data operations
type Repository struct {
	db *database.Database
}

// NewRepository creates a new build repository
func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

// Create creates a new build
func (r *Repository) Create(ctx context.Context, build *Build) error {
	query := `
		INSERT INTO builds (project_id, build_number, status, trigger, commit_sha, commit_message, commit_author, branch, repository_url, container_image, started_at, finished_at, log_path, artifacts)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Adjust query for PostgreSQL
	if r.db.GetType() == "postgres" {
		query = `
			INSERT INTO builds (project_id, build_number, status, trigger, commit_sha, commit_message, commit_author, branch, repository_url, container_image, started_at, finished_at, log_path, artifacts)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			RETURNING id
		`
	}

	var artifactsStr interface{}
	if len(build.Artifacts) > 0 {
		artifactsStr = string(build.Artifacts)
	}

	if r.db.GetType() == "postgres" {
		err := r.db.QueryRow(ctx, query,
			build.ProjectID,
			build.BuildNumber,
			string(build.Status),
			string(build.Trigger),
			build.CommitSHA,
			build.CommitMessage,
			build.CommitAuthor,
			build.Branch,
			build.RepositoryURL,
			build.ContainerImage,
			build.StartedAt,
			build.FinishedAt,
			build.LogPath,
			artifactsStr,
		).Scan(&build.ID)
		return err
	}

	result, err := r.db.Exec(ctx, query,
		build.ProjectID,
		build.BuildNumber,
		string(build.Status),
		string(build.Trigger),
		build.CommitSHA,
		build.CommitMessage,
		build.CommitAuthor,
		build.Branch,
		build.RepositoryURL,
		build.ContainerImage,
		build.StartedAt,
		build.FinishedAt,
		build.LogPath,
		artifactsStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create build: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get build ID: %w", err)
	}

	build.ID = int(id)
	return nil
}

// GetByID retrieves a build by ID
func (r *Repository) GetByID(ctx context.Context, id int) (*Build, error) {
	query := `
		SELECT id, project_id, build_number, status, trigger, commit_sha, commit_message, commit_author, branch, repository_url, container_image, started_at, finished_at, log_path, artifacts
		FROM builds
		WHERE id = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, project_id, build_number, status, trigger, commit_sha, commit_message, commit_author, branch, repository_url, container_image, started_at, finished_at, log_path, artifacts
			FROM builds
			WHERE id = $1
		`
	}

	build := &Build{}
	var artifactsStr sql.NullString
	var commitMessage, commitAuthor, branch, repoURL, containerImage sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&build.ID,
		&build.ProjectID,
		&build.BuildNumber,
		&build.Status,
		&build.Trigger,
		&build.CommitSHA,
		&commitMessage,
		&commitAuthor,
		&branch,
		&repoURL,
		&containerImage,
		&build.StartedAt,
		&build.FinishedAt,
		&build.LogPath,
		&artifactsStr,
	)

	if err == sql.ErrNoRows {
		return nil, ErrBuildNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get build: %w", err)
	}

	if commitMessage.Valid {
		build.CommitMessage = commitMessage.String
	}
	if commitAuthor.Valid {
		build.CommitAuthor = commitAuthor.String
	}
	if branch.Valid {
		build.Branch = branch.String
	}
	if repoURL.Valid {
		build.RepositoryURL = repoURL.String
	}
	if containerImage.Valid {
		build.ContainerImage = containerImage.String
	}
	if artifactsStr.Valid {
		build.Artifacts = json.RawMessage(artifactsStr.String)
	}

	build.CalculateDuration()
	return build, nil
}

// GetByProjectAndNumber retrieves a build by project ID and build number
func (r *Repository) GetByProjectAndNumber(ctx context.Context, projectID, buildNumber int) (*Build, error) {
	query := `
		SELECT id, project_id, build_number, status, trigger, commit_sha, commit_message, commit_author, branch, repository_url, container_image, started_at, finished_at, log_path, artifacts
		FROM builds
		WHERE project_id = ? AND build_number = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, project_id, build_number, status, trigger, commit_sha, commit_message, commit_author, branch, repository_url, container_image, started_at, finished_at, log_path, artifacts
			FROM builds
			WHERE project_id = $1 AND build_number = $2
		`
	}

	build := &Build{}
	var artifactsStr sql.NullString
	var commitMessage, commitAuthor, branch, repoURL, containerImage sql.NullString

	err := r.db.QueryRow(ctx, query, projectID, buildNumber).Scan(
		&build.ID,
		&build.ProjectID,
		&build.BuildNumber,
		&build.Status,
		&build.Trigger,
		&build.CommitSHA,
		&commitMessage,
		&commitAuthor,
		&branch,
		&repoURL,
		&containerImage,
		&build.StartedAt,
		&build.FinishedAt,
		&build.LogPath,
		&artifactsStr,
	)

	if err == sql.ErrNoRows {
		return nil, ErrBuildNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get build: %w", err)
	}

	if commitMessage.Valid {
		build.CommitMessage = commitMessage.String
	}
	if commitAuthor.Valid {
		build.CommitAuthor = commitAuthor.String
	}
	if branch.Valid {
		build.Branch = branch.String
	}
	if repoURL.Valid {
		build.RepositoryURL = repoURL.String
	}
	if containerImage.Valid {
		build.ContainerImage = containerImage.String
	}
	if artifactsStr.Valid {
		build.Artifacts = json.RawMessage(artifactsStr.String)
	}

	build.CalculateDuration()
	return build, nil
}

// ListByProject retrieves all builds for a project
func (r *Repository) ListByProject(ctx context.Context, projectID int, limit int) ([]*Build, error) {
	query := `
		SELECT id, project_id, build_number, status, trigger, commit_sha, commit_message, commit_author, branch, repository_url, container_image, started_at, finished_at, log_path, artifacts
		FROM builds
		WHERE project_id = ?
		ORDER BY build_number DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, project_id, build_number, status, trigger, commit_sha, commit_message, commit_author, branch, repository_url, container_image, started_at, finished_at, log_path, artifacts
			FROM builds
			WHERE project_id = $1
			ORDER BY build_number DESC
		`
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
		}
	}

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list builds: %w", err)
	}
	defer rows.Close()

	var builds []*Build
	for rows.Next() {
		build := &Build{}
		var artifactsStr sql.NullString
		var commitMessage, commitAuthor, branch, repoURL, containerImage sql.NullString

		err := rows.Scan(
			&build.ID,
			&build.ProjectID,
			&build.BuildNumber,
			&build.Status,
			&build.Trigger,
			&build.CommitSHA,
			&commitMessage,
			&commitAuthor,
			&branch,
			&repoURL,
			&containerImage,
			&build.StartedAt,
			&build.FinishedAt,
			&build.LogPath,
			&artifactsStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan build: %w", err)
		}

		if commitMessage.Valid {
			build.CommitMessage = commitMessage.String
		}
		if commitAuthor.Valid {
			build.CommitAuthor = commitAuthor.String
		}
		if branch.Valid {
			build.Branch = branch.String
		}
		if repoURL.Valid {
			build.RepositoryURL = repoURL.String
		}
		if containerImage.Valid {
			build.ContainerImage = containerImage.String
		}
		if artifactsStr.Valid {
			build.Artifacts = json.RawMessage(artifactsStr.String)
		}

		build.CalculateDuration()
		builds = append(builds, build)
	}

	return builds, nil
}

// ListByStatus retrieves builds by status
func (r *Repository) ListByStatus(ctx context.Context, status BuildStatus) ([]*Build, error) {
	query := `
		SELECT id, project_id, build_number, status, trigger, commit_sha, commit_message, commit_author, branch, repository_url, container_image, started_at, finished_at, log_path, artifacts
		FROM builds
		WHERE status = ?
		ORDER BY id ASC
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, project_id, build_number, status, trigger, commit_sha, commit_message, commit_author, branch, repository_url, container_image, started_at, finished_at, log_path, artifacts
			FROM builds
			WHERE status = $1
			ORDER BY id ASC
		`
	}

	rows, err := r.db.Query(ctx, query, string(status))
	if err != nil {
		return nil, fmt.Errorf("failed to list builds by status: %w", err)
	}
	defer rows.Close()

	var builds []*Build
	for rows.Next() {
		build := &Build{}
		var artifactsStr sql.NullString
		var commitMessage, commitAuthor, branch, repoURL, containerImage sql.NullString

		err := rows.Scan(
			&build.ID,
			&build.ProjectID,
			&build.BuildNumber,
			&build.Status,
			&build.Trigger,
			&build.CommitSHA,
			&commitMessage,
			&commitAuthor,
			&branch,
			&repoURL,
			&containerImage,
			&build.StartedAt,
			&build.FinishedAt,
			&build.LogPath,
			&artifactsStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan build: %w", err)
		}

		if commitMessage.Valid {
			build.CommitMessage = commitMessage.String
		}
		if commitAuthor.Valid {
			build.CommitAuthor = commitAuthor.String
		}
		if branch.Valid {
			build.Branch = branch.String
		}
		if repoURL.Valid {
			build.RepositoryURL = repoURL.String
		}
		if containerImage.Valid {
			build.ContainerImage = containerImage.String
		}
		if artifactsStr.Valid {
			build.Artifacts = json.RawMessage(artifactsStr.String)
		}

		build.CalculateDuration()
		builds = append(builds, build)
	}

	return builds, nil
}

// Update updates a build
func (r *Repository) Update(ctx context.Context, build *Build) error {
	query := `
		UPDATE builds
		SET status = ?, commit_sha = ?, commit_message = ?, commit_author = ?, branch = ?, repository_url = ?, container_image = ?, started_at = ?, finished_at = ?, log_path = ?, artifacts = ?
		WHERE id = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			UPDATE builds
			SET status = $1, commit_sha = $2, commit_message = $3, commit_author = $4, branch = $5, repository_url = $6, container_image = $7, started_at = $8, finished_at = $9, log_path = $10, artifacts = $11
			WHERE id = $12
		`
	}

	var artifactsStr interface{}
	if len(build.Artifacts) > 0 {
		artifactsStr = string(build.Artifacts)
	}

	_, err := r.db.Exec(ctx, query,
		string(build.Status),
		build.CommitSHA,
		build.CommitMessage,
		build.CommitAuthor,
		build.Branch,
		build.RepositoryURL,
		build.ContainerImage,
		build.StartedAt,
		build.FinishedAt,
		build.LogPath,
		artifactsStr,
		build.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update build: %w", err)
	}

	return nil
}

// Delete deletes a build
func (r *Repository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM builds WHERE id = ?`

	if r.db.GetType() == "postgres" {
		query = `DELETE FROM builds WHERE id = $1`
	}

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete build: %w", err)
	}

	return nil
}

// GetNextBuildNumber returns the next build number for a project
func (r *Repository) GetNextBuildNumber(ctx context.Context, projectID int) (int, error) {
	query := `SELECT COALESCE(MAX(build_number), 0) + 1 FROM builds WHERE project_id = ?`

	if r.db.GetType() == "postgres" {
		query = `SELECT COALESCE(MAX(build_number), 0) + 1 FROM builds WHERE project_id = $1`
	}

	var nextNumber int
	err := r.db.QueryRow(ctx, query, projectID).Scan(&nextNumber)
	if err != nil {
		return 0, fmt.Errorf("failed to get next build number: %w", err)
	}

	return nextNumber, nil
}

// Count returns the total number of builds for a project
func (r *Repository) Count(ctx context.Context, projectID int) (int, error) {
	query := `SELECT COUNT(*) FROM builds WHERE project_id = ?`

	if r.db.GetType() == "postgres" {
		query = `SELECT COUNT(*) FROM builds WHERE project_id = $1`
	}

	var count int
	err := r.db.QueryRow(ctx, query, projectID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count builds: %w", err)
	}

	return count, nil
}

// GetStats returns build statistics for a project
func (r *Repository) GetStats(ctx context.Context, projectID int) (*BuildStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN status = 'queued' THEN 1 ELSE 0 END) as queued,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running
		FROM builds
		WHERE project_id = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT
				COUNT(*) as total,
				SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success,
				SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
				SUM(CASE WHEN status = 'queued' THEN 1 ELSE 0 END) as queued,
				SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running
			FROM builds
			WHERE project_id = $1
		`
	}

	stats := &BuildStats{}
	err := r.db.QueryRow(ctx, query, projectID).Scan(
		&stats.TotalBuilds,
		&stats.SuccessfulBuilds,
		&stats.FailedBuilds,
		&stats.QueuedBuilds,
		&stats.RunningBuilds,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get build stats: %w", err)
	}

	// Calculate success rate
	if stats.TotalBuilds > 0 {
		stats.SuccessRate = float64(stats.SuccessfulBuilds) / float64(stats.TotalBuilds) * 100
	}

	return stats, nil
}
