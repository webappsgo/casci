package config

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Build metadata, stamped into package main through -ldflags and published
// here so every package reports the same version without importing main.
// AI.md line 670 defines the ldflag targets as main.Version / main.CommitID /
// main.BuildEpoch, so main forwards them once at startup.
var (
	buildVersion  = "dev"
	buildCommitID = "unknown"
	buildEpoch    = "0"
)

// SetBuildInfo records the linker-supplied build metadata. It is called once
// from main before anything else reads it.
func SetBuildInfo(version, commitID, epoch string) {
	if version != "" {
		buildVersion = version
	}
	if commitID != "" {
		buildCommitID = commitID
	}
	if epoch != "" {
		buildEpoch = epoch
	}
}

// Version returns the build version stamped into the binary
func Version() string { return buildVersion }

// CommitID returns the git commit the binary was built from
func CommitID() string { return buildCommitID }

// BuildEpoch returns the build timestamp as a Unix epoch string
func BuildEpoch() string { return buildEpoch }

// Config represents the complete configuration
type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	EncryptionKey string
	// WebhookSecret verifies inbound provider webhooks; empty means reject
	WebhookSecret string
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host            string
	Port            int
	Mode            string
	Domain          string
	ApplicationName string
	Tagline         string
	// AdminPath is the single path segment the admin panel is mounted under,
	// below /server. It is operator-configurable and defaults to
	// "administration"; it is never a hardcoded literal in a route.
	AdminPath string
	// APIVersion is the version segment of every /api/{api_version}/ route
	APIVersion string
	// HealthzRootEnabled mounts /healthz to the same handler as the canonical
	// /server/healthz. It is an alias, never a redirect, and is off by default.
	HealthzRootEnabled bool
	TLSEnabled         bool
	TLSCertPath        string
	TLSKeyPath         string
	// TLSMinVersion is the lowest TLS version the listener negotiates, named
	// as in server.yml ("TLS1.2", "TLS1.3"). Resolve it with TLSVersion.
	TLSMinVersion string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	Footer        FooterConfig
	// CORS origins: a single origin, a comma-separated list, or "*"
	CORS string
	// Security carries the RFC 9116 security.txt fields
	Security SecurityContactConfig
}

// SecurityContactConfig holds the security.txt reporting channels
type SecurityContactConfig struct {
	// ReportURL is the primary contact: private vulnerability reporting
	ReportURL string
	// Contact is the secondary CC email address, without a mailto: prefix
	Contact string
	// Expires is the RFC 3339 expiry stamp published in security.txt
	Expires string
}

// FooterConfig holds footer customization settings
type FooterConfig struct {
	// TrackingID is the analytics tracking ID; empty disables tracking
	TrackingID    string
	CookieConsent CookieConsentConfig
	// CustomHTML is custom branding HTML rendered above the application footer
	CustomHTML string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	// Type is the database driver: sqlite, postgres, or mysql
	Type string
	// DSN is the driver-specific connection string
	DSN         string
	MaxConns    int
	MaxIdleConn int
	Replicas    []string
}

// CookieConsentConfig holds cookie consent popup settings
type CookieConsentConfig struct {
	Enabled   bool
	Message   string
	PolicyURL string
}

// fileConfig mirrors server.yml, the persisted source of truth for every
// setting that is not a documented runtime environment variable.
type fileConfig struct {
	Server fileServer `yaml:"server"`
	Web    fileWeb    `yaml:"web"`
}

// fileServer is the `server:` block of server.yml.
type fileServer struct {
	Port       int          `yaml:"port"`
	FQDN       string       `yaml:"fqdn"`
	Address    string       `yaml:"address"`
	Mode       string       `yaml:"mode"`
	AdminPath  string       `yaml:"admin_path"`
	APIVersion string       `yaml:"api_version"`
	Healthz    fileHealthz  `yaml:"healthz"`
	Branding   fileBranding `yaml:"branding"`
	SSL        fileSSL      `yaml:"ssl"`
	Security   fileSecurity `yaml:"security"`
	Footer     fileFooter   `yaml:"footer"`
	Database   fileDatabase `yaml:"database"`
}

