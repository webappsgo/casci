package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// PostgresDriver implements the Driver interface for PostgreSQL
type PostgresDriver struct{}

// Connect establishes a connection to PostgreSQL
func (d *PostgresDriver) Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	return db, nil
}

// Migrate runs PostgreSQL-specific migrations
func (d *PostgresDriver) Migrate(ctx context.Context, db *sql.DB) error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			api_token VARCHAR(255) UNIQUE,
			is_admin BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Projects table
		`CREATE TABLE IF NOT EXISTS projects (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			repository_url TEXT NOT NULL,
			branch VARCHAR(255) DEFAULT 'main',
			pipeline_config JSONB,
			auto_detect BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, name)
		)`,

		// Builds table
		`CREATE TABLE IF NOT EXISTS builds (
			id SERIAL PRIMARY KEY,
			project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			build_number INTEGER NOT NULL,
			status VARCHAR(50),
			trigger VARCHAR(50),
			commit_sha VARCHAR(40),
			started_at TIMESTAMP,
			finished_at TIMESTAMP,
			log_path TEXT,
			artifacts JSONB,
			UNIQUE(project_id, build_number)
		)`,

		// Nodes table
		`CREATE TABLE IF NOT EXISTS nodes (
			id SERIAL PRIMARY KEY,
			hostname VARCHAR(255) UNIQUE NOT NULL,
			ip_address INET,
			port INTEGER DEFAULT 8080,
			architecture VARCHAR(50),
			os VARCHAR(50),
			role VARCHAR(50),
			capacity JSONB,
			labels JSONB,
			last_heartbeat TIMESTAMP,
			status VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Node tokens table
		`CREATE TABLE IF NOT EXISTS node_tokens (
			id SERIAL PRIMARY KEY,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			used_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Artifacts table
		`CREATE TABLE IF NOT EXISTS artifacts (
			id SERIAL PRIMARY KEY,
			build_id INTEGER NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			path TEXT NOT NULL,
			size BIGINT NOT NULL,
			content_type VARCHAR(255),
			hash VARCHAR(64) NOT NULL,
			storage_type VARCHAR(50) NOT NULL,
			storage_path TEXT NOT NULL,
			compressed BOOLEAN DEFAULT FALSE,
			expires_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_build_id (build_id),
			INDEX idx_hash (hash)
		)`,

		// Server settings table
		`CREATE TABLE IF NOT EXISTS server_settings (
			key VARCHAR(255) PRIMARY KEY,
			value TEXT,
			type VARCHAR(50),
			category VARCHAR(50),
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// User credentials table
		`CREATE TABLE IF NOT EXISTS user_credentials (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			credential_type VARCHAR(50),
			public_key TEXT,
			private_key_encrypted TEXT,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Project credentials table
		`CREATE TABLE IF NOT EXISTS project_credentials (
			id SERIAL PRIMARY KEY,
			project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			credential_type VARCHAR(50),
			name VARCHAR(255),
			value_encrypted TEXT,
			UNIQUE(project_id, credential_type, name)
		)`,

		// Cloud accounts table
		`CREATE TABLE IF NOT EXISTS cloud_accounts (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			provider VARCHAR(50),
			account_name VARCHAR(255),
			credentials_encrypted TEXT,
			monthly_budget DECIMAL(10,2),
			current_spend DECIMAL(10,2),
			UNIQUE(user_id, provider, account_name)
		)`,

		// User nodes table
		`CREATE TABLE IF NOT EXISTS user_nodes (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255),
			connection_type VARCHAR(50),
			connection_details_encrypted JSONB,
			architecture VARCHAR(50),
			exclusive BOOLEAN DEFAULT TRUE,
			UNIQUE(user_id, name)
		)`,

		// Audit log table
		`CREATE TABLE IF NOT EXISTS audit_log (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
			action VARCHAR(255),
			resource_type VARCHAR(50),
			resource_id INTEGER,
			details JSONB,
			ip_address INET,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Build security reports table
		`CREATE TABLE IF NOT EXISTS build_security_reports (
			id SERIAL PRIMARY KEY,
			build_id INTEGER NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
			scan_type VARCHAR(50),
			tool VARCHAR(255),
			critical_count INTEGER DEFAULT 0,
			high_count INTEGER DEFAULT 0,
			medium_count INTEGER DEFAULT 0,
			low_count INTEGER DEFAULT 0,
			info_count INTEGER DEFAULT 0,
			total_count INTEGER DEFAULT 0,
			passed BOOLEAN DEFAULT FALSE,
			details JSONB,
			raw_output TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Notification configs table
		`CREATE TABLE IF NOT EXISTS notification_configs (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			enabled BOOLEAN DEFAULT TRUE,
			config JSONB NOT NULL,
			events JSONB NOT NULL,
			filter TEXT,
			template TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Notification logs table
		`CREATE TABLE IF NOT EXISTS notification_logs (
			id SERIAL PRIMARY KEY,
			config_id INTEGER NOT NULL REFERENCES notification_configs(id) ON DELETE CASCADE,
			build_id INTEGER NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
			event VARCHAR(50) NOT NULL,
			success BOOLEAN DEFAULT FALSE,
			error TEXT,
			sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			duration INTEGER DEFAULT 0
		)`,

		// User credentials table
		`CREATE TABLE IF NOT EXISTS user_credentials (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			credential_type VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			public_key TEXT,
			private_key_encrypted TEXT,
			fingerprint VARCHAR(255),
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP,
			is_default BOOLEAN DEFAULT FALSE
		)`,

		// Project credentials table
		`CREATE TABLE IF NOT EXISTS project_credentials (
			id SERIAL PRIMARY KEY,
			project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			credential_type VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			value_encrypted TEXT NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP,
			last_used_at TIMESTAMP
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
func (d *PostgresDriver) GetType() string {
	return "postgres"
}
