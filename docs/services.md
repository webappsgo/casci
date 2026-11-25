# Service installers

The repository includes installer scripts and service definitions, but the
binary's `--service` command is not implemented. The scripts and units have
not been validated as a complete lifecycle and currently contain paths and
environment variable names that do not all match the application's config
resolver. Review them before installation; do not use them on a host with
existing CASCI data without first resolving those differences.

## Available files

| Platform | Files |
|----------|-------|
| Linux/systemd | `scripts/install-systemd.sh`, `scripts/casci.service` |
| macOS/launchd | `scripts/install-launchd.sh`, `scripts/com.casapps.casci.plist` |
| Windows | `scripts/install-windows.ps1` |
| Linux/macOS removal | `scripts/uninstall-service.sh` |

The installers require administrative privileges and modify system accounts,
service managers, and install paths. Read the script and service definition
before running it.

## Known integration gaps

- The systemd and launchd unit files define `CASCI_DB_TYPE` and
  `CASCI_DB_DSN`; the current application reads `DATABASE_DRIVER` and
  `DATABASE_URL` instead.
- Installer paths do not consistently match the application's resolved
  `casapps/casci` config/data/log paths.
- The Windows installer registers a service but its environment values use
  the same unsupported `CASCI_DB_*` names.
- The binary does not currently implement `--service install`, start, stop,
  restart, or reload operations.

See [Installation](installation.md) for the high-level deployment options and
[Configuration](configuration.md) for environment variables the application
actually reads.