// fileWeb is the `web:` block of server.yml.
type fileWeb struct {
	// CORS is a single origin, a comma-separated list, or "*"
	CORS     string          `yaml:"cors"`
	Security fileWebSecurity `yaml:"security"`
}

// fileWebSecurity is the `web.security:` block of server.yml, which supplies
// the RFC 9116 security.txt fields.
type fileWebSecurity struct {
	// ReportURL is the primary contact: private vulnerability reporting
	ReportURL string `yaml:"report_url"`
	// Contact is the secondary CC address; the mailto: prefix is added here
	Contact string `yaml:"contact"`
	// Expires is an RFC 3339 timestamp, auto-renewed a year out when stale
	Expires string `yaml:"expires"`
}

// fileHealthz is the `server.healthz:` block of server.yml.
type fileHealthz struct {
	Root fileHealthzRoot `yaml:"root"`
}

// fileHealthzRoot is the `server.healthz.root:` block of server.yml.
type fileHealthzRoot struct {
	Enabled bool `yaml:"enabled"`
}

// fileBranding is the `server.branding:` block of server.yml.
type fileBranding struct {
	Title   string `yaml:"title"`
	Tagline string `yaml:"tagline"`
}

// fileSSL is the `server.ssl:` block of server.yml.
type fileSSL struct {
	Enabled bool `yaml:"enabled"`
	// Cert and Key are optional manual overrides for cert auto-detection
	Cert       string `yaml:"cert"`
	Key        string `yaml:"key"`
	MinVersion string `yaml:"min_version"`
}

// fileSecurity is the `server.security:` block of server.yml. The encryption
// key lives here rather than in a separate file or an environment variable.
type fileSecurity struct {
	EncryptionKey string `yaml:"encryption_key"`
	// EncryptionKeyVersion starts at 1 and increments on every rotation
	EncryptionKeyVersion int `yaml:"encryption_key_version"`
	// WebhookSecret is the shared secret inbound webhooks are verified against
	WebhookSecret string `yaml:"webhook_secret"`
}

// fileFooter is the `server.footer:` block of server.yml.
type fileFooter struct {
	TrackingID      string `yaml:"tracking_id"`
	CustomHTML      string `yaml:"custom_html"`
	CookieConsent   bool   `yaml:"cookie_consent"`
	CookieMessage   string `yaml:"cookie_message"`
	CookiePolicyURL string `yaml:"cookie_policy_url"`
}

// fileDatabase is the `server.database:` block of server.yml.
type fileDatabase struct {
	Driver      string   `yaml:"driver"`
	URL         string   `yaml:"url"`
	MaxConns    int      `yaml:"max_conns"`
	MaxIdle     int      `yaml:"max_idle"`
	ReplicaURLs []string `yaml:"replica_urls"`
}

// Default network and timeout values applied when nothing overrides them.
const (
	// defaultListenAddress binds every IPv4 and IPv6 interface
	defaultListenAddress = "[::]"
	defaultMode          = "production"
	defaultAdminPath     = "administration"
	defaultAPIVersion    = "v1"
	defaultTLSMinVersion = "TLS1.2"
	defaultCORS          = "*"
	defaultDatabase      = "sqlite"
	defaultReadTimeout   = 30 * time.Second
	defaultWriteTimeout  = 30 * time.Second
	defaultIdleTimeout   = 120 * time.Second
	defaultMaxConns      = 25
	defaultMaxIdleConns  = 5
	defaultCookieMessage = "This site uses cookies for functionality and analytics."
	defaultCookiePolicy  = "/server/privacy"
	randomPortBase       = 64000
	randomPortRange      = 1000
	encryptionKeyBytes   = 32
)

