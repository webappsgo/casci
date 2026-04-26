package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLDriver implements the Driver interface for MySQL/MariaDB
type MySQLDriver struct{}

// Connect establishes a connection to MySQL
func (d *MySQLDriver) Connect(dsn string) (*sql.DB, error) {
	// Ensure UTF8MB4 is used
	if dsn != "" && dsn[len(dsn)-1] != '?' && dsn[len(dsn)-1] != '&' {
		if dsn[len(dsn)-1] != '/' {
			dsn += "?"
		}
		dsn += "charset=utf8mb4&parseTime=true&loc=Local"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL database: %w", err)
	}

	return db, nil
}

// Migrate runs MySQL-specific migrations
func (d *MySQLDriver) Migrate(ctx context.Context, db *sql.DB) error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			api_token VARCHAR(255) UNIQUE,
			is_admin BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Projects table
		`CREATE TABLE IF NOT EXISTS projects (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			repository_url TEXT NOT NULL,
			branch VARCHAR(255) DEFAULT 'main',
			pipeline_config JSON,
			auto_detect BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY unique_user_project (user_id, name),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Builds table
		"CREATE TABLE IF NOT EXISTS builds (" +
			"id INT AUTO_INCREMENT PRIMARY KEY," +
			"project_id INT NOT NULL," +
			"build_number INT NOT NULL," +
			"status VARCHAR(50)," +
			"`trigger` VARCHAR(50)," +
			"commit_sha VARCHAR(40)," +
			"commit_message TEXT," +
			"commit_author VARCHAR(255)," +
			"branch VARCHAR(255)," +
			"repository_url TEXT," +
			"container_image VARCHAR(255)," +
			"started_at TIMESTAMP NULL," +
			"finished_at TIMESTAMP NULL," +
			"duration INT DEFAULT 0," +
			"log_path TEXT," +
			"artifacts JSON," +
			"UNIQUE KEY unique_project_build (project_id, build_number)," +
			"FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE," +
			"INDEX idx_project_id (project_id)," +
			"INDEX idx_status (status)" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",

		// Nodes table
		`CREATE TABLE IF NOT EXISTS nodes (
			id INT AUTO_INCREMENT PRIMARY KEY,
			hostname VARCHAR(255) UNIQUE NOT NULL,
			ip_address VARCHAR(45),
			port INT DEFAULT 8080,
			architecture VARCHAR(50),
			os VARCHAR(50),
			role VARCHAR(50),
			capacity JSON,
			labels JSON,
			last_heartbeat TIMESTAMP NULL,
			status VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_hostname (hostname)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Node tokens table
		`CREATE TABLE IF NOT EXISTS node_tokens (
			id INT AUTO_INCREMENT PRIMARY KEY,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			used_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_token (token)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Artifacts table
		`CREATE TABLE IF NOT EXISTS artifacts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			build_id INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			path TEXT NOT NULL,
			size BIGINT NOT NULL,
			content_type VARCHAR(255),
			hash VARCHAR(64) NOT NULL,
			storage_type VARCHAR(50) NOT NULL,
			storage_path TEXT NOT NULL,
			compressed BOOLEAN DEFAULT FALSE,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE,
			INDEX idx_build_id (build_id),
			INDEX idx_hash (hash)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Server settings table
		"CREATE TABLE IF NOT EXISTS server_settings (" +
			"`key` VARCHAR(255) PRIMARY KEY," +
			"value TEXT," +
			"type VARCHAR(50)," +
			"category VARCHAR(50)," +
			"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",

		// User credentials table
		`CREATE TABLE IF NOT EXISTS user_credentials (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			credential_type VARCHAR(50),
			public_key TEXT,
			private_key_encrypted TEXT,
			metadata JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Project credentials table
		`CREATE TABLE IF NOT EXISTS project_credentials (
			id INT AUTO_INCREMENT PRIMARY KEY,
			project_id INT NOT NULL,
			credential_type VARCHAR(50),
			name VARCHAR(255),
			value_encrypted TEXT,
			UNIQUE KEY unique_project_cred (project_id, credential_type, name),
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
			INDEX idx_project_id (project_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Cloud accounts table
		`CREATE TABLE IF NOT EXISTS cloud_accounts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			provider VARCHAR(50),
			account_name VARCHAR(255),
			credentials_encrypted TEXT,
			monthly_budget DECIMAL(10,2),
			current_spend DECIMAL(10,2),
			UNIQUE KEY unique_cloud_account (user_id, provider, account_name),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// User nodes table
		`CREATE TABLE IF NOT EXISTS user_nodes (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			name VARCHAR(255),
			connection_type VARCHAR(50),
			connection_details_encrypted JSON,
			architecture VARCHAR(50),
			exclusive BOOLEAN DEFAULT TRUE,
			UNIQUE KEY unique_user_node (user_id, name),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Audit log table
		`CREATE TABLE IF NOT EXISTS audit_log (
			id INT AUTO_INCREMENT PRIMARY KEY,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			user_id INT NOT NULL,
			username VARCHAR(255) NOT NULL,
			action VARCHAR(255) NOT NULL,
			resource VARCHAR(50) NOT NULL,
			resource_id INT,
			details TEXT,
			ip_address VARCHAR(45),
			user_agent TEXT,
			success BOOLEAN DEFAULT TRUE,
			error TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_id (user_id),
			INDEX idx_timestamp (timestamp),
			INDEX idx_action (action),
			INDEX idx_resource (resource)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Build security reports table
		`CREATE TABLE IF NOT EXISTS build_security_reports (
			id INT AUTO_INCREMENT PRIMARY KEY,
			build_id INT NOT NULL,
			scan_type VARCHAR(50),
			tool VARCHAR(255),
			critical_count INT DEFAULT 0,
			high_count INT DEFAULT 0,
			medium_count INT DEFAULT 0,
			low_count INT DEFAULT 0,
			info_count INT DEFAULT 0,
			total_count INT DEFAULT 0,
			passed BOOLEAN DEFAULT FALSE,
			details JSON,
			raw_output LONGTEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE,
			INDEX idx_build_id (build_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Notification configs table
		`CREATE TABLE IF NOT EXISTS notification_configs (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			project_id INT,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			enabled BOOLEAN DEFAULT TRUE,
			config JSON NOT NULL,
			events JSON NOT NULL,
			filter TEXT,
			template TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
			INDEX idx_user_id (user_id),
			INDEX idx_project_id (project_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Notification logs table
		`CREATE TABLE IF NOT EXISTS notification_logs (
			id INT AUTO_INCREMENT PRIMARY KEY,
			config_id INT NOT NULL,
			build_id INT NOT NULL,
			event VARCHAR(50) NOT NULL,
			success BOOLEAN DEFAULT FALSE,
			error TEXT,
			sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			duration INT DEFAULT 0,
			FOREIGN KEY (config_id) REFERENCES notification_configs(id) ON DELETE CASCADE,
			FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE,
			INDEX idx_build_id (build_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// User credentials table
		"CREATE TABLE IF NOT EXISTS user_credentials (" +
			"id INT AUTO_INCREMENT PRIMARY KEY," +
			"user_id INT NOT NULL," +
			"credential_type VARCHAR(50) NOT NULL," +
			"name VARCHAR(255) NOT NULL," +
			"public_key TEXT," +
			"private_key_encrypted TEXT," +
			"fingerprint VARCHAR(255)," +
			"metadata JSON," +
			"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP," +
			"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP," +
			"expires_at TIMESTAMP NULL DEFAULT NULL," +
			"is_default BOOLEAN DEFAULT FALSE," +
			"FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE," +
			"INDEX idx_user_id (user_id)," +
			"INDEX idx_credential_type (credential_type)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",

		// Project credentials table
		"CREATE TABLE IF NOT EXISTS project_credentials (" +
			"id INT AUTO_INCREMENT PRIMARY KEY," +
			"project_id INT NOT NULL," +
			"user_id INT NOT NULL," +
			"credential_type VARCHAR(50) NOT NULL," +
			"name VARCHAR(255) NOT NULL," +
			"description TEXT," +
			"value_encrypted TEXT NOT NULL," +
			"metadata JSON," +
			"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP," +
			"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP," +
			"expires_at TIMESTAMP NULL DEFAULT NULL," +
			"last_used_at TIMESTAMP NULL DEFAULT NULL," +
			"FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE," +
			"FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE," +
			"INDEX idx_project_id (project_id)," +
			"INDEX idx_user_id (user_id)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",
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
func (d *MySQLDriver) GetType() string {
	return "mysql"
}
