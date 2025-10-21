package security

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

// NewSQLRepository creates a new SQL-based security repository
func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

// CreateReport creates a new security report
func (r *SQLRepository) CreateReport(ctx context.Context, report *SecurityReport) error {
	detailsJSON, err := json.Marshal(report.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	query := `
		INSERT INTO build_security_reports
		(build_id, scan_type, tool, critical_count, high_count, medium_count,
		 low_count, info_count, total_count, passed, details, raw_output, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		report.BuildID,
		report.ScanType,
		report.Tool,
		report.CriticalCount,
		report.HighCount,
		report.MediumCount,
		report.LowCount,
		report.InfoCount,
		report.TotalCount,
		report.Passed,
		detailsJSON,
		report.RawOutput,
		report.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get report ID: %w", err)
	}

	report.ID = int(id)
	return nil
}

// GetReportByBuildID retrieves all reports for a specific build
func (r *SQLRepository) GetReportByBuildID(ctx context.Context, buildID int) ([]*SecurityReport, error) {
	query := `
		SELECT id, build_id, scan_type, tool, critical_count, high_count,
		       medium_count, low_count, info_count, total_count, passed,
		       details, raw_output, created_at
		FROM build_security_reports
		WHERE build_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, buildID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reports: %w", err)
	}
	defer rows.Close()

	var reports []*SecurityReport
	for rows.Next() {
		report, err := r.scanReport(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return reports, nil
}

// GetReportByID retrieves a specific report by ID
func (r *SQLRepository) GetReportByID(ctx context.Context, id int) (*SecurityReport, error) {
	query := `
		SELECT id, build_id, scan_type, tool, critical_count, high_count,
		       medium_count, low_count, info_count, total_count, passed,
		       details, raw_output, created_at
		FROM build_security_reports
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanReportRow(row)
}

// ListReports retrieves reports based on filters
func (r *SQLRepository) ListReports(ctx context.Context, filters map[string]interface{}) ([]*SecurityReport, error) {
	query := `
		SELECT id, build_id, scan_type, tool, critical_count, high_count,
		       medium_count, low_count, info_count, total_count, passed,
		       details, raw_output, created_at
		FROM build_security_reports
		WHERE 1=1
	`

	args := []interface{}{}

	if buildID, ok := filters["build_id"]; ok {
		query += " AND build_id = ?"
		args = append(args, buildID)
	}

	if scanType, ok := filters["scan_type"]; ok {
		query += " AND scan_type = ?"
		args = append(args, scanType)
	}

	if tool, ok := filters["tool"]; ok {
		query += " AND tool = ?"
		args = append(args, tool)
	}

	if passed, ok := filters["passed"]; ok {
		query += " AND passed = ?"
		args = append(args, passed)
	}

	if hasCritical, ok := filters["has_critical"]; ok && hasCritical.(bool) {
		query += " AND critical_count > 0"
	}

	if hasHigh, ok := filters["has_high"]; ok && hasHigh.(bool) {
		query += " AND high_count > 0"
	}

	query += " ORDER BY created_at DESC"

	if limit, ok := filters["limit"]; ok {
		query += fmt.Sprintf(" LIMIT %d", limit.(int))
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query reports: %w", err)
	}
	defer rows.Close()

	var reports []*SecurityReport
	for rows.Next() {
		report, err := r.scanReport(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return reports, nil
}

// scanReport scans a row into a SecurityReport
func (r *SQLRepository) scanReport(rows *sql.Rows) (*SecurityReport, error) {
	var report SecurityReport
	var detailsJSON []byte
	var rawOutput sql.NullString
	var createdAt time.Time

	err := rows.Scan(
		&report.ID,
		&report.BuildID,
		&report.ScanType,
		&report.Tool,
		&report.CriticalCount,
		&report.HighCount,
		&report.MediumCount,
		&report.LowCount,
		&report.InfoCount,
		&report.TotalCount,
		&report.Passed,
		&detailsJSON,
		&rawOutput,
		&createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan report: %w", err)
	}

	if err := json.Unmarshal(detailsJSON, &report.Details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal details: %w", err)
	}

	if rawOutput.Valid {
		report.RawOutput = rawOutput.String
	}

	report.CreatedAt = createdAt

	return &report, nil
}

// scanReportRow scans a single row into a SecurityReport
func (r *SQLRepository) scanReportRow(row *sql.Row) (*SecurityReport, error) {
	var report SecurityReport
	var detailsJSON []byte
	var rawOutput sql.NullString
	var createdAt time.Time

	err := row.Scan(
		&report.ID,
		&report.BuildID,
		&report.ScanType,
		&report.Tool,
		&report.CriticalCount,
		&report.HighCount,
		&report.MediumCount,
		&report.LowCount,
		&report.InfoCount,
		&report.TotalCount,
		&report.Passed,
		&detailsJSON,
		&rawOutput,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("report not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan report: %w", err)
	}

	if err := json.Unmarshal(detailsJSON, &report.Details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal details: %w", err)
	}

	if rawOutput.Valid {
		report.RawOutput = rawOutput.String
	}

	report.CreatedAt = createdAt

	return &report, nil
}

// DeleteReportsByBuildID deletes all reports for a build
func (r *SQLRepository) DeleteReportsByBuildID(ctx context.Context, buildID int) error {
	query := `DELETE FROM build_security_reports WHERE build_id = ?`

	_, err := r.db.ExecContext(ctx, query, buildID)
	if err != nil {
		return fmt.Errorf("failed to delete reports: %w", err)
	}

	return nil
}

// GetStatistics retrieves security statistics
func (r *SQLRepository) GetStatistics(ctx context.Context) (*SecurityStatistics, error) {
	stats := &SecurityStatistics{}

	// Total reports
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM build_security_reports").Scan(&stats.TotalReports)
	if err != nil {
		return nil, fmt.Errorf("failed to count reports: %w", err)
	}

	// Total vulnerabilities
	err = r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(total_count), 0) FROM build_security_reports WHERE scan_type = 'vulnerability'").Scan(&stats.TotalVulnerabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to count vulnerabilities: %w", err)
	}

	// Critical vulnerabilities
	err = r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(critical_count), 0) FROM build_security_reports WHERE scan_type = 'vulnerability'").Scan(&stats.CriticalVulnerabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to count critical vulnerabilities: %w", err)
	}

	// High vulnerabilities
	err = r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(high_count), 0) FROM build_security_reports WHERE scan_type = 'vulnerability'").Scan(&stats.HighVulnerabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to count high vulnerabilities: %w", err)
	}

	// Failed scans
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM build_security_reports WHERE passed = 0").Scan(&stats.FailedScans)
	if err != nil {
		return nil, fmt.Errorf("failed to count failed scans: %w", err)
	}

	return stats, nil
}

// SecurityStatistics contains aggregate security statistics
type SecurityStatistics struct {
	TotalReports           int
	TotalVulnerabilities   int
	CriticalVulnerabilities int
	HighVulnerabilities    int
	FailedScans            int
}
