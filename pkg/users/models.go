package users

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	APIToken     string    `json:"api_token,omitempty"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateUserRequest represents a user registration request
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// Validate validates a create user request
func (r *CreateUserRequest) Validate() error {
	if r.Username == "" {
		return ErrInvalidUsername
	}
	if len(r.Username) < 3 {
		return ErrUsernameTooShort
	}
	if r.Email == "" {
		return ErrInvalidEmail
	}
	if len(r.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

// Validate validates a login request
func (r *LoginRequest) Validate() error {
	if r.Username == "" {
		return ErrInvalidUsername
	}
	if r.Password == "" {
		return ErrInvalidPassword
	}
	return nil
}
