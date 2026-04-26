package artifacts

import (
	"time"
)

// Artifact represents a build artifact
type Artifact struct {
	ID          int       `json:"id"`
	BuildID     int       `json:"build_id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	Hash        string    `json:"hash"` // SHA256 for deduplication
	StorageType string    `json:"storage_type"` // local, s3, gcs, azure
	StoragePath string    `json:"storage_path"`
	Compressed  bool      `json:"compressed"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// StorageConfig represents artifact storage configuration
type StorageConfig struct {
	Type             string `json:"type"` // local, s3, gcs, azure
	LocalPath        string `json:"local_path,omitempty"`
	S3Bucket         string `json:"s3_bucket,omitempty"`
	S3Region         string `json:"s3_region,omitempty"`
	S3AccessKey      string `json:"s3_access_key,omitempty"`
	S3SecretKey      string `json:"s3_secret_key,omitempty"`
	GCSBucket        string `json:"gcs_bucket,omitempty"`
	GCSCredentials   string `json:"gcs_credentials,omitempty"`
	AzureContainer   string `json:"azure_container,omitempty"`
	AzureAccountName string `json:"azure_account_name,omitempty"`
	AzureAccountKey  string `json:"azure_account_key,omitempty"`
	RetentionDays    int    `json:"retention_days"` // Default 30
	Compression      bool   `json:"compression"`    // Default true
	Deduplication    bool   `json:"deduplication"`  // Default true
}

// ArtifactMetadata represents metadata for an artifact
type ArtifactMetadata struct {
	BuildID     int               `json:"build_id"`
	BuildNumber int               `json:"build_number"`
	ProjectName string            `json:"project_name"`
	Branch      string            `json:"branch"`
	CommitSHA   string            `json:"commit_sha"`
	Timestamp   time.Time         `json:"timestamp"`
	Artifacts   []ArtifactInfo    `json:"artifacts"`
	Environment map[string]string `json:"environment"`
}

// ArtifactInfo represents basic artifact information
type ArtifactInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Hash        string `json:"hash"`
}

// UploadRequest represents an artifact upload request
type UploadRequest struct {
	BuildID  int    `json:"build_id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Contents []byte `json:"-"` // Binary data
}

// DownloadResponse represents an artifact download response
type DownloadResponse struct {
	Artifact *Artifact `json:"artifact"`
	Contents []byte    `json:"-"` // Binary data
	URL      string    `json:"url,omitempty"` // Pre-signed URL for cloud storage
}
