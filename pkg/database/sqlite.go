package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDriver implements the Driver interface for SQLite
type SQLiteDriver struct{}

// Connect establishes a connection to SQLite
func (d *SQLiteDriver) Connect(dsn string) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dsn)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Add pragmas for better performance
	if dsn != ":memory:" {
		dsn = dsn + "?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-64000"
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// SQLite specific settings
	db.SetMaxOpenConns(1) // SQLite doesn't support multiple concurrent writes

	return db, nil
}

// Migrate runs SQLite-specific migrations
func (d *SQLiteDriver) Migrate(ctx context.Context, db *sql.DB) error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			api_token TEXT UNIQUE,
			is_admin INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Projects table
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			repository_url TEXT NOT NULL,
			branch TEXT DEFAULT 'main',
			pipeline_config TEXT,
			auto_detect INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, name)
		)`,

		// Builds table
		`CREATE TABLE IF NOT EXISTS builds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			build_number INTEGER NOT NULL,
			status TEXT,
			trigger TEXT,
			commit_sha TEXT,
			started_at DATETIME,
			finished_at DATETIME,
			log_path TEXT,
			artifacts TEXT,
			FOREIGN KEY (project_id) REFERENCES projects(id),
			UNIQUE(project_id, build_number)
		)`,

		// Nodes table
		`CREATE TABLE IF NOT EXISTS nodes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hostname TEXT UNIQUE NOT NULL,
			ip_address TEXT,
			port INTEGER DEFAULT 8080,
			architecture TEXT,
			os TEXT,
			role TEXT,
			capacity TEXT,
			labels TEXT,
			last_heartbeat DATETIME,
			status TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Node tokens table
		`CREATE TABLE IF NOT EXISTS node_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			used_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Artifacts table
		`CREATE TABLE IF NOT EXISTS artifacts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			build_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER NOT NULL,
			content_type TEXT,
			hash TEXT NOT NULL,
			storage_type TEXT NOT NULL,
			storage_path TEXT NOT NULL,
			compressed INTEGER DEFAULT 0,
			expires_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (build_id) REFERENCES builds(id)
		)`,

		// Server settings table
		`CREATE TABLE IF NOT EXISTS server_settings (
			key TEXT PRIMARY KEY,
			value TEXT,
			type TEXT,
			category TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// User credentials table
		`CREATE TABLE IF NOT EXISTS user_credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			credential_type TEXT,
			public_key TEXT,
			private_key_encrypted TEXT,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		// Project credentials table
		`CREATE TABLE IF NOT EXISTS project_credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			credential_type TEXT,
			name TEXT,
			value_encrypted TEXT,
			FOREIGN KEY (project_id) REFERENCES projects(id),
			UNIQUE(project_id, credential_type, name)
		)`,

		// Cloud accounts table
		`CREATE TABLE IF NOT EXISTS cloud_accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			provider TEXT,
			account_name TEXT,
			credentials_encrypted TEXT,
			monthly_budget REAL,
			current_spend REAL,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, provider, account_name)
		)`,

		// User nodes table
		`CREATE TABLE IF NOT EXISTS user_nodes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT,
			connection_type TEXT,
			connection_details_encrypted TEXT,
			architecture TEXT,
			exclusive INTEGER DEFAULT 1,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, name)
		)`,

		// Audit log table
		`CREATE TABLE IF NOT EXISTS audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			action TEXT,
			resource_type TEXT,
			resource_id INTEGER,
			details TEXT,
			ip_address TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		// Build security reports table
		`CREATE TABLE IF NOT EXISTS build_security_reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			build_id INTEGER NOT NULL,
			scan_type TEXT,
			tool TEXT,
			critical_count INTEGER DEFAULT 0,
			high_count INTEGER DEFAULT 0,
			medium_count INTEGER DEFAULT 0,
			low_count INTEGER DEFAULT 0,
			info_count INTEGER DEFAULT 0,
			total_count INTEGER DEFAULT 0,
			passed BOOLEAN DEFAULT FALSE,
			details TEXT,
			raw_output TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE
		)`,

		// Notification configs table
		`CREATE TABLE IF NOT EXISTS notification_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			project_id INTEGER,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			enabled BOOLEAN DEFAULT TRUE,
			config TEXT NOT NULL,
			events TEXT NOT NULL,
			filter TEXT,
			template TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)`,

		// Notification logs table
		`CREATE TABLE IF NOT EXISTS notification_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			config_id INTEGER NOT NULL,
			build_id INTEGER NOT NULL,
			event TEXT NOT NULL,
			success BOOLEAN DEFAULT FALSE,
			error TEXT,
			sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			duration INTEGER DEFAULT 0,
			FOREIGN KEY (config_id) REFERENCES notification_configs(id) ON DELETE CASCADE,
			FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE
		)`,

		// User credentials table
		`CREATE TABLE IF NOT EXISTS user_credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			credential_type TEXT NOT NULL,
			name TEXT NOT NULL,
			public_key TEXT,
			private_key_encrypted TEXT,
			fingerprint TEXT,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			is_default BOOLEAN DEFAULT FALSE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Project credentials table
		`CREATE TABLE IF NOT EXISTS project_credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			credential_type TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			value_encrypted TEXT NOT NULL,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			last_used_at DATETIME,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Create indices for performance
		`CREATE INDEX IF NOT EXISTS idx_builds_project_id ON builds(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_builds_status ON builds(status)`,
		`CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_notification_configs_user_id ON notification_configs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notification_configs_project_id ON notification_configs(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notification_logs_build_id ON notification_logs(build_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_credentials_user_id ON user_credentials(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_credentials_type ON user_credentials(credential_type)`,
		`CREATE INDEX IF NOT EXISTS idx_project_credentials_project_id ON project_credentials(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_project_credentials_user_id ON project_credentials(user_id)`,
	}

	// Execute migrations
	for _, migration := range migrations {
		if _, err := db.ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, migration)
		}
	}

	return nil
}

// GetType returns the database type
func (d *SQLiteDriver) GetType() string {
	return "sqlite"
}
