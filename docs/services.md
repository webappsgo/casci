# Service management

The `casci` server binary installs and controls native service managers. Use
`casci --service --help` to inspect the detected manager and current state.

## Commands

| Command | Behavior |
|---------|----------|
| `casci --service --install` | Install, enable, and start the service |
| `casci --service start` | Start an installed service |
| `casci --service stop` | Stop the service |
| `casci --service restart` | Restart the service |
| `casci --service reload` | Ask the service manager to reload or reconfigure |
| `casci --service --disable` | Stop and disable automatic start while keeping application data |
| `casci --service --uninstall` | Remove the service and application data after confirmation; keeps the binary |

Run user-level systemd or launchd operations without elevation when that manager
is available. System-level installation and management may prompt for elevated
privileges. Linux supports systemd, OpenRC, runit, s6, and SysVinit; macOS uses
launchd, FreeBSD uses rc.d, and Windows uses the Service Control Manager.

## Data and paths

System installation prepares the service account and application directories.
The server binds its listener before dropping privileges to that account.
Windows service paths are under `%PROGRAMDATA%/{org}/{internal_name}`. On Unix,
paths follow the configured project path resolver; inspect `casci --help` and
the [configuration reference](configuration.md) for overrides.

Uninstall is destructive: it deletes the service's config, data, cache, logs,
backups, and service account when the application created that account. Back up
data first if it must be retained. `--disable` is the reversible option that
keeps data.

## Installer scripts

The repository also contains scripts and unit examples under `scripts/` for
platforms that need operator-managed installation. The binary command is the
supported lifecycle interface; inspect any script before running it because
scripts execute host-level changes outside the binary's service lifecycle.
