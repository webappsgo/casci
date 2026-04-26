package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// Manager handles build workspace operations
type Manager struct {
	baseDir string
}

// NewManager creates a new workspace manager
func NewManager(baseDir string) *Manager {
	if baseDir == "" {
		baseDir = "/var/lib/casci/workspaces"
	}
	return &Manager{
		baseDir: baseDir,
	}
}

// Create creates a new workspace for a build
func (m *Manager) Create(projectID, buildID int) (string, error) {
	workspacePath := m.GetPath(projectID, buildID)

	// Create workspace directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace: %w", err)
	}

	return workspacePath, nil
}

// GetPath returns the workspace path for a build
func (m *Manager) GetPath(projectID, buildID int) string {
	return filepath.Join(m.baseDir, fmt.Sprintf("project-%d", projectID), fmt.Sprintf("build-%d", buildID))
}

// Cleanup removes a workspace
func (m *Manager) Cleanup(projectID, buildID int) error {
	workspacePath := m.GetPath(projectID, buildID)

	if err := os.RemoveAll(workspacePath); err != nil {
		return fmt.Errorf("failed to cleanup workspace: %w", err)
	}

	return nil
}

// Exists checks if a workspace exists
func (m *Manager) Exists(projectID, buildID int) bool {
	workspacePath := m.GetPath(projectID, buildID)
	_, err := os.Stat(workspacePath)
	return err == nil
}

// GetSize returns the size of a workspace in bytes
func (m *Manager) GetSize(projectID, buildID int) (int64, error) {
	workspacePath := m.GetPath(projectID, buildID)

	var size int64
	err := filepath.Walk(workspacePath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// CleanupOld removes workspaces older than the specified age
func (m *Manager) CleanupOld(maxAge int64) error {
	// Walk through all project directories
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(m.baseDir, entry.Name())
		buildDirs, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		for _, buildDir := range buildDirs {
			if !buildDir.IsDir() {
				continue
			}

			buildPath := filepath.Join(projectPath, buildDir.Name())
			info, err := os.Stat(buildPath)
			if err != nil {
				continue
			}

			// Check age
			if info.ModTime().Unix() < maxAge {
				os.RemoveAll(buildPath)
			}
		}
	}

	return nil
}
