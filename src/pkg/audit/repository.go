package audit

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Repository interface {
	Log(ctx context.Context, event *AuditEvent) error
	Query(ctx context.Context, filter *AuditFilter) ([]AuditEvent, error)
	Count(ctx context.Context, filter *AuditFilter) (int64, error)
	Cleanup(ctx context.Context, olderThan time.Time) (int64, error)
}

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Log(ctx context.Context, event *AuditEvent) error {
	query := `
		INSERT INTO audit_log (
			timestamp, user_id, username, action, resource, resource_id,
			details, ip_address, user_agent, success, error
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	
	result, err := r.db.ExecContext(ctx, query,
		event.Timestamp, event.UserID, event.Username, event.Action,
		event.Resource, event.ResourceID, event.Details, event.IPAddress,
		event.UserAgent, event.Success, event.Error,
	)
	if err != nil {
		return fmt.Errorf("failed to log audit event: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err == nil {
		event.ID = id
	}
	
	return nil
}

func (r *SQLRepository) Query(ctx context.Context, filter *AuditFilter) ([]AuditEvent, error) {
	query := "SELECT id, timestamp, user_id, username, action, resource, resource_id, details, ip_address, user_agent, success, error FROM audit_log"
	var conditions []string
	var args []interface{}
	
	if filter != nil {
		if filter.UserID != nil {
			conditions = append(conditions, "user_id = ?")
			args = append(args, *filter.UserID)
		}
		if filter.Action != "" {
			conditions = append(conditions, "action = ?")
			args = append(args, filter.Action)
		}
		if filter.Resource != "" {
			conditions = append(conditions, "resource = ?")
			args = append(args, filter.Resource)
		}
		if filter.StartTime != nil {
			conditions = append(conditions, "timestamp >= ?")
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			conditions = append(conditions, "timestamp <= ?")
			args = append(args, *filter.EndTime)
		}
		if filter.Success != nil {
			conditions = append(conditions, "success = ?")
			args = append(args, *filter.Success)
		}
	}
	
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	
	query += " ORDER BY timestamp DESC"
	
	if filter != nil {
		if filter.Limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		}
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit log: %w", err)
	}
	defer rows.Close()
	
	var events []AuditEvent
	for rows.Next() {
		var event AuditEvent
		var resourceID sql.NullInt64
		var details, ipAddress, userAgent, errMsg sql.NullString
		
		err := rows.Scan(
			&event.ID, &event.Timestamp, &event.UserID, &event.Username,
			&event.Action, &event.Resource, &resourceID, &details,
			&ipAddress, &userAgent, &event.Success, &errMsg,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}
		
		if resourceID.Valid {
			event.ResourceID = resourceID.Int64
		}
		if details.Valid {
			event.Details = details.String
		}
		if ipAddress.Valid {
			event.IPAddress = ipAddress.String
		}
		if userAgent.Valid {
			event.UserAgent = userAgent.String
		}
		if errMsg.Valid {
			event.Error = errMsg.String
		}
		
		events = append(events, event)
	}
	
	return events, rows.Err()
}

func (r *SQLRepository) Count(ctx context.Context, filter *AuditFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM audit_log"
	var conditions []string
	var args []interface{}
	
	if filter != nil {
		if filter.UserID != nil {
			conditions = append(conditions, "user_id = ?")
			args = append(args, *filter.UserID)
		}
		if filter.Action != "" {
			conditions = append(conditions, "action = ?")
			args = append(args, filter.Action)
		}
		if filter.Resource != "" {
			conditions = append(conditions, "resource = ?")
			args = append(args, filter.Resource)
		}
		if filter.StartTime != nil {
			conditions = append(conditions, "timestamp >= ?")
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			conditions = append(conditions, "timestamp <= ?")
			args = append(args, *filter.EndTime)
		}
		if filter.Success != nil {
			conditions = append(conditions, "success = ?")
			args = append(args, *filter.Success)
		}
	}
	
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	
	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *SQLRepository) Cleanup(ctx context.Context, olderThan time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM audit_log WHERE timestamp < ?", olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup audit log: %w", err)
	}
	
	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	
	return count, nil
}