// Load resolves configuration from server.yml, the runtime environment
// variables, and built-in defaults, persisting anything generated on first run.
func Load() (*Config, error) {
	// Ensure required directories exist before any key material is written
	if err := ensureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	stored, err := loadFileConfig()
	if err != nil {
		return nil, err
	}

	// Init-only variables seed server.yml on first run and are ignored after
	// the corresponding value has been persisted
	changed, err := applyInitOnlyEnv(&stored)
	if err != nil {
		return nil, err
	}

	if stored.Server.Security.EncryptionKey == "" {
		key, err := generateEncryptionKey()
		if err != nil {
			return nil, err
		}
		stored.Server.Security.EncryptionKey = key
		stored.Server.Security.EncryptionKeyVersion = 1
		changed = true
	}

	if err := validateEncryptionKey(stored.Server.Security.EncryptionKey); err != nil {
		return nil, fmt.Errorf("%s: server.security.encryption_key: %w", ConfigFile(), err)
	}

	if changed {
		if err := saveFileConfig(stored); err != nil {
			return nil, err
		}
	}

	serverCfg, err := buildServerConfig(stored)
	if err != nil {
		return nil, err
	}

	return &Config{
		Server:        serverCfg,
		Database:      buildDatabaseConfig(stored.Server.Database),
		EncryptionKey: stored.Server.Security.EncryptionKey,
		WebhookSecret: stored.Server.Security.WebhookSecret,
	}, nil
}

// applyInitOnlyEnv fills empty server.yml fields from the PART 5 init-only
// environment variables, or from generated defaults on first run. It reports
// whether server.yml needs to be rewritten.
func applyInitOnlyEnv(stored *fileConfig) (bool, error) {
	changed := false
	srv := &stored.Server

	if srv.Port == 0 {
		port, err := initialPort()
		if err != nil {
			return false, err
		}
		srv.Port = port
		changed = true
	}

	defaults := []struct {
		field *string
		env   string
		value string
	}{
		{&srv.Address, "LISTEN", defaultListenAddress},
		{&srv.Branding.Title, "APPLICATION_NAME", InternalName},
		{&srv.Branding.Tagline, "APPLICATION_TAGLINE", ""},
		{&srv.AdminPath, "", defaultAdminPath},
		{&srv.APIVersion, "", defaultAPIVersion},
		{&srv.SSL.MinVersion, "", defaultTLSMinVersion},
		{&stored.Web.CORS, "", defaultCORS},
		{&srv.Footer.CookieMessage, "", defaultCookieMessage},
		{&srv.Footer.CookiePolicyURL, "", defaultCookiePolicy},
		{&srv.SSL.Cert, "", filepath.Join(SSLDir(), "local", "server.crt")},
		{&srv.SSL.Key, "", filepath.Join(SSLDir(), "local", "server.key")},
		{&srv.Database.Driver, "", defaultDatabase},
	}

	for _, d := range defaults {
		if *d.field != "" {
			continue
		}
		value := d.value
		if d.env != "" {
			if fromEnv := os.Getenv(d.env); fromEnv != "" {
				value = fromEnv
			}
		}
		if value == "" {
			continue
		}
		*d.field = value
		changed = true
	}

	if srv.Mode == "" {
		srv.Mode = defaultMode
		changed = true
	}

	if stored.Web.Security.ReportURL == "" {
		stored.Web.Security.ReportURL = fmt.Sprintf(
			"https://github.com/%s/%s/security/advisories/new", ProjectOrg, ProjectName)
		changed = true
	}

	// security.txt is invalid once Expires passes, so a stale stamp is renewed
	// rather than published as-is
	if expiry, err := time.Parse(time.RFC3339, stored.Web.Security.Expires); err != nil || expiry.Before(time.Now()) {
		stored.Web.Security.Expires = time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339)
		changed = true
	}

	if srv.Database.MaxConns == 0 {
		srv.Database.MaxConns = defaultMaxConns
		changed = true
	}

	if srv.Database.MaxIdle == 0 {
		srv.Database.MaxIdle = defaultMaxIdleConns
		changed = true
	}

	return changed, nil
}

