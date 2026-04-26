package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// SQLRepository implements Repository using SQL database
type SQLRepository struct {
	db *sql.DB
}

// NewSQLRepository creates a new SQL-based notification repository
func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

// CreateConfig creates a new notification configuration
func (r *SQLRepository) CreateConfig(ctx context.Context, config *NotificationConfig) error {
	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	eventsJSON, err := json.Marshal(config.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	query := `
		INSERT INTO notification_configs
		(user_id, project_id, name, type, enabled, config, events, filter, template, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		config.UserID,
		config.ProjectID,
		config.Name,
		config.Type,
		config.Enabled,
		configJSON,
		eventsJSON,
		config.Filter,
		config.Template,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get config ID: %w", err)
	}

	config.ID = int(id)
	config.CreatedAt = now
	config.UpdatedAt = now

	return nil
}

// GetConfig retrieves a notification configuration by ID
func (r *SQLRepository) GetConfig(ctx context.Context, id int) (*NotificationConfig, error) {
	query := `
		SELECT id, user_id, project_id, name, type, enabled, config, events, filter, template, created_at, updated_at
		FROM notification_configs
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanConfig(row)
}

// GetConfigsByUser retrieves all notification configs for a user
func (r *SQLRepository) GetConfigsByUser(ctx context.Context, userID int) ([]*NotificationConfig, error) {
	query := `
		SELECT id, user_id, project_id, name, type, enabled, config, events, filter, template, created_at, updated_at
		FROM notification_configs
		WHERE user_id = ? AND project_id IS NULL AND enabled = 1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query configs: %w", err)
	}
	defer rows.Close()

	var configs []*NotificationConfig
	for rows.Next() {
		config, err := r.scanConfigRow(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return configs, nil
}

// GetConfigsByProject retrieves all notification configs for a project
func (r *SQLRepository) GetConfigsByProject(ctx context.Context, projectID int) ([]*NotificationConfig, error) {
	query := `
		SELECT id, user_id, project_id, name, type, enabled, config, events, filter, template, created_at, updated_at
		FROM notification_configs
		WHERE project_id = ? AND enabled = 1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query configs: %w", err)
	}
	defer rows.Close()

	var configs []*NotificationConfig
	for rows.Next() {
		config, err := r.scanConfigRow(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return configs, nil
}

// UpdateConfig updates a notification configuration
func (r *SQLRepository) UpdateConfig(ctx context.Context, config *NotificationConfig) error {
	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	eventsJSON, err := json.Marshal(config.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	query := `
		UPDATE notification_configs
		SET name = ?, type = ?, enabled = ?, config = ?, events = ?, filter = ?, template = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		config.Name,
		config.Type,
		config.Enabled,
		configJSON,
		eventsJSON,
		config.Filter,
		config.Template,
		now,
		config.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	config.UpdatedAt = now
	return nil
}

// DeleteConfig deletes a notification configuration
func (r *SQLRepository) DeleteConfig(ctx context.Context, id int) error {
	query := `DELETE FROM notification_configs WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	return nil
}

// LogNotification logs a sent notification
func (r *SQLRepository) LogNotification(ctx context.Context, log *NotificationLog) error {
	query := `
		INSERT INTO notification_logs
		(config_id, build_id, event, success, error, sent_at, duration)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		log.ConfigID,
		log.BuildID,
		log.Event,
		log.Success,
		log.Error,
		log.SentAt,
		log.Duration.Milliseconds(),
	)

	if err != nil {
		return fmt.Errorf("failed to log notification: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get log ID: %w", err)
	}

	log.ID = int(id)
	return nil
}

// GetNotificationLogs retrieves notification logs for a build
func (r *SQLRepository) GetNotificationLogs(ctx context.Context, buildID int) ([]*NotificationLog, error) {
	query := `
		SELECT id, config_id, build_id, event, success, error, sent_at, duration
		FROM notification_logs
		WHERE build_id = ?
		ORDER BY sent_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, buildID)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []*NotificationLog
	for rows.Next() {
		log, err := r.scanLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return logs, nil
}

// scanConfig scans a single row into a NotificationConfig
func (r *SQLRepository) scanConfig(row *sql.Row) (*NotificationConfig, error) {
	var config NotificationConfig
	var projectID sql.NullInt64
	var configJSON, eventsJSON []byte
	var filter, template sql.NullString

	err := row.Scan(
		&config.ID,
		&config.UserID,
		&projectID,
		&config.Name,
		&config.Type,
		&config.Enabled,
		&configJSON,
		&eventsJSON,
		&filter,
		&template,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("config not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan config: %w", err)
	}

	if projectID.Valid {
		pid := int(projectID.Int64)
		config.ProjectID = &pid
	}

	if err := json.Unmarshal(configJSON, &config.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := json.Unmarshal(eventsJSON, &config.Events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	if filter.Valid {
		config.Filter = filter.String
	}

	if template.Valid {
		config.Template = template.String
	}

	return &config, nil
}

// scanConfigRow scans a row from query results
func (r *SQLRepository) scanConfigRow(rows *sql.Rows) (*NotificationConfig, error) {
	var config NotificationConfig
	var projectID sql.NullInt64
	var configJSON, eventsJSON []byte
	var filter, template sql.NullString

	err := rows.Scan(
		&config.ID,
		&config.UserID,
		&projectID,
		&config.Name,
		&config.Type,
		&config.Enabled,
		&configJSON,
		&eventsJSON,
		&filter,
		&template,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan config: %w", err)
	}

	if projectID.Valid {
		pid := int(projectID.Int64)
		config.ProjectID = &pid
	}

	if err := json.Unmarshal(configJSON, &config.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := json.Unmarshal(eventsJSON, &config.Events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	if filter.Valid {
		config.Filter = filter.String
	}

	if template.Valid {
		config.Template = template.String
	}

	return &config, nil
}

// scanLog scans a notification log row
func (r *SQLRepository) scanLog(rows *sql.Rows) (*NotificationLog, error) {
	var log NotificationLog
	var errorMsg sql.NullString
	var durationMs int64

	err := rows.Scan(
		&log.ID,
		&log.ConfigID,
		&log.BuildID,
		&log.Event,
		&log.Success,
		&errorMsg,
		&log.SentAt,
		&durationMs,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan log: %w", err)
	}

	if errorMsg.Valid {
		log.Error = errorMsg.String
	}

	log.Duration = time.Duration(durationMs) * time.Millisecond

	return &log, nil
}
