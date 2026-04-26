package credentials

import (
	"time"
)

// CredentialType defines the type of credential
type CredentialType string

const (
	// User credential types
	CredentialTypeGPG         CredentialType = "gpg"
	CredentialTypeSSH         CredentialType = "ssh"
	CredentialTypeSigningCert CredentialType = "signing_cert"

	// Project credential types
	CredentialTypeGitToken       CredentialType = "git_token"
	CredentialTypeDockerRegistry CredentialType = "docker_registry"
	CredentialTypeAWSKeys        CredentialType = "aws_keys"
	CredentialTypeGCPKeys        CredentialType = "gcp_keys"
	CredentialTypeAzureKeys      CredentialType = "azure_keys"
	CredentialTypeSSHKey         CredentialType = "ssh_key"
	CredentialTypeAPIToken       CredentialType = "api_token"
	CredentialTypePassword       CredentialType = "password"
	CredentialTypeSecret         CredentialType = "secret"
	CredentialTypeCertificate    CredentialType = "certificate"
)

// UserCredential represents a user's cryptographic credential
type UserCredential struct {
	ID                   int               `json:"id"`
	UserID               int               `json:"user_id"`
	Type                 CredentialType    `json:"type"`
	Name                 string            `json:"name"`
	PublicKey            string            `json:"public_key,omitempty"`
	PrivateKeyEncrypted  string            `json:"-"` // Never expose in JSON
	Fingerprint          string            `json:"fingerprint,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	ExpiresAt            *time.Time        `json:"expires_at,omitempty"`
	IsDefault            bool              `json:"is_default"`
}

// ProjectCredential represents a project-specific credential
type ProjectCredential struct {
	ID              int               `json:"id"`
	ProjectID       int               `json:"project_id"`
	UserID          int               `json:"user_id"`
	Type            CredentialType    `json:"type"`
	Name            string            `json:"name"`
	Description     string            `json:"description,omitempty"`
	ValueEncrypted  string            `json:"-"` // Never expose in JSON
	Metadata        map[string]string `json:"metadata,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	ExpiresAt       *time.Time        `json:"expires_at,omitempty"`
	LastUsedAt      *time.Time        `json:"last_used_at,omitempty"`
}

// CreateUserCredentialRequest represents a request to create a user credential
type CreateUserCredentialRequest struct {
	Type       CredentialType    `json:"type"`
	Name       string            `json:"name"`
	PublicKey  string            `json:"public_key,omitempty"`
	PrivateKey string            `json:"private_key,omitempty"`
	Passphrase string            `json:"passphrase,omitempty"`
	Generate   bool              `json:"generate,omitempty"`
	KeySize    int               `json:"key_size,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	IsDefault  bool              `json:"is_default,omitempty"`
}

// CreateProjectCredentialRequest represents a request to create a project credential
type CreateProjectCredentialRequest struct {
	Type        CredentialType    `json:"type"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Value       string            `json:"value"`
	Username    string            `json:"username,omitempty"`
	Password    string            `json:"password,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// UpdateUserCredentialRequest represents a request to update a user credential
type UpdateUserCredentialRequest struct {
	Name       string            `json:"name,omitempty"`
	PublicKey  string            `json:"public_key,omitempty"`
	PrivateKey string            `json:"private_key,omitempty"`
	Passphrase string            `json:"passphrase,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	IsDefault  bool              `json:"is_default,omitempty"`
}

// UpdateProjectCredentialRequest represents a request to update a project credential
type UpdateProjectCredentialRequest struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Value       string            `json:"value,omitempty"`
	Username    string            `json:"username,omitempty"`
	Password    string            `json:"password,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// GPGKey represents a GPG key pair
type GPGKey struct {
	PublicKey   string
	PrivateKey  string
	Fingerprint string
	KeyID       string
	UserID      string
}

// SSHKey represents an SSH key pair
type SSHKey struct {
	PublicKey   string
	PrivateKey  string
	Fingerprint string
	KeyType     string // rsa, ed25519, etc.
}

// SigningCertificate represents a code signing certificate
type SigningCertificate struct {
	Certificate string
	PrivateKey  string
	Chain       []string
	Fingerprint string
	Subject     string
	Issuer      string
	ValidFrom   time.Time
	ValidTo     time.Time
}

// CredentialUsage represents credential usage tracking
type CredentialUsage struct {
	CredentialID int       `json:"credential_id"`
	BuildID      int       `json:"build_id"`
	UsedAt       time.Time `json:"used_at"`
	Purpose      string    `json:"purpose"`
}

// Validate validates a create user credential request
func (r *CreateUserCredentialRequest) Validate() error {
	if r.Type == "" {
		return ErrInvalidCredentialType
	}

	if r.Name == "" {
		return ErrCredentialNameRequired
	}

	if !r.Generate && r.PublicKey == "" && r.PrivateKey == "" {
		return ErrCredentialValueRequired
	}

	return nil
}

// Validate validates a create project credential request
func (r *CreateProjectCredentialRequest) Validate() error {
	if r.Type == "" {
		return ErrInvalidCredentialType
	}

	if r.Name == "" {
		return ErrCredentialNameRequired
	}

	if r.Value == "" && r.Password == "" {
		return ErrCredentialValueRequired
	}

	return nil
}

// IsExpired checks if a credential is expired
func (c *UserCredential) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// IsExpired checks if a project credential is expired
func (c *ProjectCredential) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// Errors
var (
	ErrInvalidCredentialType   = NewError("invalid credential type")
	ErrCredentialNameRequired  = NewError("credential name is required")
	ErrCredentialValueRequired = NewError("credential value is required")
	ErrCredentialNotFound      = NewError("credential not found")
	ErrCredentialExpired       = NewError("credential has expired")
	ErrUnauthorizedAccess      = NewError("unauthorized access to credential")
	ErrEncryptionFailed        = NewError("failed to encrypt credential")
	ErrDecryptionFailed        = NewError("failed to decrypt credential")
	ErrKeyGenerationFailed     = NewError("failed to generate key")
)

// Error represents a credential error
type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new credential error
func NewError(message string) error {
	return &Error{Message: message}
}