// initialPort returns the port for a first run: PORT when set, otherwise a
// random unused port in the 64000-64999 range.
func initialPort() (int, error) {
	if val := os.Getenv("PORT"); val != "" {
		port, err := strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("PORT: invalid port %q: %w", val, err)
		}
		if port < 1 || port > 65535 {
			return 0, fmt.Errorf("PORT: %d out of range 1-65535", port)
		}
		return port, nil
	}

	return randomPort()
}

// buildServerConfig merges the persisted server block with the runtime
// environment variables that are re-read on every start.
func buildServerConfig(cfg fileConfig) (ServerConfig, error) {
	srv := cfg.Server

	mode := srv.Mode
	if fromEnv := os.Getenv("MODE"); fromEnv != "" {
		mode = fromEnv
	}

	domain := srv.FQDN
	if fromEnv := os.Getenv("DOMAIN"); fromEnv != "" {
		domain = fromEnv
	}

	adminPath, err := SafeAdminPath(srv.AdminPath)
	if err != nil {
		return ServerConfig{}, fmt.Errorf("%s: server.admin_path: %w", ConfigFile(), err)
	}

	return ServerConfig{
		Host:               srv.Address,
		Port:               srv.Port,
		Mode:               mode,
		Domain:             domain,
		ApplicationName:    srv.Branding.Title,
		Tagline:            srv.Branding.Tagline,
		AdminPath:          adminPath,
		APIVersion:         srv.APIVersion,
		HealthzRootEnabled: srv.Healthz.Root.Enabled,
		TLSEnabled:         srv.SSL.Enabled,
		TLSCertPath:        srv.SSL.Cert,
		TLSKeyPath:         srv.SSL.Key,
		TLSMinVersion:      srv.SSL.MinVersion,
		ReadTimeout:        defaultReadTimeout,
		WriteTimeout:       defaultWriteTimeout,
		IdleTimeout:        defaultIdleTimeout,
		Footer: FooterConfig{
			TrackingID: srv.Footer.TrackingID,
			CookieConsent: CookieConsentConfig{
				Enabled:   srv.Footer.CookieConsent,
				Message:   srv.Footer.CookieMessage,
				PolicyURL: srv.Footer.CookiePolicyURL,
			},
			CustomHTML: srv.Footer.CustomHTML,
		},
		CORS: cfg.Web.CORS,
		Security: SecurityContactConfig{
			ReportURL: cfg.Web.Security.ReportURL,
			Contact:   cfg.Web.Security.Contact,
			Expires:   cfg.Web.Security.Expires,
		},
	}, nil
}

