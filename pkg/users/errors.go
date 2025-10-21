package users

import "errors"

var (
	// Validation errors
	ErrInvalidUsername  = errors.New("invalid username")
	ErrUsernameTooShort = errors.New("username must be at least 3 characters")
	ErrInvalidEmail     = errors.New("invalid email address")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")

	// Database errors
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
	ErrEmailExists  = errors.New("email already registered")

	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")

	// Session errors
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)
