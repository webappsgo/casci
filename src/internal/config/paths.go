package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// Organization and application identifiers used to build every OS path.
// These are the frozen internal_org/internal_name values from IDEA.md.
const (
	InternalOrg  = "casapps"
	InternalName = "casci"
)

// Public repository coordinates, used to build the security.txt primary
// contact URL. These are the project_org/project_name values from IDEA.md.
const (
	ProjectOrg  = "webappsgo"
	ProjectName = "casci"
)

// isPrivileged reports whether the process runs with administrative rights.
// Windows has no euid, so it is always treated as unprivileged and uses the
// per-user location unless an explicit environment override is supplied.
func isPrivileged() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	return os.Geteuid() == 0
}

// homeDir returns the current user's home directory, falling back to the
// working directory when the home cannot be determined.
func homeDir() string {
	if dir, err := os.UserHomeDir(); err == nil && dir != "" {
		return dir
	}
	return "."
}

// appPath joins the organization and application name onto a base directory.
func appPath(base string) string {
	return filepath.Join(base, InternalOrg, InternalName)
}

// ConfigDir returns the directory holding server.yml.
// CONFIG_DIR overrides the OS default.
func ConfigDir() string {
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return dir
	}
	home := homeDir()
	switch runtime.GOOS {
	case "darwin":
		if isPrivileged() {
			return appPath("/Library/Application Support")
		}
		return appPath(filepath.Join(home, "Library", "Application Support"))
	case "windows":
		if data := os.Getenv("ProgramData"); data != "" && isPrivileged() {
			return appPath(data)
		}
		return appPath(os.Getenv("APPDATA"))
	case "freebsd", "openbsd", "netbsd":
		if isPrivileged() {
			return appPath("/usr/local/etc")
		}
		return appPath(filepath.Join(home, ".config"))
	default:
		if isPrivileged() {
			return appPath("/etc")
		}
		return appPath(filepath.Join(home, ".config"))
	}
}

// DataDir returns the directory holding persistent application state.
// DATA_DIR overrides the OS default.
func DataDir() string {
	if dir := os.Getenv("DATA_DIR"); dir != "" {
		return dir
	}
	home := homeDir()
	switch runtime.GOOS {
	case "darwin":
		if isPrivileged() {
			return filepath.Join(appPath("/Library/Application Support"), "data")
		}
		return appPath(filepath.Join(home, "Library", "Application Support"))
	case "windows":
		return ConfigDir()
	case "freebsd", "openbsd", "netbsd":
		if isPrivileged() {
			return appPath("/var/db")
		}
		return appPath(filepath.Join(home, ".local", "share"))
	default:
		if isPrivileged() {
			return appPath("/var/lib")
		}
		return appPath(filepath.Join(home, ".local", "share"))
	}
}

// CacheDir returns the directory holding regenerable cached artifacts.
// CACHE_DIR overrides the OS default.
func CacheDir() string {
	if dir := os.Getenv("CACHE_DIR"); dir != "" {
		return dir
	}
	home := homeDir()
	switch runtime.GOOS {
	case "darwin":
		if isPrivileged() {
			return appPath("/Library/Caches")
		}
		return appPath(filepath.Join(home, "Library", "Caches"))
	case "windows":
		return filepath.Join(DataDir(), "cache")
	default:
		if isPrivileged() {
			return appPath("/var/cache")
		}
		return appPath(filepath.Join(home, ".cache"))
	}
}

// LogDir returns the directory holding server.log and its subdirectories.
// LOG_DIR overrides the OS default.
func LogDir() string {
	if dir := os.Getenv("LOG_DIR"); dir != "" {
		return dir
	}
	home := homeDir()
	switch runtime.GOOS {
	case "darwin":
		if isPrivileged() {
			return appPath("/Library/Logs")
		}
		return appPath(filepath.Join(home, "Library", "Logs"))
	case "windows":
		return filepath.Join(DataDir(), "logs")
	default:
		if isPrivileged() {
			return appPath("/var/log")
		}
		return appPath(filepath.Join(home, ".local", "log"))
	}
}

// DatabaseDir returns the directory holding the SQLite databases.
// DATABASE_DIR overrides the OS default.
func DatabaseDir() string {
	if dir := os.Getenv("DATABASE_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(DataDir(), "db")
}

// BackupDir returns the directory holding generated backups.
// BACKUP_DIR overrides the OS default.
func BackupDir() string {
	if dir := os.Getenv("BACKUP_DIR"); dir != "" {
		return dir
	}
	home := homeDir()
	switch runtime.GOOS {
	case "darwin":
		if isPrivileged() {
			return appPath("/Library/Backups")
		}
		return appPath(filepath.Join(home, "Library", "Backups"))
	case "windows":
		return filepath.Join(DataDir(), "backups")
	case "freebsd", "openbsd", "netbsd":
		if isPrivileged() {
			return appPath("/var/backups")
		}
		return appPath(filepath.Join(home, ".local", "share", "Backups"))
	default:
		if isPrivileged() {
			return appPath("/mnt/Backups")
		}
		return appPath(filepath.Join(home, ".local", "share", "Backups"))
	}
}

// SSLDir returns the directory holding TLS material.
func SSLDir() string {
	return filepath.Join(ConfigDir(), "ssl")
}

// SecurityDir returns the directory holding security feeds and scanner data.
func SecurityDir() string {
	return filepath.Join(DataDir(), "security")
}

// ConfigFile returns the absolute path of server.yml.
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "server.yml")
}

// LogFile returns the absolute path of server.log.
func LogFile() string {
	return filepath.Join(LogDir(), "server.log")
}

// PIDFile returns the absolute path of the daemon PID file.
func PIDFile() string {
	if isPrivileged() && runtime.GOOS != "windows" {
		return filepath.Join("/var/run", InternalOrg, InternalName+".pid")
	}
	return filepath.Join(DataDir(), InternalName+".pid")
}
