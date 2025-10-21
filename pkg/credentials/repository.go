package credentials

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Repository handles credential data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new credential repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// User Credentials

// CreateUserCredential creates a new user credential
func (r *Repository) CreateUserCredential(ctx context.Context, cred *UserCredential) error {
	metadataJSON, err := json.Marshal(cred.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO user_credentials (
			user_id, credential_type, name, public_key, private_key_encrypted,
			fingerprint, metadata, expires_at, is_default, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		cred.UserID,
		cred.Type,
		cred.Name,
		cred.PublicKey,
		cred.PrivateKeyEncrypted,
		cred.Fingerprint,
		string(metadataJSON),
		cred.ExpiresAt,
		cred.IsDefault,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create user credential: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted ID: %w", err)
	}

	cred.ID = int(id)
	cred.CreatedAt = time.Now()
	cred.UpdatedAt = time.Now()

	return nil
}

// GetUserCredential retrieves a user credential by ID
func (r *Repository) GetUserCredential(ctx context.Context, id int) (*UserCredential, error) {
	query := `
		SELECT id, user_id, credential_type, name, public_key, private_key_encrypted,
			fingerprint, metadata, created_at, updated_at, expires_at, is_default
		FROM user_credentials
		WHERE id = ?
	`

	var cred UserCredential
	var metadataJSON string
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cred.ID,
		&cred.UserID,
		&cred.Type,
		&cred.Name,
		&cred.PublicKey,
		&cred.PrivateKeyEncrypted,
		&cred.Fingerprint,
		&metadataJSON,
		&cred.CreatedAt,
		&cred.UpdatedAt,
		&expiresAt,
		&cred.IsDefault,
	)
	if err == sql.ErrNoRows {
		return nil, ErrCredentialNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user credential: %w", err)
	}

	if expiresAt.Valid {
		cred.ExpiresAt = &expiresAt.Time
	}

	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &cred.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &cred, nil
}

// ListUserCredentials lists all credentials for a user
func (r *Repository) ListUserCredentials(ctx context.Context, userID int) ([]*UserCredential, error) {
	query := `
		SELECT id, user_id, credential_type, name, public_key, private_key_encrypted,
			fingerprint, metadata, created_at, updated_at, expires_at, is_default
		FROM user_credentials
		WHERE user_id = ?
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user credentials: %w", err)
	}
	defer rows.Close()

	var credentials []*UserCredential
	for rows.Next() {
		var cred UserCredential
		var metadataJSON string
		var expiresAt sql.NullTime

		err := rows.Scan(
			&cred.ID,
			&cred.UserID,
			&cred.Type,
			&cred.Name,
			&cred.PublicKey,
			&cred.PrivateKeyEncrypted,
			&cred.Fingerprint,
			&metadataJSON,
			&cred.CreatedAt,
			&cred.UpdatedAt,
			&expiresAt,
			&cred.IsDefault,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user credential: %w", err)
		}

		if expiresAt.Valid {
			cred.ExpiresAt = &expiresAt.Time
		}

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &cred.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		credentials = append(credentials, &cred)
	}

	return credentials, nil
}

// UpdateUserCredential updates a user credential
func (r *Repository) UpdateUserCredential(ctx context.Context, cred *UserCredential) error {
	metadataJSON, err := json.Marshal(cred.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE user_credentials
		SET name = ?, public_key = ?, private_key_encrypted = ?,
			fingerprint = ?, metadata = ?, expires_at = ?, is_default = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContext(ctx, query,
		cred.Name,
		cred.PublicKey,
		cred.PrivateKeyEncrypted,
		cred.Fingerprint,
		string(metadataJSON),
		cred.ExpiresAt,
		cred.IsDefault,
		time.Now(),
		cred.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user credential: %w", err)
	}

	cred.UpdatedAt = time.Now()
	return nil
}

// DeleteUserCredential deletes a user credential
func (r *Repository) DeleteUserCredential(ctx context.Context, id int) error {
	query := `DELETE FROM user_credentials WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user credential: %w", err)
	}
	return nil
}

