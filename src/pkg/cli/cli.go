package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Description is the one-line summary shown in --help.
const Description = "CI/CD application server"

// Config holds every value the server binary accepts on the command line.
type Config struct {
	Mode           string
	ConfigDir      string
	DataDir        string
	CacheDir       string
	LogDir         string
	BackupDir      string
	PIDFile        string
	Address        string
	Port           int
	BaseURL        string
	Daemon         bool
	Debug          bool
	Color          string
	Lang           string
	ShowHelp       bool
	ShowVersion    bool
	ShowStatus     bool
	Service        string
	Maintenance    string
	MaintenanceArg string
	Update         string
	UpdateBranch   string
	Shell          string
	ShellArg       string
}

// BinaryName returns the name the binary was invoked as, so renamed binaries
// display their actual name in help, version, and error output.
func BinaryName() string {
	return filepath.Base(os.Args[0])
}

// Parse parses command line arguments using the stdlib flag package.
func Parse(version string) (*Config, error) {
	cfg := &Config{}

	fs := flag.NewFlagSet(BinaryName(), flag.ExitOnError)

	fs.BoolVar(&cfg.ShowHelp, "help", false, "Show help")
	fs.BoolVar(&cfg.ShowHelp, "h", false, "Show help")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "Show version")
	fs.BoolVar(&cfg.ShowVersion, "v", false, "Show version")
	fs.BoolVar(&cfg.ShowStatus, "status", false, "Show server status and health")

	fs.StringVar(&cfg.Mode, "mode", "", "Application mode (production|development|debug)")
	fs.StringVar(&cfg.ConfigDir, "config", "", "Config directory")
	fs.StringVar(&cfg.DataDir, "data", "", "Data directory")
	fs.StringVar(&cfg.CacheDir, "cache", "", "Cache directory")
	fs.StringVar(&cfg.LogDir, "log", "", "Log directory")
	fs.StringVar(&cfg.BackupDir, "backup", "", "Backup directory")
	fs.StringVar(&cfg.PIDFile, "pid", "", "PID file path")
	fs.StringVar(&cfg.Address, "address", "", "Listen address")
	fs.IntVar(&cfg.Port, "port", 0, "Listen port")
	fs.StringVar(&cfg.BaseURL, "baseurl", "", "URL path prefix")
	fs.BoolVar(&cfg.Daemon, "daemon", false, "Run as daemon (detach from terminal)")
	fs.BoolVar(&cfg.Debug, "debug", false, "Enable debug mode")
	fs.StringVar(&cfg.Color, "color", "auto", "Color output (auto|yes|no)")
	fs.StringVar(&cfg.Lang, "lang", "", "Language for output")

	fs.StringVar(&cfg.Service, "service", "", "Service management command")
	fs.StringVar(&cfg.Maintenance, "maintenance", "", "Maintenance operation")
	fs.StringVar(&cfg.Update, "update", "", "Check or perform updates")
	fs.StringVar(&cfg.Shell, "shell", "", "Shell integration (completions|init|help)")

	fs.Usage = func() { ShowHelp(version) }

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	// The subcommand-style flags take one optional positional argument
	if fs.NArg() > 0 {
		arg := fs.Arg(0)
		switch {
		case cfg.Shell != "":
			cfg.ShellArg = arg
		case cfg.Update == "branch":
			cfg.UpdateBranch = arg
		case cfg.Maintenance != "":
			cfg.MaintenanceArg = arg
		}
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate rejects out-of-range values before they reach the server.
func (c *Config) validate() error {
	switch c.Mode {
	case "", "production", "development", "debug":
	default:
		return fmt.Errorf("invalid --mode %q: expected production, development, or debug", c.Mode)
	}

	switch strings.ToLower(c.Color) {
	case "", "auto", "yes", "no":
	default:
		return fmt.Errorf("invalid --color %q: expected auto, yes, or no", c.Color)
	}

	if c.Port != 0 && (c.Port < 1 || c.Port > 65535) {
		return fmt.Errorf("invalid --port %d: expected 1-65535", c.Port)
	}

	if c.BaseURL != "" && !strings.HasPrefix(c.BaseURL, "/") {
		return fmt.Errorf("invalid --baseurl %q: must start with /", c.BaseURL)
	}

	return nil
}

// ShowHelp displays the full flag reference for the server binary.
func ShowHelp(version string) {
	name := BinaryName()

	fmt.Printf("%s %s - %s\n\n", name, version, Description)
	fmt.Printf("Usage:\n  %s [flags]\n\n", name)
	fmt.Print(`Information:
-h, --help                             - Show help
-v, --version                          - Show version
--status                               - Show server status and health

Shell Integration:
--shell completions [SHELL]            - Print shell completions
--shell init [SHELL]                   - Print shell init command
--shell help                           - Show shell help

Server Configuration:
--mode {production|development|debug}  - Application mode (default: production)
--config DIR                           - Config directory
--data DIR                             - Data directory
--cache DIR                            - Cache directory
--log DIR                              - Log directory
--backup DIR                           - Backup directory
--pid FILE                             - PID file path
--address ADDR                         - Listen address (default: 0.0.0.0)
--port PORT                            - Listen port (default: random 64xxx, 80 in container)
--baseurl PATH                         - URL path prefix (default: /)
--daemon                               - Run as daemon (detach from terminal)
--debug                                - Enable debug mode
--color {auto|yes|no}                  - Color output (default: auto)
--lang CODE                            - Language for output (default: auto)

Service Management:
--service CMD                          - Service management (run --service help for details)
--maintenance CMD                      - Maintenance operations (run --maintenance help for details)
--update [CMD]                         - Check/perform updates (run --update help for details)

`)
	fmt.Printf("Run '%s <command> help' for detailed help on any command.\n", name)
}

// ShowVersion displays the build metadata embedded at link time.
func ShowVersion(version, buildDate, commit string) {
	fmt.Printf("%s %s (%s)\nBuilt: %s\nGo: %s\nOS/Arch: %s/%s\n",
		BinaryName(), version, commit, buildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// ShouldRunServer reports whether the parsed flags ask for the HTTP server
// rather than a one-shot informational or management command.
func (c *Config) ShouldRunServer() bool {
	return !c.ShowHelp &&
		!c.ShowVersion &&
		!c.ShowStatus &&
		c.Shell == "" &&
		c.Service == "" &&
		c.Maintenance == "" &&
		c.Update == ""
}

// ApplyEnvironment exports directory overrides so the config package resolves
// the same paths the user asked for on the command line.
func (c *Config) ApplyEnvironment() error {
	overrides := map[string]string{
		"CONFIG_DIR": c.ConfigDir,
		"DATA_DIR":   c.DataDir,
		"CACHE_DIR":  c.CacheDir,
		"LOG_DIR":    c.LogDir,
		"BACKUP_DIR": c.BackupDir,
		"MODE":       c.Mode,
		"LISTEN":     c.Address,
	}

	if c.Port != 0 {
		overrides["PORT"] = fmt.Sprintf("%d", c.Port)
	}

	for key, val := range overrides {
		if val == "" {
			continue
		}
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}
