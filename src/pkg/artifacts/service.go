package artifacts

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/casapps/casci/src/pkg/database"
)

// Service handles artifact business logic
type Service struct {
	db      *database.Database
	storage Storage
	config  *StorageConfig
	dedupDB map[string]string // hash -> storage_path for deduplication
}

// NewService creates a new artifact service
func NewService(db *database.Database, config *StorageConfig) *Service {
	var storage Storage

	switch config.Type {
	case "s3":
		storage = NewS3Storage(config.S3Bucket, config.S3Region, config.S3AccessKey, config.S3SecretKey)
	case "gcs":
		storage = NewGCSStorage(config.GCSBucket, config.GCSCredentials)
	case "azure":
		storage = NewAzureStorage(config.AzureContainer, config.AzureAccountName, config.AzureAccountKey)
	default:
		// Default to local storage
		if config.LocalPath == "" {
			config.LocalPath = "/var/lib/casci/artifacts"
		}
		storage = NewLocalStorage(config.LocalPath)
	}

	// Set defaults
	if config.RetentionDays == 0 {
		config.RetentionDays = 30
	}

	return &Service{
		db:      db,
		storage: storage,
		config:  config,
		dedupDB: make(map[string]string),
	}
}

// Store stores an artifact
func (s *Service) Store(ctx context.Context, buildID int, name string, data []byte) (*Artifact, error) {
	// Calculate hash for deduplication
	hash := CalculateHash(data)

	// Check for existing artifact with same hash (deduplication)
	var storagePath string
	if s.config.Deduplication {
		if existingPath, ok := s.dedupDB[hash]; ok {
			log.Printf("Artifact deduplicated: hash %s already exists", hash[:8])
			storagePath = existingPath
		}
	}

	// Compress if enabled
	var dataToStore []byte
	var compressed bool
	if s.config.Compression && storagePath == "" {
		var err error
		dataToStore, err = CompressData(data)
		if err != nil {
			log.Printf("Warning: failed to compress artifact: %v", err)
			dataToStore = data
		} else {
			compressed = true
			log.Printf("Compressed artifact: %d -> %d bytes (%.1f%%)",
				len(data), len(dataToStore),
				float64(len(dataToStore))/float64(len(data))*100)
		}
	} else {
		dataToStore = data
	}

	// Store data if not deduplicated
	if storagePath == "" {
		// Generate storage path: buildID/name
		storagePath = filepath.Join(fmt.Sprintf("build-%d", buildID), name)

		if err := s.storage.Store(ctx, storagePath, dataToStore); err != nil {
			return nil, fmt.Errorf("failed to store artifact: %w", err)
		}

		// Add to dedup database
		if s.config.Deduplication {
			s.dedupDB[hash] = storagePath
		}
	}

	// Create artifact record
	artifact := &Artifact{
		BuildID:     buildID,
		Name:        name,
		Path:        name,
		Size:        int64(len(data)), // Original size
		Hash:        hash,
		StorageType: s.config.Type,
		StoragePath: storagePath,
		Compressed:  compressed,
		CreatedAt:   time.Now(),
	}

	// Set expiry
	if s.config.RetentionDays > 0 {
		artifact.ExpiresAt = time.Now().AddDate(0, 0, s.config.RetentionDays)
	}

	// Save to database
	if err := s.createArtifactRecord(ctx, artifact); err != nil {
		return nil, fmt.Errorf("failed to create artifact record: %w", err)
	}

	log.Printf("Stored artifact: %s (build %d, size %d bytes, hash %s)",
		name, buildID, len(data), hash[:8])

	return artifact, nil
}

// Retrieve retrieves an artifact
func (s *Service) Retrieve(ctx context.Context, id int) (*Artifact, []byte, error) {
	// Get artifact record
	artifact, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// Retrieve data from storage
	data, err := s.storage.Retrieve(ctx, artifact.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve artifact: %w", err)
	}

	// Decompress if needed
	if artifact.Compressed {
		data, err = DecompressData(data)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decompress artifact: %w", err)
		}
	}

	return artifact, data, nil
}

// GetByID retrieves artifact metadata by ID
func (s *Service) GetByID(ctx context.Context, id int) (*Artifact, error) {
	query := `
		SELECT id, build_id, name, path, size, content_type, hash,
			storage_type, storage_path, compressed, expires_at, created_at
		FROM artifacts
		WHERE id = ?
	`

	var artifact Artifact
	var expiresAt *time.Time

	err := s.db.QueryRow(ctx, query, id).Scan(
		&artifact.ID,
		&artifact.BuildID,
		&artifact.Name,
		&artifact.Path,
		&artifact.Size,
		&artifact.ContentType,
		&artifact.Hash,
		&artifact.StorageType,
		&artifact.StoragePath,
		&artifact.Compressed,
		&expiresAt,
		&artifact.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("artifact not found: %w", err)
	}

	if expiresAt != nil {
		artifact.ExpiresAt = *expiresAt
	}

	return &artifact, nil
}

