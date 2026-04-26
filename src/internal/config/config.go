package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
)

// Config represents the complete configuration
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	EncryptionKey string
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host         string
	Port         int
	TLSEnabled   bool
	TLSCertPath  string
	TLSKeyPath   string
	ReadTimeout  int // Note: No timeouts as per spec, but keeping for future
	WriteTimeout int
	Footer       FooterConfig
	CORS         string // CORS origins: "*", "https://example.com", or comma-separated list
}

// FooterConfig holds footer customization settings
type FooterConfig struct {
	TrackingID    string // Google Analytics tracking ID (empty = disabled)
	CookieConsent CookieConsentConfig
	CustomHTML    string // Custom branding HTML above application footer
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type        string // sqlite, postgres, mysql
	DSN         string // Connection string
	MaxConns    int
	MaxIdleConn int
	Replicas    []string
}

// CookieConsentConfig holds cookie consent popup settings
type CookieConsentConfig struct {
	Enabled    bool
	Message    string
	PolicyURL  string
}

// Load loads configuration from environment or uses defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server:       loadServerConfig(),
		Database:     loadDatabaseConfig(),
		EncryptionKey: loadEncryptionKey(),
	}

	// Ensure required directories exist
	if err := ensureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	return cfg, nil
}

// loadServerConfig loads server configuration
func loadServerConfig() ServerConfig {
	cfg := ServerConfig{
		Host:         getEnv("CASCI_HOST", "0.0.0.0"),
		Port:         randomPort(),
		TLSEnabled:   getEnvBool("CASCI_TLS_ENABLED", false),
		TLSCertPath:  getEnv("CASCI_TLS_CERT", "/var/lib/casci/certs/server.crt"),
		TLSKeyPath:   getEnv("CASCI_TLS_KEY", "/var/lib/casci/certs/server.key"),
		ReadTimeout:  0,
		WriteTimeout: 0,
		Footer: FooterConfig{
			TrackingID: getEnv("CASCI_TRACKING_ID", ""),
			CookieConsent: CookieConsentConfig{
				Enabled:   getEnvBool("CASCI_COOKIE_CONSENT", true),
				Message:   getEnv("CASCI_COOKIE_MESSAGE", "This site uses cookies for functionality and analytics."),
				PolicyURL: getEnv("CASCI_COOKIE_POLICY_URL", "/server/privacy"),
			},
			CustomHTML: getEnv("CASCI_FOOTER_HTML", ""),
		},
		CORS: getEnv("CASCI_CORS_ORIGINS", "*"),
	}

	// Check if port is explicitly set
	if port := os.Getenv("CASCI_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Port)
	}

	return cfg
}

// loadDatabaseConfig loads database configuration
func loadDatabaseConfig() DatabaseConfig {
	dbType := getEnv("CASCI_DB_TYPE", "sqlite")
	dsn := getEnv("CASCI_DB_DSN", "")

	// Default DSN for SQLite
	if dsn == "" && dbType == "sqlite" {
		dsn = filepath.Join(getDataDir(), "casci.db")
	}

	return DatabaseConfig{
		Type:        dbType,
		DSN:         dsn,
		MaxConns:    getEnvInt("CASCI_DB_MAX_CONNS", 25),
		MaxIdleConn: getEnvInt("CASCI_DB_MAX_IDLE", 5),
		Replicas:    []string{}, // TODO: Parse from env
	}
}

// loadEncryptionKey loads or generates the master encryption key
func loadEncryptionKey() string {
	// Try environment variable first
	if key := os.Getenv("CASCI_ENCRYPTION_KEY"); key != "" {
		return key
	}

	// Try loading from file
	keyPath := filepath.Join(getDataDir(), "encryption.key")
	if data, err := os.ReadFile(keyPath); err == nil {
		return string(data)
	}

	// Generate new key
	key := generateEncryptionKey()

	// Save to file
	os.WriteFile(keyPath, []byte(key), 0600)

	return key
}

// generateEncryptionKey generates a new random encryption key
func generateEncryptionKey() string {
	key := make([]byte, 32) // 256 bits
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		// Fallback to a deterministic key (not recommended for production)
		return "casci-default-encryption-key-change-this-in-production"
	}
	return base64.StdEncoding.EncodeToString(key)
}

// randomPort generates a random port in the range 64000-64999
func randomPort() int {
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return 64500 // Fallback to middle of range
	}
	return 64000 + int(n.Int64())
}

// ensureDirectories creates required directories if they don't exist
func ensureDirectories() error {
	dirs := []string{
		getDataDir(),
		getLogDir(),
		filepath.Join(getDataDir(), "scanners"),
		filepath.Join(getDataDir(), "artifacts"),
		filepath.Join(getDataDir(), "workspaces"),
		filepath.Join(getDataDir(), "cache"),
		filepath.Join(getDataDir(), "certs"),
		filepath.Join(getLogDir(), "builds"),
		filepath.Join(getLogDir(), "users"),
		"/etc/casci/security/merged",
		"/etc/casci/security/cache",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			// Try local directory if system directories fail
			if os.IsPermission(err) {
				localDir := filepath.Join(".", ".casci", filepath.Base(dir))
				if err := os.MkdirAll(localDir, 0755); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}

// Helper functions
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var i int
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}

func getDataDir() string {
	if dir := os.Getenv("CASCI_DATA_DIR"); dir != "" {
		return dir
	}
	if _, err := os.Stat("/var/lib/casci"); err == nil {
		return "/var/lib/casci"
	}
	return ".casci/data"
}

func getLogDir() string {
	if dir := os.Getenv("CASCI_LOG_DIR"); dir != "" {
		return dir
	}
	if _, err := os.Stat("/var/log/casci"); err == nil {
		return "/var/log/casci"
	}
	return ".casci/logs"
}
