package users

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/casapps/casci/pkg/database"
)

// Repository handles user data operations
type Repository struct {
	db *database.Database
}

// NewRepository creates a new user repository
func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

// Create creates a new user
func (r *Repository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (username, email, password_hash, api_token, is_admin, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	// Adjust query for PostgreSQL
	if r.db.GetType() == "postgres" {
		query = `
			INSERT INTO users (username, email, password_hash, api_token, is_admin, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
	}

	user.CreatedAt = time.Now()

	if r.db.GetType() == "postgres" {
		err := r.db.QueryRow(ctx, query,
			user.Username,
			user.Email,
			user.PasswordHash,
			user.APIToken,
			user.IsAdmin,
			user.CreatedAt,
		).Scan(&user.ID)
		return err
	}

	result, err := r.db.Exec(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.APIToken,
		user.IsAdmin,
		user.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	user.ID = int(id)
	return nil
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(ctx context.Context, id int) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, api_token, is_admin, created_at
		FROM users
		WHERE id = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, username, email, password_hash, api_token, is_admin, created_at
			FROM users
			WHERE id = $1
		`
	}

	user := &User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.APIToken,
		&user.IsAdmin,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *Repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, api_token, is_admin, created_at
		FROM users
		WHERE username = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, username, email, password_hash, api_token, is_admin, created_at
			FROM users
			WHERE username = $1
		`
	}

	user := &User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.APIToken,
		&user.IsAdmin,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, api_token, is_admin, created_at
		FROM users
		WHERE email = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, username, email, password_hash, api_token, is_admin, created_at
			FROM users
			WHERE email = $1
		`
	}

	user := &User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.APIToken,
		&user.IsAdmin,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByAPIToken retrieves a user by API token
func (r *Repository) GetByAPIToken(ctx context.Context, token string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, api_token, is_admin, created_at
		FROM users
		WHERE api_token = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			SELECT id, username, email, password_hash, api_token, is_admin, created_at
			FROM users
			WHERE api_token = $1
		`
	}

	user := &User{}
	err := r.db.QueryRow(ctx, query, token).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.APIToken,
		&user.IsAdmin,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// List retrieves all users
func (r *Repository) List(ctx context.Context) ([]*User, error) {
	query := `
		SELECT id, username, email, password_hash, api_token, is_admin, created_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.APIToken,
			&user.IsAdmin,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// Update updates a user
func (r *Repository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET email = ?, password_hash = ?, api_token = ?, is_admin = ?
		WHERE id = ?
	`

	if r.db.GetType() == "postgres" {
		query = `
			UPDATE users
			SET email = $1, password_hash = $2, api_token = $3, is_admin = $4
			WHERE id = $5
		`
	}

	_, err := r.db.Exec(ctx, query,
		user.Email,
		user.PasswordHash,
		user.APIToken,
		user.IsAdmin,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user
func (r *Repository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = ?`

	if r.db.GetType() == "postgres" {
		query = `DELETE FROM users WHERE id = $1`
	}

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Count returns the total number of users
func (r *Repository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// UsernameExists checks if a username already exists
func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = ?`

	if r.db.GetType() == "postgres" {
		query = `SELECT COUNT(*) FROM users WHERE username = $1`
	}

	var count int
	err := r.db.QueryRow(ctx, query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username: %w", err)
	}

	return count > 0, nil
}

// EmailExists checks if an email already exists
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email = ?`

	if r.db.GetType() == "postgres" {
		query = `SELECT COUNT(*) FROM users WHERE email = $1`
	}

	var count int
	err := r.db.QueryRow(ctx, query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email: %w", err)
	}

	return count > 0, nil
}
