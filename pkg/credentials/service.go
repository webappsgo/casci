package credentials

import (
	"context"
	"fmt"
)

// Service handles credential business logic
type Service struct {
	repo      *Repository
	encryptor *Encryptor
	keygen    *KeyGenerator
}

// NewService creates a new credential service
func NewService(repo *Repository, masterKey string) *Service {
	return &Service{
		repo:      repo,
		encryptor: NewEncryptor(masterKey),
		keygen:    NewKeyGenerator(),
	}
}

// User Credentials

// CreateUserCredential creates a new user credential
func (s *Service) CreateUserCredential(ctx context.Context, userID int, req *CreateUserCredentialRequest) (*UserCredential, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	cred := &UserCredential{
		UserID:    userID,
		Type:      req.Type,
		Name:      req.Name,
		Metadata:  req.Metadata,
		IsDefault: req.IsDefault,
	}

	// Generate or import key based on type
	var err error
	switch req.Type {
	case CredentialTypeGPG:
		err = s.handleGPGCredential(cred, req)
	case CredentialTypeSSH:
		err = s.handleSSHCredential(cred, req)
	case CredentialTypeSigningCert:
		err = s.handleSigningCertCredential(cred, req)
	default:
		return nil, ErrInvalidCredentialType
	}

	if err != nil {
		return nil, err
	}

	// Encrypt private key
	if cred.PrivateKeyEncrypted != "" {
		encrypted, err := s.encryptor.EncryptWithUserKey(cred.PrivateKeyEncrypted, userID)
		if err != nil {
			return nil, ErrEncryptionFailed
		}
		cred.PrivateKeyEncrypted = encrypted
	}

	// Set as default if requested or if it's the first credential of this type
	if req.IsDefault {
		// Unset other default credentials of the same type
		if err := s.unsetDefaultCredentials(ctx, userID, req.Type); err != nil {
			return nil, err
		}
	}

	// Create in database
	if err := s.repo.CreateUserCredential(ctx, cred); err != nil {
		return nil, err
	}

	// Don't return encrypted private key
	cred.PrivateKeyEncrypted = ""

	return cred, nil
}

