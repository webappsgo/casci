# Service Installation Guide

## Overview

CASCI provides native service installers for Linux (systemd), macOS (launchd), and Windows.

## Linux (systemd)

### Installation

```bash
sudo ./scripts/install-systemd.sh
```

This will:
- Create `casci:casci` user and group
- Install binary to `/usr/local/bin/casci`
- Create data directory at `/var/lib/casci`
- Create log directory at `/var/log/casci`
- Install systemd service file
- Enable service for automatic startup

### Service Management

```bash
# Start service
sudo systemctl start casci

# Stop service
sudo systemctl stop casci

# Restart service
sudo systemctl restart casci

# Check status
sudo systemctl status casci

# View logs
sudo journalctl -u casci -f

# Enable auto-start
sudo systemctl enable casci

# Disable auto-start
sudo systemctl disable casci
```

### Configuration

Edit `/etc/systemd/system/casci.service` and add environment variables:

```ini
[Service]
Environment="CASCI_DB_TYPE=postgres"
Environment="CASCI_DB_DSN=postgres://user:pass@localhost/casci"
Environment="CASCI_PORT=8080"
```

Then reload:
```bash
sudo systemctl daemon-reload
sudo systemctl restart casci
```

## macOS (launchd)

### Installation

```bash
sudo ./scripts/install-launchd.sh
```

This will:
- Create `_casci:_casci` user and group with system UID/GID
- Install binary to `/usr/local/bin/casci`
- Create data directory at `/usr/local/var/casci`
- Create log directory at `/usr/local/var/log/casci`
- Install launchd plist file
- Load service for automatic startup

### Service Management

```bash
# Start service
sudo launchctl start com.casapps.casci

# Stop service
sudo launchctl stop com.casapps.casci

# Check status
sudo launchctl list | grep casci

# View logs
tail -f /usr/local/var/log/casci/stdout.log
tail -f /usr/local/var/log/casci/stderr.log

# Unload service
sudo launchctl unload /Library/LaunchDaemons/com.casapps.casci.plist

# Load service
sudo launchctl load /Library/LaunchDaemons/com.casapps.casci.plist
```

### Configuration

Edit `/Library/LaunchDaemons/com.casapps.casci.plist` and modify EnvironmentVariables:

```xml
<key>EnvironmentVariables</key>
<dict>
    <key>CASCI_DB_TYPE</key>
    <string>postgres</string>
    <key>CASCI_DB_DSN</key>
    <string>postgres://user:pass@localhost/casci</string>
</dict>
```

Then reload:
```bash
sudo launchctl unload /Library/LaunchDaemons/com.casapps.casci.plist
sudo launchctl load /Library/LaunchDaemons/com.casapps.casci.plist
```

## Windows

### Installation

Run PowerShell as Administrator:

```powershell
.\scripts\install-windows.ps1
```

This will:
- Create Virtual Service Account `NT SERVICE\CASCI` (automatic)
- Install binary to `C:\Program Files\CASCI\casci.exe`
- Create data directory at `C:\ProgramData\CASCI`
- Create log directory at `C:\ProgramData\CASCI\logs`
- Create Windows service
- Configure automatic startup
- Set recovery options (restart on failure)

### Service Management

```powershell
# Start service
Start-Service -Name CASCI

# Stop service
Stop-Service -Name CASCI

# Restart service
Restart-Service -Name CASCI

# Check status
Get-Service -Name CASCI

# View logs
Get-EventLog -LogName Application -Source CASCI -Newest 50

# Change startup type
Set-Service -Name CASCI -StartupType Automatic  # or Manual, Disabled
```

### Configuration

Set environment variables using registry:

```powershell
$serviceName = "CASCI"
$envPath = "HKLM:\SYSTEM\CurrentControlSet\Services\$serviceName"

# Set environment variables
Set-ItemProperty -Path $envPath -Name "Environment" -Value @(
    "CASCI_DB_TYPE=postgres",
    "CASCI_DB_DSN=postgres://user:pass@localhost/casci"
)

# Restart service
Restart-Service -Name CASCI
```

## Uninstallation

### Linux / macOS

```bash
sudo ./scripts/uninstall-service.sh
```

The script will:
- Stop and disable the service
- Remove service configuration files
- Prompt to remove binary and data directories

### Windows

Create `uninstall-windows.ps1`:

```powershell
Stop-Service -Name CASCI -Force
sc.exe delete CASCI
Remove-Item -Recurse -Force "$env:ProgramFiles\CASCI"
# Optionally remove data:
# Remove-Item -Recurse -Force "$env:ProgramData\CASCI"
```

## Troubleshooting

### Linux

**Service won't start:**
```bash
# Check journal for errors
sudo journalctl -u casci -n 50 --no-pager

# Check service status
sudo systemctl status casci

# Verify binary exists and is executable
ls -l /usr/local/bin/casci

# Check permissions
ls -ld /var/lib/casci
```

### macOS

**Service won't load:**
```bash
# Check plist syntax
plutil /Library/LaunchDaemons/com.casapps.casci.plist

# Check logs
tail -50 /usr/local/var/log/casci/stderr.log

# Verify user exists
dscl . -read /Users/_casci
```

### Windows

**Service won't start:**
```powershell
# Check event log
Get-EventLog -LogName Application -Source CASCI -Newest 20

# Verify service configuration
Get-Service -Name CASCI | Format-List *

# Check binary exists
Test-Path "$env:ProgramFiles\CASCI\casci.exe"
```

## Security Notes

### Linux systemd
- Service runs as non-privileged `casci` user
- Security hardening enabled (NoNewPrivileges, PrivateTmp, ProtectSystem, etc.)
- Minimal filesystem access (only /var/lib/casci and /var/log/casci writable)

### macOS launchd
- Service runs as system user `_casci` (hidden from login window)
- UID/GID allocated in system range (100-499)
- Limited to data directory access only

### Windows
- Uses Virtual Service Account (no password, managed by Windows)
- Automatic identity management by Windows
- Least-privilege access to required directories only
