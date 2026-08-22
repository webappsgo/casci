package users

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/webappsgo/casci/src/pkg/database"
)

// Argon2id parameters (AI.md PART 11 - all passwords hashed with Argon2id)
const (
	argonMemory      uint32 = 64 * 1024
	argonTime        uint32 = 3
	argonParallelism uint8  = 4
	argonSaltLength         = 16
	argonKeyLength   uint32 = 32
)

// ErrInvalidPasswordHash is returned when a stored hash is not a valid
// Argon2id PHC-format string
var ErrInvalidPasswordHash = errors.New("invalid password hash format")

// Service handles user business logic
type Service struct {
	repo *Repository
}

// NewService creates a new user service
func NewService(db *database.Database) *Service {
	return &Service{
		repo: NewRepository(db),
	}
}

// Register registers a new user
func (s *Service) Register(ctx context.Context, req *CreateUserRequest) (*User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if username exists
	exists, err := s.repo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// Check if email exists
	exists, err = s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password
	passwordHash, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate API token
	apiToken, err := s.GenerateAPIToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API token: %w", err)
	}

	// Check if this is the first user (becomes admin)
	count, err := s.Count(ctx)
	if err != nil {
		return nil, err
	}
	isFirstUser := count == 0

	// Create user
	user := &User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		APIToken:     apiToken,
		IsAdmin:      isFirstUser,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get user by username
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		if err == ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := s.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (s *Service) GetByID(ctx context.Context, id int) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByUsername retrieves a user by username
func (s *Service) GetByUsername(ctx context.Context, username string) (*User, error) {
	return s.repo.GetByUsername(ctx, username)
}

// GetByAPIToken retrieves a user by API token
func (s *Service) GetByAPIToken(ctx context.Context, token string) (*User, error) {
	return s.repo.GetByAPIToken(ctx, token)
}

// List retrieves all users
func (s *Service) List(ctx context.Context) ([]*User, error) {
	return s.repo.List(ctx)
}

// Update updates a user
func (s *Service) Update(ctx context.Context, id int, req *UpdateUserRequest) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update email if provided
	if req.Email != "" {
		user.Email = req.Email
	}

	// Update password if provided
	if req.Password != "" {
		if len(req.Password) < 8 {
			return nil, ErrPasswordTooShort
		}
		passwordHash, err := s.HashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = passwordHash
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Delete deletes a user
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// RegenerateAPIToken generates a new API token for a user
func (s *Service) RegenerateAPIToken(ctx context.Context, id int) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	apiToken, err := s.GenerateAPIToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API token: %w", err)
	}

	user.APIToken = apiToken

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// HashPassword hashes a password using Argon2id and returns a PHC-format
// encoded string: $argon2id$v=19$m=65536,t=3,p=4$<b64salt>$<b64hash>
func (s *Service) HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate password salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonParallelism, argonKeyLength)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonTime,
		argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)

	return encoded, nil
}

// VerifyPassword verifies a password against an Argon2id PHC-format hash
// using a constant-time comparison
func (s *Service) VerifyPassword(hash, password string) error {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return ErrInvalidPasswordHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return ErrInvalidPasswordHash
	}
	if version != argon2.Version {
		return ErrInvalidPasswordHash
	}

	var memory, timeCost uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &parallelism); err != nil {
		return ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return ErrInvalidPasswordHash
	}

	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return ErrInvalidPasswordHash
	}

	computed := argon2.IDKey([]byte(password), salt, timeCost, memory, parallelism, uint32(len(expected)))
	if subtle.ConstantTimeCompare(computed, expected) != 1 {
		return ErrInvalidCredentials
	}

	return nil
}

// GenerateAPIToken generates a random API token
func (s *Service) GenerateAPIToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "casci_" + hex.EncodeToString(bytes), nil
}

// IsAdmin checks if a user is an admin
func (s *Service) IsAdmin(user *User) bool {
	return user != nil && user.IsAdmin
}

// MakeAdmin makes a user an admin
func (s *Service) MakeAdmin(ctx context.Context, id int) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.IsAdmin = true
	return s.repo.Update(ctx, user)
}

// RevokeAdmin revokes admin privileges from a user
// Count returns the total number of registered users
func (s *Service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func (s *Service) RevokeAdmin(ctx context.Context, id int) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Cannot revoke admin from the first user
	count, err := s.Count(ctx)
	if err != nil {
		return err
	}
	if count == 1 {
		return fmt.Errorf("cannot revoke admin from the only user")
	}

	user.IsAdmin = false
	return s.repo.Update(ctx, user)
}