// GetUserCredential retrieves a user credential
func (s *Service) GetUserCredential(ctx context.Context, id, userID int) (*UserCredential, error) {
	cred, err := s.repo.GetUserCredential(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if cred.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Don't return encrypted private key in normal responses
	cred.PrivateKeyEncrypted = ""

	return cred, nil
}

// GetUserCredentialWithPrivateKey retrieves a credential with decrypted private key
func (s *Service) GetUserCredentialWithPrivateKey(ctx context.Context, id, userID int) (*UserCredential, string, error) {
	cred, err := s.repo.GetUserCredential(ctx, id)
	if err != nil {
		return nil, "", err
	}

	// Verify ownership
	if cred.UserID != userID {
		return nil, "", ErrUnauthorizedAccess
	}

	// Check expiration
	if cred.IsExpired() {
		return nil, "", ErrCredentialExpired
	}

	// Decrypt private key
	privateKey := ""
	if cred.PrivateKeyEncrypted != "" {
		decrypted, err := s.encryptor.DecryptWithUserKey(cred.PrivateKeyEncrypted, userID)
		if err != nil {
			return nil, "", ErrDecryptionFailed
		}
		privateKey = decrypted
	}

	// Don't include encrypted key in response
	cred.PrivateKeyEncrypted = ""

	return cred, privateKey, nil
}

// ListUserCredentials lists all credentials for a user
func (s *Service) ListUserCredentials(ctx context.Context, userID int) ([]*UserCredential, error) {
	creds, err := s.repo.ListUserCredentials(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Remove encrypted private keys from response
	for _, cred := range creds {
		cred.PrivateKeyEncrypted = ""
	}

	return creds, nil
}

// UpdateUserCredential updates a user credential
func (s *Service) UpdateUserCredential(ctx context.Context, id, userID int, req *UpdateUserCredentialRequest) (*UserCredential, error) {
	cred, err := s.repo.GetUserCredential(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if cred.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Update fields
	if req.Name != "" {
		cred.Name = req.Name
	}
	if req.Metadata != nil {
		cred.Metadata = req.Metadata
	}
	if req.IsDefault {
		if err := s.unsetDefaultCredentials(ctx, userID, cred.Type); err != nil {
			return nil, err
		}
		cred.IsDefault = true
	}

	// Update keys if provided
	if req.PrivateKey != "" {
		encrypted, err := s.encryptor.EncryptWithUserKey(req.PrivateKey, userID)
		if err != nil {
			return nil, ErrEncryptionFailed
		}
		cred.PrivateKeyEncrypted = encrypted
	}
	if req.PublicKey != "" {
		cred.PublicKey = req.PublicKey
	}

	// Save
	if err := s.repo.UpdateUserCredential(ctx, cred); err != nil {
		return nil, err
	}

	cred.PrivateKeyEncrypted = ""
	return cred, nil
}

// DeleteUserCredential deletes a user credential
func (s *Service) DeleteUserCredential(ctx context.Context, id, userID int) error {
	cred, err := s.repo.GetUserCredential(ctx, id)
	if err != nil {
		return err
	}

	// Verify ownership
	if cred.UserID != userID {
		return ErrUnauthorizedAccess
	}

	return s.repo.DeleteUserCredential(ctx, id)
}

// Project Credentials

// CreateProjectCredential creates a new project credential
func (s *Service) CreateProjectCredential(ctx context.Context, projectID, userID int, req *CreateProjectCredentialRequest) (*ProjectCredential, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Combine username/password if both provided
	value := req.Value
	if req.Username != "" && req.Password != "" {
		value = fmt.Sprintf("%s:%s", req.Username, req.Password)
	} else if req.Password != "" {
		value = req.Password
	}

	// Encrypt value
	encrypted, err := s.encryptor.EncryptWithUserKey(value, userID)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	cred := &ProjectCredential{
		ProjectID:      projectID,
		UserID:         userID,
		Type:           req.Type,
		Name:           req.Name,
		Description:    req.Description,
		ValueEncrypted: encrypted,
		Metadata:       req.Metadata,
		ExpiresAt:      req.ExpiresAt,
	}

	if err := s.repo.CreateProjectCredential(ctx, cred); err != nil {
		return nil, err
	}

	// Don't return encrypted value
	cred.ValueEncrypted = ""

	return cred, nil
}

// GetProjectCredential retrieves a project credential
func (s *Service) GetProjectCredential(ctx context.Context, id, userID int) (*ProjectCredential, error) {
	cred, err := s.repo.GetProjectCredential(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if cred.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Don't return encrypted value
	cred.ValueEncrypted = ""

	return cred, nil
}

// GetProjectCredentialValue retrieves a credential with decrypted value
func (s *Service) GetProjectCredentialValue(ctx context.Context, id, userID int) (string, error) {
	cred, err := s.repo.GetProjectCredential(ctx, id)
	if err != nil {
		return "", err
	}

	// Verify ownership
	if cred.UserID != userID {
		return "", ErrUnauthorizedAccess
	}

	// Check expiration
	if cred.IsExpired() {
		return "", ErrCredentialExpired
	}

	// Decrypt value
	value, err := s.encryptor.DecryptWithUserKey(cred.ValueEncrypted, userID)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	// Update last used
	go s.repo.UpdateLastUsed(context.Background(), id)

	return value, nil
}

// ListProjectCredentials lists all credentials for a project
func (s *Service) ListProjectCredentials(ctx context.Context, projectID, userID int) ([]*ProjectCredential, error) {
	creds, err := s.repo.ListProjectCredentials(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Filter by user and remove encrypted values
	var result []*ProjectCredential
	for _, cred := range creds {
		if cred.UserID == userID {
			cred.ValueEncrypted = ""
			result = append(result, cred)
		}
	}

	return result, nil
}

// UpdateProjectCredential updates a project credential
func (s *Service) UpdateProjectCredential(ctx context.Context, id, userID int, req *UpdateProjectCredentialRequest) (*ProjectCredential, error) {
	cred, err := s.repo.GetProjectCredential(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if cred.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Update fields
	if req.Name != "" {
		cred.Name = req.Name
	}
	if req.Description != "" {
		cred.Description = req.Description
	}
	if req.Metadata != nil {
		cred.Metadata = req.Metadata
	}
	if req.ExpiresAt != nil {
		cred.ExpiresAt = req.ExpiresAt
	}

	// Update value if provided
	if req.Value != "" || req.Password != "" {
		value := req.Value
		if req.Username != "" && req.Password != "" {
			value = fmt.Sprintf("%s:%s", req.Username, req.Password)
		} else if req.Password != "" {
			value = req.Password
		}

		encrypted, err := s.encryptor.EncryptWithUserKey(value, userID)
		if err != nil {
			return nil, ErrEncryptionFailed
		}
		cred.ValueEncrypted = encrypted
	}

	// Save
	if err := s.repo.UpdateProjectCredential(ctx, cred); err != nil {
		return nil, err
	}

	cred.ValueEncrypted = ""
	return cred, nil
}

// DeleteProjectCredential deletes a project credential
func (s *Service) DeleteProjectCredential(ctx context.Context, id, userID int) error {
	cred, err := s.repo.GetProjectCredential(ctx, id)
	if err != nil {
		return err
	}

	// Verify ownership
	if cred.UserID != userID {
		return ErrUnauthorizedAccess
	}

	return s.repo.DeleteProjectCredential(ctx, id)
}

// Helper methods

func (s *Service) handleGPGCredential(cred *UserCredential, req *CreateUserCredentialRequest) error {
	if req.Generate {
		// Generate new GPG key
		name := req.Metadata["name"]
		email := req.Metadata["email"]
		keySize := req.KeySize
		if keySize == 0 {
			keySize = 4096
		}

		gpgKey, err := s.keygen.GenerateGPGKey(name, email, keySize)
		if err != nil {
			return ErrKeyGenerationFailed
		}

		cred.PublicKey = gpgKey.PublicKey
		cred.PrivateKeyEncrypted = gpgKey.PrivateKey
		cred.Fingerprint = gpgKey.Fingerprint
	} else {
		// Import existing key
		cred.PublicKey = req.PublicKey
		cred.PrivateKeyEncrypted = req.PrivateKey

		// Calculate fingerprint
		fingerprint, err := ParseGPGPublicKey(req.PublicKey)
		if err != nil {
			return fmt.Errorf("invalid GPG public key: %w", err)
		}
		cred.Fingerprint = fingerprint
	}

	return nil
}

func (s *Service) handleSSHCredential(cred *UserCredential, req *CreateUserCredentialRequest) error {
	if req.Generate {
		// Generate new SSH key
		keyType := "ed25519"
		if req.Metadata != nil && req.Metadata["key_type"] != "" {
			keyType = req.Metadata["key_type"]
		}

		sshKey, err := s.keygen.GenerateSSHKey(keyType)
		if err != nil {
			return ErrKeyGenerationFailed
		}

		cred.PublicKey = sshKey.PublicKey
		cred.PrivateKeyEncrypted = sshKey.PrivateKey
		cred.Fingerprint = sshKey.Fingerprint
	} else {
		// Import existing key
		cred.PublicKey = req.PublicKey
		cred.PrivateKeyEncrypted = req.PrivateKey

		// Calculate fingerprint
		fingerprint, err := ParseSSHPublicKey(req.PublicKey)
		if err != nil {
			return fmt.Errorf("invalid SSH public key: %w", err)
		}
		cred.Fingerprint = fingerprint
	}

	return nil
}

func (s *Service) handleSigningCertCredential(cred *UserCredential, req *CreateUserCredentialRequest) error {
	if req.Generate {
		// Generate new signing certificate
		name := req.Metadata["name"]
		org := req.Metadata["organization"]

		cert, err := s.keygen.GenerateSigningCertificate(name, org)
		if err != nil {
			return ErrKeyGenerationFailed
		}

		cred.PublicKey = cert.Certificate
		cred.PrivateKeyEncrypted = cert.PrivateKey
		cred.Fingerprint = cert.Fingerprint
	} else {
		// Import existing certificate
		cred.PublicKey = req.PublicKey
		cred.PrivateKeyEncrypted = req.PrivateKey
		// Fingerprint calculation would require parsing the certificate
		cred.Fingerprint = ""
	}

	return nil
}

func (s *Service) unsetDefaultCredentials(ctx context.Context, userID int, credType CredentialType) error {
	// Get all credentials of this type
	allCreds, err := s.repo.ListUserCredentials(ctx, userID)
	if err != nil {
		return err
	}

	// Unset default flag for all credentials of this type
	for _, c := range allCreds {
		if c.Type == credType && c.IsDefault {
			c.IsDefault = false
			if err := s.repo.UpdateUserCredential(ctx, c); err != nil {
				return err
			}
		}
	}

	return nil
}
