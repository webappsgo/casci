package artifacts

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Storage interface for artifact storage backends
type Storage interface {
	Store(ctx context.Context, path string, data []byte) error
	Retrieve(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	GetURL(ctx context.Context, path string, expiryMinutes int) (string, error)
}

// LocalStorage implements local filesystem storage
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage backend
func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

// Store stores data to local filesystem
func (ls *LocalStorage) Store(ctx context.Context, path string, data []byte) error {
	fullPath := filepath.Join(ls.basePath, path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Retrieve retrieves data from local filesystem
func (ls *LocalStorage) Retrieve(ctx context.Context, path string) ([]byte, error) {
	fullPath := filepath.Join(ls.basePath, path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// Delete deletes data from local filesystem
func (ls *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(ls.basePath, path)

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if a file exists
func (ls *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(ls.basePath, path)

	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetURL returns a local file path (not a URL)
func (ls *LocalStorage) GetURL(ctx context.Context, path string, expiryMinutes int) (string, error) {
	fullPath := filepath.Join(ls.basePath, path)
	return fullPath, nil
}

// CompressData compresses data using gzip
func CompressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	if _, err := gzWriter.Write(data); err != nil {
		return nil, fmt.Errorf("failed to compress: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close compressor: %w", err)
	}

	return buf.Bytes(), nil
}

// DecompressData decompresses gzip data
func DecompressData(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	gzReader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create decompressor: %w", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}

	return decompressed, nil
}

// CalculateHash calculates SHA256 hash of data
func CalculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// S3Storage implements S3-compatible storage
type S3Storage struct {
	bucket    string
	region    string
	accessKey string
	secretKey string
}

// NewS3Storage creates a new S3 storage backend
func NewS3Storage(bucket, region, accessKey, secretKey string) *S3Storage {
	return &S3Storage{
		bucket:    bucket,
		region:    region,
		accessKey: accessKey,
		secretKey: secretKey,
	}
}

// Store stores data to S3 (stub - requires AWS SDK)
func (s3 *S3Storage) Store(ctx context.Context, path string, data []byte) error {
	// TODO: Implement S3 upload using AWS SDK
	return fmt.Errorf("S3 storage not yet implemented")
}

// Retrieve retrieves data from S3 (stub)
func (s3 *S3Storage) Retrieve(ctx context.Context, path string) ([]byte, error) {
	// TODO: Implement S3 download using AWS SDK
	return nil, fmt.Errorf("S3 storage not yet implemented")
}

// Delete deletes data from S3 (stub)
func (s3 *S3Storage) Delete(ctx context.Context, path string) error {
	// TODO: Implement S3 delete using AWS SDK
	return fmt.Errorf("S3 storage not yet implemented")
}

// Exists checks if object exists in S3 (stub)
func (s3 *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	// TODO: Implement S3 head object using AWS SDK
	return false, fmt.Errorf("S3 storage not yet implemented")
}

// GetURL generates a pre-signed URL for S3 object (stub)
func (s3 *S3Storage) GetURL(ctx context.Context, path string, expiryMinutes int) (string, error) {
	// TODO: Implement S3 pre-signed URL using AWS SDK
	return "", fmt.Errorf("S3 storage not yet implemented")
}

// GCSStorage implements Google Cloud Storage
type GCSStorage struct {
	bucket      string
	credentials string
}

// NewGCSStorage creates a new GCS storage backend
func NewGCSStorage(bucket, credentials string) *GCSStorage {
	return &GCSStorage{
		bucket:      bucket,
		credentials: credentials,
	}
}

// Store stores data to GCS (stub)
func (gcs *GCSStorage) Store(ctx context.Context, path string, data []byte) error {
	// TODO: Implement GCS upload
	return fmt.Errorf("GCS storage not yet implemented")
}

// Retrieve retrieves data from GCS (stub)
func (gcs *GCSStorage) Retrieve(ctx context.Context, path string) ([]byte, error) {
	// TODO: Implement GCS download
	return nil, fmt.Errorf("GCS storage not yet implemented")
}

// Delete deletes data from GCS (stub)
func (gcs *GCSStorage) Delete(ctx context.Context, path string) error {
	// TODO: Implement GCS delete
	return fmt.Errorf("GCS storage not yet implemented")
}

// Exists checks if object exists in GCS (stub)
func (gcs *GCSStorage) Exists(ctx context.Context, path string) (bool, error) {
	// TODO: Implement GCS exists check
	return false, fmt.Errorf("GCS storage not yet implemented")
}

// GetURL generates a signed URL for GCS object (stub)
func (gcs *GCSStorage) GetURL(ctx context.Context, path string, expiryMinutes int) (string, error) {
	// TODO: Implement GCS signed URL
	return "", fmt.Errorf("GCS storage not yet implemented")
}

// AzureStorage implements Azure Blob Storage
type AzureStorage struct {
	container   string
	accountName string
	accountKey  string
}

// NewAzureStorage creates a new Azure storage backend
func NewAzureStorage(container, accountName, accountKey string) *AzureStorage {
	return &AzureStorage{
		container:   container,
		accountName: accountName,
		accountKey:  accountKey,
	}
}

// Store stores data to Azure Blob Storage (stub)
func (az *AzureStorage) Store(ctx context.Context, path string, data []byte) error {
	// TODO: Implement Azure upload
	return fmt.Errorf("Azure storage not yet implemented")
}

// Retrieve retrieves data from Azure Blob Storage (stub)
func (az *AzureStorage) Retrieve(ctx context.Context, path string) ([]byte, error) {
	// TODO: Implement Azure download
	return nil, fmt.Errorf("Azure storage not yet implemented")
}

// Delete deletes data from Azure Blob Storage (stub)
func (az *AzureStorage) Delete(ctx context.Context, path string) error {
	// TODO: Implement Azure delete
	return fmt.Errorf("Azure storage not yet implemented")
}

// Exists checks if blob exists in Azure (stub)
func (az *AzureStorage) Exists(ctx context.Context, path string) (bool, error) {
	// TODO: Implement Azure exists check
	return false, fmt.Errorf("Azure storage not yet implemented")
}

// GetURL generates a SAS URL for Azure blob (stub)
func (az *AzureStorage) GetURL(ctx context.Context, path string, expiryMinutes int) (string, error) {
	// TODO: Implement Azure SAS URL
	return "", fmt.Errorf("Azure storage not yet implemented")
}