// ListByBuild retrieves all artifacts for a build
func (s *Service) ListByBuild(ctx context.Context, buildID int) ([]*Artifact, error) {
	query := `
		SELECT id, build_id, name, path, size, content_type, hash,
			storage_type, storage_path, compressed, expires_at, created_at
		FROM artifacts
		WHERE build_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(ctx, query, buildID)
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}
	defer rows.Close()

	var artifacts []*Artifact
	for rows.Next() {
		var artifact Artifact
		var expiresAt *time.Time

		err := rows.Scan(
			&artifact.ID,
			&artifact.BuildID,
			&artifact.Name,
			&artifact.Path,
			&artifact.Size,
			&artifact.ContentType,
			&artifact.Hash,
			&artifact.StorageType,
			&artifact.StoragePath,
			&artifact.Compressed,
			&expiresAt,
			&artifact.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan artifact: %w", err)
		}

		if expiresAt != nil {
			artifact.ExpiresAt = *expiresAt
		}

		artifacts = append(artifacts, &artifact)
	}

	return artifacts, nil
}

// Delete deletes an artifact
func (s *Service) Delete(ctx context.Context, id int) error {
	// Get artifact record
	artifact, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if other artifacts use the same storage path (deduplication)
	if s.config.Deduplication {
		count, err := s.countArtifactsByStoragePath(ctx, artifact.StoragePath)
		if err != nil {
			return err
		}

		// Only delete from storage if this is the last reference
		if count == 1 {
			if err := s.storage.Delete(ctx, artifact.StoragePath); err != nil {
				log.Printf("Warning: failed to delete artifact from storage: %v", err)
			}

			// Remove from dedup database
			delete(s.dedupDB, artifact.Hash)
		}
	} else {
		// Delete from storage
		if err := s.storage.Delete(ctx, artifact.StoragePath); err != nil {
			log.Printf("Warning: failed to delete artifact from storage: %v", err)
		}
	}

	// Delete record from database
	query := `DELETE FROM artifacts WHERE id = ?`
	_, err = s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete artifact record: %w", err)
	}

	return nil
}

// CleanupExpired removes expired artifacts
func (s *Service) CleanupExpired(ctx context.Context) error {
	query := `
		SELECT id FROM artifacts
		WHERE expires_at IS NOT NULL AND expires_at < ?
	`

	rows, err := s.db.Query(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to query expired artifacts: %w", err)
	}
	defer rows.Close()

	var expiredIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		expiredIDs = append(expiredIDs, id)
	}

	log.Printf("Found %d expired artifacts to cleanup", len(expiredIDs))

	for _, id := range expiredIDs {
		if err := s.Delete(ctx, id); err != nil {
			log.Printf("Warning: failed to delete expired artifact %d: %v", id, err)
		}
	}

	return nil
}

// createArtifactRecord creates an artifact record in the database
func (s *Service) createArtifactRecord(ctx context.Context, artifact *Artifact) error {
	query := `
		INSERT INTO artifacts (build_id, name, path, size, content_type, hash,
			storage_type, storage_path, compressed, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var expiresAt interface{}
	if !artifact.ExpiresAt.IsZero() {
		expiresAt = artifact.ExpiresAt
	}

	result, err := s.db.Exec(ctx, query,
		artifact.BuildID,
		artifact.Name,
		artifact.Path,
		artifact.Size,
		artifact.ContentType,
		artifact.Hash,
		artifact.StorageType,
		artifact.StoragePath,
		artifact.Compressed,
		expiresAt,
		artifact.CreatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	artifact.ID = int(id)
	return nil
}

// countArtifactsByStoragePath counts artifacts using the same storage path
func (s *Service) countArtifactsByStoragePath(ctx context.Context, storagePath string) (int, error) {
	query := `SELECT COUNT(*) FROM artifacts WHERE storage_path = ?`

	var count int
	err := s.db.QueryRow(ctx, query, storagePath).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetStorageStats returns storage statistics
func (s *Service) GetStorageStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_artifacts,
			SUM(size) as total_size,
			COUNT(DISTINCT storage_path) as unique_files
		FROM artifacts
	`

	var totalArtifacts, totalSize, uniqueFiles int64
	err := s.db.QueryRow(ctx, query).Scan(&totalArtifacts, &totalSize, &uniqueFiles)
	if err != nil {
		return nil, err
	}

	dedupRatio := float64(0)
	if totalArtifacts > 0 {
		dedupRatio = (1 - float64(uniqueFiles)/float64(totalArtifacts)) * 100
	}

	stats := map[string]interface{}{
		"total_artifacts":  totalArtifacts,
		"total_size":       totalSize,
		"unique_files":     uniqueFiles,
		"dedup_ratio":      fmt.Sprintf("%.1f%%", dedupRatio),
		"storage_type":     s.config.Type,
		"compression":      s.config.Compression,
		"deduplication":    s.config.Deduplication,
		"retention_days":   s.config.RetentionDays,
	}

	return stats, nil
}