// GetDefaultUserCredential gets the default credential of a specific type for a user
func (r *Repository) GetDefaultUserCredential(ctx context.Context, userID int, credType CredentialType) (*UserCredential, error) {
	query := `
		SELECT id, user_id, credential_type, name, public_key, private_key_encrypted,
			fingerprint, metadata, created_at, updated_at, expires_at, is_default
		FROM user_credentials
		WHERE user_id = ? AND credential_type = ? AND is_default = 1
		LIMIT 1
	`

	var cred UserCredential
	var metadataJSON string
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID, credType).Scan(
		&cred.ID,
		&cred.UserID,
		&cred.Type,
		&cred.Name,
		&cred.PublicKey,
		&cred.PrivateKeyEncrypted,
		&cred.Fingerprint,
		&metadataJSON,
		&cred.CreatedAt,
		&cred.UpdatedAt,
		&expiresAt,
		&cred.IsDefault,
	)
	if err == sql.ErrNoRows {
		return nil, ErrCredentialNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default user credential: %w", err)
	}

	if expiresAt.Valid {
		cred.ExpiresAt = &expiresAt.Time
	}

	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &cred.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &cred, nil
}

// Project Credentials

// CreateProjectCredential creates a new project credential
func (r *Repository) CreateProjectCredential(ctx context.Context, cred *ProjectCredential) error {
	metadataJSON, err := json.Marshal(cred.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO project_credentials (
			project_id, user_id, credential_type, name, description, value_encrypted,
			metadata, expires_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		cred.ProjectID,
		cred.UserID,
		cred.Type,
		cred.Name,
		cred.Description,
		cred.ValueEncrypted,
		string(metadataJSON),
		cred.ExpiresAt,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create project credential: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted ID: %w", err)
	}

	cred.ID = int(id)
	cred.CreatedAt = time.Now()
	cred.UpdatedAt = time.Now()

	return nil
}

// GetProjectCredential retrieves a project credential by ID
func (r *Repository) GetProjectCredential(ctx context.Context, id int) (*ProjectCredential, error) {
	query := `
		SELECT id, project_id, user_id, credential_type, name, description, value_encrypted,
			metadata, created_at, updated_at, expires_at, last_used_at
		FROM project_credentials
		WHERE id = ?
	`

	var cred ProjectCredential
	var metadataJSON string
	var expiresAt, lastUsedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cred.ID,
		&cred.ProjectID,
		&cred.UserID,
		&cred.Type,
		&cred.Name,
		&cred.Description,
		&cred.ValueEncrypted,
		&metadataJSON,
		&cred.CreatedAt,
		&cred.UpdatedAt,
		&expiresAt,
		&lastUsedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrCredentialNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project credential: %w", err)
	}

	if expiresAt.Valid {
		cred.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		cred.LastUsedAt = &lastUsedAt.Time
	}

	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &cred.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &cred, nil
}

// ListProjectCredentials lists all credentials for a project
func (r *Repository) ListProjectCredentials(ctx context.Context, projectID int) ([]*ProjectCredential, error) {
	query := `
		SELECT id, project_id, user_id, credential_type, name, description, value_encrypted,
			metadata, created_at, updated_at, expires_at, last_used_at
		FROM project_credentials
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project credentials: %w", err)
	}
	defer rows.Close()

	var credentials []*ProjectCredential
	for rows.Next() {
		var cred ProjectCredential
		var metadataJSON string
		var expiresAt, lastUsedAt sql.NullTime

		err := rows.Scan(
			&cred.ID,
			&cred.ProjectID,
			&cred.UserID,
			&cred.Type,
			&cred.Name,
			&cred.Description,
			&cred.ValueEncrypted,
			&metadataJSON,
			&cred.CreatedAt,
			&cred.UpdatedAt,
			&expiresAt,
			&lastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project credential: %w", err)
		}

		if expiresAt.Valid {
			cred.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			cred.LastUsedAt = &lastUsedAt.Time
		}

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &cred.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		credentials = append(credentials, &cred)
	}

	return credentials, nil
}

// UpdateProjectCredential updates a project credential
func (r *Repository) UpdateProjectCredential(ctx context.Context, cred *ProjectCredential) error {
	metadataJSON, err := json.Marshal(cred.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE project_credentials
		SET name = ?, description = ?, value_encrypted = ?,
			metadata = ?, expires_at = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContext(ctx, query,
		cred.Name,
		cred.Description,
		cred.ValueEncrypted,
		string(metadataJSON),
		cred.ExpiresAt,
		time.Now(),
		cred.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update project credential: %w", err)
	}

	cred.UpdatedAt = time.Now()
	return nil
}

// DeleteProjectCredential deletes a project credential
func (r *Repository) DeleteProjectCredential(ctx context.Context, id int) error {
	query := `DELETE FROM project_credentials WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project credential: %w", err)
	}
	return nil
}

// UpdateLastUsed updates the last_used_at timestamp for a credential
func (r *Repository) UpdateLastUsed(ctx context.Context, credentialID int) error {
	query := `UPDATE project_credentials SET last_used_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), credentialID)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	return nil
}
