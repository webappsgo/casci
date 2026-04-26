-- CASCI Database Schema

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    api_token VARCHAR(255) UNIQUE,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    repository_url TEXT NOT NULL,
    branch VARCHAR(255) DEFAULT 'main',
    pipeline_config TEXT,
    auto_detect BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE(user_id, name)
);

-- Builds table
CREATE TABLE IF NOT EXISTS builds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    build_number INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'queued',
    trigger VARCHAR(50) NOT NULL DEFAULT 'manual',
    commit_sha VARCHAR(40),
    commit_message TEXT,
    commit_author VARCHAR(255),
    branch VARCHAR(255),
    repository_url TEXT,
    container_image VARCHAR(255),
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    duration INTEGER DEFAULT 0,
    log_path TEXT,
    artifacts TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(project_id, build_number)
);

-- Nodes table
CREATE TABLE IF NOT EXISTS nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hostname VARCHAR(255) UNIQUE NOT NULL,
    ip_address VARCHAR(45),
    port INTEGER DEFAULT 8080,
    architecture VARCHAR(50),
    os VARCHAR(50),
    role VARCHAR(50),
    capacity TEXT,
    labels TEXT,
    last_heartbeat TIMESTAMP,
    status VARCHAR(50) DEFAULT 'offline'
);

-- Server settings table
CREATE TABLE IF NOT EXISTS server_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT,
    type VARCHAR(50),
    category VARCHAR(50),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_builds_project_id ON builds(project_id);
CREATE INDEX IF NOT EXISTS idx_builds_status ON builds(status);
CREATE INDEX IF NOT EXISTS idx_builds_created_at ON builds(started_at);
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);