// TLSVersion resolves a server.yml min_version name to its crypto/tls
// constant. Unknown or empty values fall back to TLS 1.2, so a typo weakens
// nothing — it just fails to raise the floor.
func TLSVersion(name string) uint16 {
	switch strings.ToUpper(strings.ReplaceAll(name, " ", "")) {
	case "TLS1.3", "TLSV1.3", "1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}

// adminPathReserved lists the segments already mounted directly under /server.
// An operator who set admin_path to one of these would shadow a built-in route,
// so the value is rejected at startup rather than silently taking precedence.
var adminPathReserved = map[string]bool{
	"about":       true,
	"api":         true,
	"auth":        true,
	"contact":     true,
	"docs":        true,
	"healthz":     true,
	"help":        true,
	"preferences": true,
	"privacy":     true,
	"static":      true,
	".well-known": true,
}

// SafeAdminPath normalizes and validates the configured admin panel segment.
// It rejects traversal, uppercase, separators, and any reserved segment.
func SafeAdminPath(input string) (string, error) {
	path := strings.Trim(strings.TrimSpace(input), "/")
	if path == "" {
		path = defaultAdminPath
	}

	if strings.ContainsAny(path, "/\\") {
		return "", fmt.Errorf("%q must be a single path segment", path)
	}
	if path != strings.ToLower(path) {
		return "", fmt.Errorf("%q must be lowercase", path)
	}
	for _, r := range path {
		isAllowed := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !isAllowed {
			return "", fmt.Errorf("%q contains an invalid character %q", path, r)
		}
	}
	if adminPathReserved[path] {
		return "", fmt.Errorf("%q is reserved and would shadow a built-in route", path)
	}

	return path, nil
}

// loadFileConfig reads server.yml, returning a zero value when it is absent.
func loadFileConfig() (fileConfig, error) {
	var cfg fileConfig

	data, err := os.ReadFile(ConfigFile())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("failed to read %s: %w", ConfigFile(), err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse %s: %w", ConfigFile(), err)
	}

	return cfg, nil
}

// saveFileConfig writes server.yml back to disk. It holds the encryption key,
// so it is never group- or world-readable.
func saveFileConfig(cfg fileConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to encode %s: %w", ConfigFile(), err)
	}

	if err := os.WriteFile(ConfigFile(), data, 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", ConfigFile(), err)
	}

	return nil
}

// buildDatabaseConfig merges the persisted database block with DATABASE_DRIVER
// and DATABASE_URL, which PART 5 lists as runtime variables.
func buildDatabaseConfig(db fileDatabase) DatabaseConfig {
	driver := db.Driver
	if fromEnv := os.Getenv("DATABASE_DRIVER"); fromEnv != "" {
		driver = fromEnv
	}

	dsn := db.URL
	if fromEnv := os.Getenv("DATABASE_URL"); fromEnv != "" {
		dsn = fromEnv
	}

	// Default DSN for SQLite
	if dsn == "" && driver == defaultDatabase {
		dsn = filepath.Join(DatabaseDir(), "server.db")
	}

	return DatabaseConfig{
		Type:        driver,
		DSN:         dsn,
		MaxConns:    db.MaxConns,
		MaxIdleConn: db.MaxIdle,
		Replicas:    db.ReplicaURLs,
	}
}

// validateEncryptionKey rejects key material that is not a base64-encoded
// 256-bit value. Startup fails loudly rather than letting a truncated or
// hand-edited key surface much later as an opaque AES error, and the process
// never silently rotates a key that would orphan existing ciphertext.
func validateEncryptionKey(key string) error {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(key))
	if err != nil {
		return fmt.Errorf("encryption key is not valid base64: %w", err)
	}
	if len(raw) != encryptionKeyBytes {
		return fmt.Errorf("encryption key must decode to %d bytes, got %d", encryptionKeyBytes, len(raw))
	}
	return nil
}

// generateEncryptionKey generates a new random 256-bit encryption key
func generateEncryptionKey() (string, error) {
	key := make([]byte, encryptionKeyBytes)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to read from the system CSPRNG: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// randomPort generates a random port in the range 64000-64999.
// A CSPRNG failure is reported rather than silently falling back so the chosen
// port is never predictable.
func randomPort() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(randomPortRange))
	if err != nil {
		return 0, fmt.Errorf("failed to read from the system CSPRNG: %w", err)
	}
	return randomPortBase + int(n.Int64()), nil
}

// ensureDirectories creates required directories if they don't exist
func ensureDirectories() error {
	dirs := []string{
		ConfigDir(),
		SSLDir(),
		filepath.Join(SSLDir(), "letsencrypt"),
		filepath.Join(SSLDir(), "local"),
		DataDir(),
		DatabaseDir(),
		CacheDir(),
		LogDir(),
		SecurityDir(),
		filepath.Join(SecurityDir(), "geoip"),
		filepath.Join(SecurityDir(), "blocklists"),
		filepath.Join(SecurityDir(), "cve"),
		filepath.Join(SecurityDir(), "trivy"),
		filepath.Join(DataDir(), "scanners"),
		filepath.Join(DataDir(), "artifacts"),
		filepath.Join(DataDir(), "workspaces"),
		filepath.Join(LogDir(), "builds"),
		filepath.Join(LogDir(), "users"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	return nil
}
