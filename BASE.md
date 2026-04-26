# {PROJECTNAME} Specification

> **TEMPLATE USAGE**: Copy this file to your project as `SPEC.md` and replace all `{projectname}` and `{PROJECTNAME}` placeholders with your actual project name. Fill in the project-specific sections marked with `{...}`.

**Name**: {projectname}

## Working Roles

When working on this project, the following roles are assumed based on the task:

- **Senior Go Developer** - Writing production-quality Go code, making architectural decisions, following best practices, optimizing performance
- **UI/UX Designer** - Creating professional, functional, visually appealing interfaces with excellent user experience
- **Beta Tester** - Testing applications, finding bugs, edge cases, and issues before they reach users
- **User** - Thinking from the end-user perspective, ensuring things are intuitive and work as expected

These are not roleplay - they ARE these roles when the work requires it. Each project gets the full expertise of all four perspectives.

## Core Rules (Non-Negotiable and Non-Replaceable)

**THESE RULES CANNOT BE CHANGED, OVERRIDDEN, OR IGNORED.**

### Specification Compliance
- **Re-read this spec periodically** during work to ensure accuracy and no deviation
- When in doubt, check the spec
- The spec is the source of truth for all project decisions

### Required Documentation Files
These files MUST be kept in sync and read as needed during work:

| File | Purpose | When to Read |
|------|---------|--------------|
| **AI.md** | Project-specific notes + BASE.md rules merged in | Read as needed, keep in sync with project state |
| **TODO.AI.md** | Task tracking (when >2 tasks) | Read before starting work, update as tasks complete |

- **AI.md MUST contain BASE.md rules** - copy/merge BASE.md content into each project's AI.md
- **AI.md MUST always reflect current project state** - update after significant changes
- **TODO.AI.md MUST be used when doing more than 2 tasks** - keeps work organized and trackable
- **Migration**: If `CLAUDE.md` or `SPEC.md` exist, merge their content into `AI.md` and delete the old files

### Target Audience
- Self-hosted
- SMB (Small/Medium Business)
- Enterprise
- **Assume self-hosted and SMB users are not that tech savvy**

### Development Principles
- We validate everything
- We sanitize where appropriate
- Save only what is valid
- Only clear what is invalid
- **Never expose sensitive information** unless necessary:
  - Tokens and passwords shown only once on generation (must be copied)
  - Show on first run, password changes, token regeneration
  - Show in difficult environments (Docker, headless servers)
  - Never log sensitive data
  - Never in error messages or stack traces
  - Mask in UI (show `••••••••` or last 4 chars only)
- We test everything where applicable
- We show tool tips or documentation where needed
- We are security and mobile first (where applicable)
- We always set sane defaults for everything
- **Security should never get in the way of usability**
- No AI or ML - everything is very smart logic
- Responses are short, concise, yet descriptive and helpful

### Questions and Help
- Question mark (?) means asking you a question
- You can and should offer help where applicable

---

## Project Information

| Field | Value |
|-------|-------|
| **Name** | {projectname} |
| **Organization** | {projectorg} |
| **Official Site** | https://{projectname}.{projectorg}.us |
| **Repository** | https://github.com/{projectorg}/{projectname} |
| **README** | README.md |
| **License** | MIT > LICENSE.md |
| **Embedded Licenses** | Added to bottom of LICENSE.md |

## Project Description

{Brief description of what this project does}

## Project-Specific Features

{List features unique to this project}

---

## Project Structure

### Variables
- `{projectname}`: The project name (e.g., "jokes")
- `{projectorg}`: Organization name = "{projectorg}"
- **If anything is wrapped in `{}` it is a variable**
- **Anything NOT wrapped in `{}` is NOT a variable**
- Example: `/etc/letsencrypt/live/domain` is a literal directory, not a template/variable

### Directory Structure

**The root Project directory is**: `./`

```
./                          # Root project directory
├── src/                    # All source files
├── scripts/                # All production/install scripts
├── tests/                  # All development/test scripts and files
├── binaries/               # Built binaries (gitignored)
├── releases/               # Release binaries (gitignored)
├── README.md               # Production first, dev last
├── SPEC.md                 # This specification file
├── LICENSE.md              # MIT + embedded licenses
├── AI.md                   # AI/Claude working notes
├── TODO.AI.md              # Task tracking for >2 tasks
└── release.txt             # Version tracking
```

**Keep the base directory organized and clean - no clutter!**

**The working directory is `.`**

---

## Platform Support

### Operating Systems
- Linux
- BSD (FreeBSD, OpenBSD, etc.)
- macOS (Intel and Apple Silicon)
- Windows

### Architectures
- AMD64
- ARM64

**Because we are supporting AMD64 and ARM64 and all OSes, be smart about implementations**

---

## Go Version

### Always Use Latest Stable Go
- **Go is only used for building, not runtime** (single static binary)
- Always use the latest stable Go version for builds
- Use latest stable version in `go.mod` files (e.g., `go 1.23` or newer)
- Docker builds should use `golang:latest` for build/test/debug
- Do NOT pin to specific minor versions unless there's a compatibility issue
- Since we build static binaries, we can always use the latest Go version

---

# Directory Structures by OS

## Linux

### Privileged (root/sudo)

| Type | Path |
|------|------|
| Binary | `/usr/local/bin/{projectname}` |
| Config | `/etc/{projectorg}/{projectname}/` |
| Config File | `/etc/{projectorg}/{projectname}/server.yml` |
| Data | `/var/lib/{projectorg}/{projectname}/` |
| Logs | `/var/log/{projectorg}/{projectname}/` |
| Backup | `/mnt/Backups/{projectorg}/{projectname}/` |
| PID File | `/var/run/{projectorg}/{projectname}.pid` |
| SSL Certs | `/etc/{projectorg}/{projectname}/ssl/certs/` |
| SQLite DB | `/var/lib/{projectorg}/{projectname}/db/` |
| GeoIP | `/var/lib/{projectorg}/{projectname}/geoip/` |
| Service | `/etc/systemd/system/{projectname}.service` |

### User (non-privileged)

| Type | Path |
|------|------|
| Binary | `~/.local/bin/{projectname}` |
| Config | `~/.config/{projectorg}/{projectname}/` |
| Config File | `~/.config/{projectorg}/{projectname}/server.yml` |
| Data | `~/.local/share/{projectorg}/{projectname}/` |
| Logs | `~/.local/share/{projectorg}/{projectname}/logs/` |
| Backup | `~/.local/backups/{projectorg}/{projectname}/` |
| PID File | `~/.local/share/{projectorg}/{projectname}/{projectname}.pid` |
| SSL Certs | `~/.config/{projectorg}/{projectname}/ssl/certs/` |
| SQLite DB | `~/.local/share/{projectorg}/{projectname}/db/` |
| GeoIP | `~/.local/share/{projectorg}/{projectname}/geoip/` |

---

## macOS

### Privileged (root/sudo)

| Type | Path |
|------|------|
| Binary | `/usr/local/bin/{projectname}` |
| Config | `/Library/Application Support/{projectorg}/{projectname}/` |
| Config File | `/Library/Application Support/{projectorg}/{projectname}/server.yml` |
| Data | `/Library/Application Support/{projectorg}/{projectname}/data/` |
| Logs | `/Library/Logs/{projectorg}/{projectname}/` |
| Backup | `/Library/Backups/{projectorg}/{projectname}/` |
| PID File | `/var/run/{projectorg}/{projectname}.pid` |
| SSL Certs | `/Library/Application Support/{projectorg}/{projectname}/ssl/certs/` |
| SQLite DB | `/Library/Application Support/{projectorg}/{projectname}/db/` |
| GeoIP | `/Library/Application Support/{projectorg}/{projectname}/geoip/` |
| Service | `/Library/LaunchDaemons/com.{projectorg}.{projectname}.plist` |

### User (non-privileged)

| Type | Path |
|------|------|
| Binary | `~/bin/{projectname}` or `/usr/local/bin/{projectname}` |
| Config | `~/Library/Application Support/{projectorg}/{projectname}/` |
| Config File | `~/Library/Application Support/{projectorg}/{projectname}/server.yml` |
| Data | `~/Library/Application Support/{projectorg}/{projectname}/` |
| Logs | `~/Library/Logs/{projectorg}/{projectname}/` |
| Backup | `~/Library/Backups/{projectorg}/{projectname}/` |
| PID File | `~/Library/Application Support/{projectorg}/{projectname}/{projectname}.pid` |
| SSL Certs | `~/Library/Application Support/{projectorg}/{projectname}/ssl/certs/` |
| SQLite DB | `~/Library/Application Support/{projectorg}/{projectname}/db/` |
| GeoIP | `~/Library/Application Support/{projectorg}/{projectname}/geoip/` |
| Service | `~/Library/LaunchAgents/com.{projectorg}.{projectname}.plist` |

---

## BSD (FreeBSD, OpenBSD, NetBSD)

### Privileged (root/sudo/doas)

| Type | Path |
|------|------|
| Binary | `/usr/local/bin/{projectname}` |
| Config | `/usr/local/etc/{projectorg}/{projectname}/` |
| Config File | `/usr/local/etc/{projectorg}/{projectname}/server.yml` |
| Data | `/var/db/{projectorg}/{projectname}/` |
| Logs | `/var/log/{projectorg}/{projectname}/` |
| Backup | `/var/backups/{projectorg}/{projectname}/` |
| PID File | `/var/run/{projectorg}/{projectname}.pid` |
| SSL Certs | `/usr/local/etc/{projectorg}/{projectname}/ssl/certs/` |
| SQLite DB | `/var/db/{projectorg}/{projectname}/db/` |
| GeoIP | `/var/db/{projectorg}/{projectname}/geoip/` |
| Service | `/usr/local/etc/rc.d/{projectname}` |

### User (non-privileged)

| Type | Path |
|------|------|
| Binary | `~/.local/bin/{projectname}` |
| Config | `~/.config/{projectorg}/{projectname}/` |
| Config File | `~/.config/{projectorg}/{projectname}/server.yml` |
| Data | `~/.local/share/{projectorg}/{projectname}/` |
| Logs | `~/.local/share/{projectorg}/{projectname}/logs/` |
| Backup | `~/.local/backups/{projectorg}/{projectname}/` |
| PID File | `~/.local/share/{projectorg}/{projectname}/{projectname}.pid` |
| SSL Certs | `~/.config/{projectorg}/{projectname}/ssl/certs/` |
| SQLite DB | `~/.local/share/{projectorg}/{projectname}/db/` |
| GeoIP | `~/.local/share/{projectorg}/{projectname}/geoip/` |

---

## Windows

### Privileged (Administrator)

| Type | Path |
|------|------|
| Binary | `C:\Program Files\{projectorg}\{projectname}\{projectname}.exe` |
| Config | `%ProgramData%\{projectorg}\{projectname}\` |
| Config File | `%ProgramData%\{projectorg}\{projectname}\server.yml` |
| Data | `%ProgramData%\{projectorg}\{projectname}\data\` |
| Logs | `%ProgramData%\{projectorg}\{projectname}\logs\` |
| Backup | `%ProgramData%\Backups\{projectorg}\{projectname}\` |
| SSL Certs | `%ProgramData%\{projectorg}\{projectname}\ssl\certs\` |
| SQLite DB | `%ProgramData%\{projectorg}\{projectname}\db\` |
| GeoIP | `%ProgramData%\{projectorg}\{projectname}\geoip\` |
| Service | Windows Service Manager |

### User (non-privileged)

| Type | Path |
|------|------|
| Binary | `%LocalAppData%\{projectorg}\{projectname}\{projectname}.exe` |
| Config | `%AppData%\{projectorg}\{projectname}\` |
| Config File | `%AppData%\{projectorg}\{projectname}\server.yml` |
| Data | `%LocalAppData%\{projectorg}\{projectname}\` |
| Logs | `%LocalAppData%\{projectorg}\{projectname}\logs\` |
| Backup | `%LocalAppData%\Backups\{projectorg}\{projectname}\` |
| SSL Certs | `%AppData%\{projectorg}\{projectname}\ssl\certs\` |
| SQLite DB | `%LocalAppData%\{projectorg}\{projectname}\db\` |
| GeoIP | `%LocalAppData%\{projectorg}\{projectname}\geoip\` |

---

## Docker/Container

| Type | Path |
|------|------|
| Binary | `/usr/local/bin/{projectname}` |
| Config | `/config/` |
| Config File | `/config/server.yml` |
| Data | `/data/` |
| Logs | `/data/logs/` |
| SQLite DB | `/data/db/` |
| GeoIP | `/data/geoip/` |
| Internal Port | `80` |

---

# Privilege Escalation & User Creation

## Overview

Application user creation **REQUIRES** privilege escalation. If the user cannot escalate privileges, the application runs as the current user with user-level directories.

## Escalation Detection by OS

### Linux
```
Escalation Methods (in order of preference):
1. Already root (EUID == 0)
2. sudo (if user is in sudoers/wheel group)
3. su (if user knows root password)
4. pkexec (PolicyKit, if available)
5. doas (OpenBSD-style, if configured)

Detection:
- Check EUID: os.Geteuid() == 0
- Check sudo: exec.LookPath("sudo") && user in sudo/wheel group
- Check su: exec.LookPath("su")
- Check pkexec: exec.LookPath("pkexec")
- Check doas: exec.LookPath("doas") && /etc/doas.conf exists
```

### macOS
```
Escalation Methods (in order of preference):
1. Already root (EUID == 0)
2. sudo (user must be in admin group)
3. osascript with administrator privileges (GUI prompt)

Detection:
- Check EUID: os.Geteuid() == 0
- Check sudo: exec.LookPath("sudo") && user in admin group
- GUI available: os.Getenv("DISPLAY") != "" or always try osascript
```

### BSD (FreeBSD, OpenBSD, NetBSD)
```
Escalation Methods (in order of preference):
1. Already root (EUID == 0)
2. doas (OpenBSD default, others if configured)
3. sudo (if installed and configured)
4. su (if user knows root password)

Detection:
- Check EUID: os.Geteuid() == 0
- Check doas: exec.LookPath("doas") && /etc/doas.conf exists
- Check sudo: exec.LookPath("sudo")
- Check su: exec.LookPath("su")
```

### Windows
```
Escalation Methods (in order of preference):
1. Already Administrator (elevated token)
2. UAC prompt (requires GUI)
3. runas (command line, requires admin password)

Detection:
- Check Admin: windows.GetCurrentProcessToken().IsElevated()
- UAC available: GUI session detected
- runas: always available but requires password
```

## User Creation Logic

```
ON --service --install:

1. Check if can escalate privileges
   ├─ YES: Continue with privileged installation
   │   ├─ Create system user/group (UID/GID 100-999)
   │   ├─ Use system directories (/etc, /var/lib, /var/log)
   │   ├─ Install service (systemd/launchd/rc.d/Windows Service)
   │   └─ Set ownership to created user
   │
   └─ NO: Fall back to user installation
       ├─ Skip user creation (run as current user)
       ├─ Use user directories (~/.config, ~/.local/share)
       ├─ Skip system service installation
       └─ Offer alternative (cron, user systemd, launchctl user agent)
```

## System User Requirements

When creating a system user (privileged only):

| Requirement | Value |
|-------------|-------|
| Username | `{projectname}` |
| Group | `{projectname}` |
| UID/GID | Auto-detect unused in range 100-999 |
| Shell | `/sbin/nologin` or `/usr/sbin/nologin` |
| Home | Config or data directory |
| Type | System user (no password, no login) |
| Gecos | `{projectname} service account` |

### User Creation Commands by OS

**Linux:**
```bash
# Find unused UID/GID
for id in $(seq 100 999); do
  if ! getent passwd $id && ! getent group $id; then
    echo $id; break
  fi
done

# Create group and user
groupadd -r -g {UID} {projectname}
useradd -r -u {UID} -g {projectname} -s /sbin/nologin \
  -d /var/lib/{projectorg}/{projectname} -c "{projectname} service" {projectname}
```

**macOS:**
```bash
# Find unused UID/GID (use dscl)
dscl . -list /Users UniqueID | awk '{print $2}' | sort -n
# Pick unused ID in 100-999

# Create group and user
dscl . -create /Groups/{projectname}
dscl . -create /Groups/{projectname} PrimaryGroupID {GID}
dscl . -create /Users/{projectname}
dscl . -create /Users/{projectname} UniqueID {UID}
dscl . -create /Users/{projectname} PrimaryGroupID {GID}
dscl . -create /Users/{projectname} UserShell /usr/bin/false
dscl . -create /Users/{projectname} NFSHomeDirectory /Library/Application\ Support/{projectorg}/{projectname}
```

**BSD:**
```bash
# FreeBSD
pw groupadd {projectname} -g {GID}
pw useradd {projectname} -u {UID} -g {projectname} -s /sbin/nologin \
  -d /var/db/{projectorg}/{projectname} -c "{projectname} service"

# OpenBSD
groupadd -g {GID} {projectname}
useradd -u {UID} -g {projectname} -s /sbin/nologin \
  -d /var/db/{projectorg}/{projectname} -c "{projectname} service" {projectname}
```

**Windows:**
```powershell
# Windows doesn't typically create service users
# Services run as LocalSystem, LocalService, NetworkService, or a domain account
# For isolation, can create local user (requires admin):

net user {projectname} /add /active:no
# Or use a managed service account (domain environments)
```

## Privilege Check Flow

```
START
  │
  ├─ Check: Am I running as root/admin?
  │   ├─ YES → Use privileged paths, can create user
  │   └─ NO → Continue to escalation check
  │
  ├─ Check: Can I escalate privileges?
  │   │
  │   ├─ Linux:
  │   │   ├─ Can sudo? (sudo -n true 2>/dev/null)
  │   │   ├─ Can doas? (doas -n true 2>/dev/null)
  │   │   ├─ Can pkexec? (pkexec --help 2>/dev/null)
  │   │   └─ Has su access? (harder to detect without password)
  │   │
  │   ├─ macOS:
  │   │   ├─ Can sudo? (sudo -n true 2>/dev/null)
  │   │   └─ In admin group? (groups | grep -q admin)
  │   │
  │   ├─ BSD:
  │   │   ├─ Can doas? (doas -n true 2>/dev/null)
  │   │   ├─ Can sudo? (sudo -n true 2>/dev/null)
  │   │   └─ Has su access?
  │   │
  │   └─ Windows:
  │       └─ Can elevate? (check UAC settings, admin group membership)
  │
  ├─ CAN ESCALATE:
  │   ├─ Prompt: "Installation requires administrator privileges. Continue? [Y/n]"
  │   ├─ If Yes: Re-execute with escalation
  │   │   ├─ Linux: sudo/doas/pkexec {binary} --service --install
  │   │   ├─ macOS: sudo {binary} --service --install
  │   │   ├─ BSD: doas/sudo {binary} --service --install
  │   │   └─ Windows: Trigger UAC elevation
  │   └─ If No: Fall back to user installation
  │
  └─ CANNOT ESCALATE:
      ├─ Warn: "Cannot obtain administrator privileges."
      ├─ Warn: "Installing for current user only."
      ├─ Use user-level directories
      ├─ Skip system user creation
      └─ Offer user-level service alternatives:
          ├─ Linux: systemctl --user, cron @reboot
          ├─ macOS: launchctl user agent
          ├─ BSD: cron @reboot
          └─ Windows: Task Scheduler (current user)
```

## Installation Output Examples

### Privileged Installation (Success)
```
🔐 Administrator privileges detected

📦 Installing {projectname}...

Creating system user:
  ✓ Group '{projectname}' created (GID: 847)
  ✓ User '{projectname}' created (UID: 847)

Creating directories:
  ✓ /etc/{projectorg}/{projectname}
  ✓ /var/lib/{projectorg}/{projectname}
  ✓ /var/log/{projectorg}/{projectname}

Installing binary:
  ✓ /usr/local/bin/{projectname}

Installing service:
  ✓ /etc/systemd/system/{projectname}.service
  ✓ Service enabled

📋 Configuration file created:
   /etc/{projectorg}/{projectname}/server.yml

🔑 Admin credentials (SAVE THESE - shown only once):
   Username: administrator
   Password: xK9#mP2$vL5@nQ8
   API Token: {projectorg}_7f8a9b2c3d4e5f6a7b8c9d0e1f2a3b4c

✅ Installation complete!

To start the service:
  sudo systemctl start {projectname}

To check status:
  sudo systemctl status {projectname}
```

### User Installation (No Privileges)
```
⚠️  Cannot obtain administrator privileges
📦 Installing {projectname} for current user...

Creating directories:
  ✓ ~/.config/{projectorg}/{projectname}
  ✓ ~/.local/share/{projectorg}/{projectname}
  ✓ ~/.local/share/{projectorg}/{projectname}/logs

Installing binary:
  ✓ ~/.local/bin/{projectname}

📋 Configuration file created:
   ~/.config/{projectorg}/{projectname}/server.yml

🔑 Admin credentials (SAVE THESE - shown only once):
   Username: administrator
   Password: xK9#mP2$vL5@nQ8
   API Token: {projectorg}_7f8a9b2c3d4e5f6a7b8c9d0e1f2a3b4c

⚠️  System service not installed (requires administrator)

Alternative options:
  • Run manually: ~/.local/bin/{projectname}
  • Add to crontab: @reboot ~/.local/bin/{projectname}
  • User systemd: systemctl --user enable {projectname}

✅ Installation complete!
```

---

# Built-in Service Support (Non-Negotiable)

**All projects MUST have built-in service support for ALL service managers.**

### Service Management
- Built-in service support for all service managers:
  - systemd (Linux)
  - runit (Linux)
  - Windows Service Manager
  - macOS launchd
  - BSD rc.d
  - Other service managers as applicable

---

# Configuration

## Configuration Source of Truth

**Single Instance (file driver):**
- Config file is source of truth
- Support live reload where possible

**With Database (sqlite, mariadb, mysql, postgres, mssql):**
- **Database is source of truth**
- Config file kept in sync (db → config, one-way sync)
- /admin panel writes to database
- Changes propagate to all instances

## Config/Database Initialization Flow
1. First instance starts, no schema exists
2. Create schema, populate from config file
3. Database is now source of truth
4. Other instances connect, inherit settings from database
5. Config file updated to match database (backup of current state)

## Conflict Resolution
- **Optimistic locking** with version/timestamp field
- Each config record has `version` (integer) and `updated_at` (timestamp)
- On save: check version matches, increment, save
- If version mismatch: **last write wins** with warning logged
- Conflicts logged to audit log with before/after values
- /admin shows "config changed by another instance" warning if stale

## Sync Behavior
```
Database ──────► Config File (one-way sync on change)
    ▲
    │
/admin panel writes here
```
- Config file is a **readable backup**, not the source
- On startup: read database, update config file if different
- Never read config file after initial population (except manual override flag)

## Boolean Handling (Non-Negotiable)
For ease of use, accept these values for booleans:
- **Truthy**: `1`, `yes`, `true`, `enable`, `enabled`, `on`
- **Falsy**: `0`, `no`, `false`, `disable`, `disabled`, `off`

Internally convert all to `true` or `false`.

## Environment Variables (Non-Negotiable)

**Runtime Environment Variables (always respected):**
- `MODE` - Application mode: `production` (default) or `development`
  - Unlike other env vars, MODE is checked on EVERY startup
  - Can be overridden via `--mode` CLI flag
  - See "Application Modes" section for behavior differences
- `DATABASE_DRIVER` - Database driver: `file`, `sqlite`, `mariadb`, `mysql`, `postgres`, `mssql`, `mongodb`
- `DATABASE_URL` - Database connection string (overrides individual connection settings)
  - SQLite: `file:/path/to/database.db` or just path
  - MariaDB/MySQL: `user:pass@tcp(host:port)/dbname`
  - PostgreSQL: `postgres://user:pass@host:port/dbname?sslmode=disable`
  - MSSQL: `sqlserver://user:pass@host:port?database=dbname`
  - MongoDB: `mongodb://user:pass@host:port/dbname`

**Init-Only Environment Variables:**
The following can be defined for **initialization only**. Once initialized, the config file is the source of truth:
- `CONFIG_DIR` - Configuration directory
- `DATA_DIR` - Data directory
- `LOG_DIR` - Log directory
- `BACKUP_DIR` - Backup directory
- `DATABASE_DIR` - SQLite database directory (default: `{DATA_DIR}/db`)
- `PORT` - Server port
- `LISTEN` - Listen address
- `APPLICATION_NAME` - Application title (server.title)
- `APPLICATION_TAGLINE` - Application description/tagline (server.description)

**These init-only variables are used once during first run to initialize the config file, then ignored.**

**Note:** `MODE`, `DATABASE_DRIVER`, and `DATABASE_URL` are runtime variables - they are checked on every startup and can override config file settings. No CLI flags for database settings.

---

# Application Modes (Non-Negotiable)

**All projects MUST support production and development modes.**

## Mode Detection (Priority Order)
1. `--mode` CLI flag (highest priority)
2. `MODE` environment variable
3. Default: `production`

## Production Mode (Default)
Production mode is optimized for security, performance, and stability:

| Setting | Behavior |
|---------|----------|
| **Logging** | `info` level, minimal output |
| **Debug endpoints** | Disabled (`/debug/*` returns 404) |
| **Error messages** | Generic (no stack traces, no internal details) |
| **Panic recovery** | Graceful (logs error, returns 500, continues serving) |
| **Template caching** | Enabled (templates parsed once at startup) |
| **Static file caching** | Enabled (appropriate cache headers) |
| **CORS** | Configured value only (no wildcards unless explicit) |
| **Rate limiting** | Enforced per configuration |
| **Security headers** | All enabled |
| **Sensitive data** | Never shown (masked in logs, UI, responses) |
| **Auto-reload** | Disabled |
| **Profiling** | Disabled |

## Development Mode
Development mode is optimized for debugging and rapid iteration:

| Setting | Behavior |
|---------|----------|
| **Logging** | `debug` level, verbose output |
| **Debug endpoints** | Enabled (`/debug/pprof/*`, `/debug/vars`) |
| **Error messages** | Detailed (stack traces, internal error details) |
| **Panic recovery** | Verbose (full stack trace in response) |
| **Template caching** | Disabled (templates re-parsed on each request) |
| **Static file caching** | Disabled (no-cache headers) |
| **CORS** | Permissive (`*` allowed for local development) |
| **Rate limiting** | Relaxed or disabled |
| **Security headers** | Relaxed for local testing |
| **Sensitive data** | Can be shown with warning banner |
| **Auto-reload** | Config file changes trigger reload |
| **Profiling** | Available at `/debug/pprof/*` |

## Mode-Specific Console Output

**Production startup:**
```
🚀 {projectname} v{version}
   Mode: production
   Listening on: https://example.com:443
```

**Development startup:**
```
🔧 {projectname} v{version} [DEVELOPMENT MODE]
   ⚠️  Debug endpoints enabled
   ⚠️  Verbose error messages enabled
   ⚠️  Template caching disabled
   Mode: development
   Listening on: http://localhost:64xxx
   Debug: http://localhost:64xxx/debug/pprof/
```

## Implementation Requirements
- Mode MUST be stored in config struct and accessible globally
- Mode check MUST happen before any request processing
- Mode MUST be displayed in `/healthz` and `/api/v1/healthz` responses
- Mode MUST be shown in admin dashboard
- Mode changes via env var require restart (no hot-switch between modes)
- `--mode dev` and `--mode development` both accepted for development mode
- `--mode prod` and `--mode production` both accepted for production mode

---

# Configuration File

## Configuration File Design (Non-Negotiable)
- Must be clean, intuitive, and very easy to read
- **If it has a setting, it MUST be configurable via the configuration file**
- **We have sane defaults built-in** because no one wants to manage a 1000 line config file
- Comprehensive with all options (but commented/defaulted appropriately)
- Single-line comments (under 140 characters)

## Configuration Locations
- **Root users**: `/etc/{projectorg}/{projectname}/server.yml`
- **Regular users**: `~/.config/{projectorg}/{projectname}/server.yml`
- **Migration**: If `server.yaml` found, auto-migrate to `server.yml` on startup
- Auto-create config file with comprehensive defaults on first run

## Example Configuration Structure

```yaml
# =============================================================================
# SERVER CONFIGURATION
# =============================================================================

server:
  # Port: single (HTTP) or dual (HTTP,HTTPS) e.g., "8090" or "8090,64453"
  # Default: random unused port in 64xxx range, saved to config after first run
  port: {random}

  # Fully qualified domain name for this server
  # Default: auto-detected from host (hostname -f or equivalent)
  fqdn: {hostname}

  # Listen address:
  # [::] = all interfaces IPv4/IPv6 (default)
  # 0.0.0.0 = all interfaces IPv4 only
  # 127.0.0.1 = localhost only
  address: "[::]"

  # Application mode: production or development
  # Can be overridden by MODE env var or --mode CLI flag
  mode: production

  # Application branding
  title: ""
  description: ""

  # System user/group - {auto} creates on first run
  user: {auto}
  group: {auto}

  # PID file for process management
  pidfile: true

  # ---------------------------------------------------------------------------
  # Admin Panel Configuration
  # ---------------------------------------------------------------------------
  admin:
    email: admin@{fqdn}
    username: administrator
    password: {auto}
    token: {auto}

  # ---------------------------------------------------------------------------
  # SSL/TLS Configuration
  # ---------------------------------------------------------------------------
  ssl:
    enabled: false
    cert_path: /etc/{projectorg}/{projectname}/ssl/certs

    letsencrypt:
      enabled: false
      email: admin@{fqdn}
      challenge: http-01
      dns_provider_type: ""
      dns_provider_key: ""

  # ---------------------------------------------------------------------------
  # Scheduler
  # ---------------------------------------------------------------------------
  schedule:
    enabled: true
    cert_renewal: daily
    notifications: hourly
    cleanup: weekly

  # ---------------------------------------------------------------------------
  # GeoIP
  # ---------------------------------------------------------------------------
  geoip:
    enabled: true
    dir: "{datadir}/geoip"
    update: weekly
    deny_countries: []
    databases:
      asn: true
      country: true
      city: true

  # ---------------------------------------------------------------------------
  # Metrics
  # ---------------------------------------------------------------------------
  metrics:
    enabled: false
    endpoint: /metrics
    include_system: true
    token: ""

  # ---------------------------------------------------------------------------
  # Logging
  # ---------------------------------------------------------------------------
  logs:
    level: info
    debug:
      enabled: false
      filename: debug.log
      format: text
    access:
      filename: access.log
      format: apache
    server:
      filename: server.log
      format: text
    audit:
      filename: audit.log
      format: json
    security:
      filename: security.log
      format: fail2ban

  # ---------------------------------------------------------------------------
  # Rate Limiting
  # ---------------------------------------------------------------------------
  rate_limit:
    enabled: true
    requests: 120
    window: 60

  # ---------------------------------------------------------------------------
  # Database
  # ---------------------------------------------------------------------------
  database:
    driver: file

    sqlite:
      dir: "{datadir}/db"
      server_db: server.db
      users_db: users.db
      journal_mode: WAL
      busy_timeout: 5000

    # mariadb, mysql, postgres, mssql, mongodb configs...

# =============================================================================
# FRONTEND CONFIGURATION
# =============================================================================

web:
  ui:
    theme: dark
    logo: ""
    favicon: ""

  cors: "*"

  footer:
    tracking_id: ""
    cookie_consent:
      enabled: true
      message: "In accordance with the EU GDPR law this message is being displayed."
    custom_html: ""
```

---

# Port Configuration

## Port Format
- **Single port** (HTTP only): `8080`
- **Dual port** (HTTP + HTTPS): `8080,64453` (second port is always HTTPS)

## Default Behavior
- Prefer to be behind a reverse proxy
- Default to random unused port in 64xxx range using the user system
- Save port to configuration file for persistence

## Special Case: Ports 80,443
If PORT is `80,443`:
- Get certificate from Let's Encrypt
- All certs saved to `/etc/{projectorg}/{projectname}/ssl/certs`
- **Check `/etc/letsencrypt/live` first** - if cert found, use it but don't manage it

---

# SSL/TLS & Let's Encrypt

## Built-in Let's Encrypt Support (Non-Negotiable)
**All projects MUST have built-in Let's Encrypt support.**

Supported challenge types:
- **DNS-01** (all providers and RFC2136)
- **TLS-ALPN-01**
- **HTTP-01**

## Certificate Management
- Check `/etc/letsencrypt/live` first (literal path, not a variable)
- Save to `/etc/{projectorg}/{projectname}/ssl/certs`
- Auto-renewal via built-in scheduler

---

# Built-in Scheduler (Non-Negotiable)

**All projects MUST have a built-in scheduler.**

## Purpose
- Certificate renewals
- Notification checks
- Other periodic tasks
- Configurable via configuration file

---

# Web Frontend

## Requirements
- **ALL PROJECTS MUST HAVE A FANTASTIC FRONTEND BUILT IN**
- Full mobile support
- HTML5 with full web standards compliance
- Full accessibility
- Must be readable, navigable, intuitive, user friendly, accessibility enabled, self explanatory

## Technology Stack
- Use templates where/when possible (header, nav, body, footer, etc.)
- Prefer vanilla JS and CSS
- No frameworks unless absolutely necessary
- **NEVER use default JavaScript alerts/confirms/prompts**
- Always use custom CSS modals, toast notifications, and professional UI elements
- **NEVER use inline CSS styles** - always create reusable CSS classes
  - Bad: `<div style="color: red; margin: 10px;">`
  - Good: `<div class="error-text spacing-sm">`
  - All styles must be in CSS files, not in HTML elements

## Layout
- **Screens ≥ 720px**: 90% width (left 5%, right 5%)
- **Screens < 720px**: 98% width (left 1%, right 1%)
- **Footer**: Always centered and always at bottom of screen (scroll to see)

## Themes
- **Dark** (based on Dracula) - **DEFAULT**
- **Light** (based on popular light theme)
- **auto** (Based on the users system)

---

# API Structure

## Versioning
- **Use versioned API**: `/api/v1`

## API Types
- **REST API** (primary)
- **Swagger** documentation
- **GraphQL** support
- **ALL PROJECTS GET ALL THREE**

## Root-Level Endpoints (Non-Negotiable)

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/` | GET | None | Web interface (HTML) |
| `/healthz` | GET | None | Health check (HTML) |
| `/openapi` | GET | None | Swagger UI |
| `/openapi.json` | GET | None | OpenAPI spec (JSON) |
| `/openapi.yaml` | GET | None | OpenAPI spec (YAML) |
| `/graphql` | GET | None | GraphiQL interface |
| `/graphql` | POST | None | GraphQL queries |
| `/metrics` | GET | Optional | Prometheus metrics |
| `/admin` | GET | Session | Admin panel login |
| `/admin/*` | ALL | Session | Admin panel pages |
| `/api/v1/healthz` | GET | None | Health check (JSON) |
| `/api/v1/admin/*` | ALL | Bearer | Admin API |

## API Response Standards (Non-Negotiable)

**Response Formats:**
- All `/` routes return HTML
- All `/api` routes return JSON (default) or text based on Accept header
- All `/api/**/*.txt` routes return text

**Error Response Format:**
```json
{
  "error": "Human readable message",
  "code": "ERROR_CODE",
  "status": 400,
  "details": {}
}
```

**Pagination (default: 250 items):**
```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "limit": 250,
    "total": 1000,
    "pages": 4
  }
}
```

---

# Admin Panel (Non-Negotiable)

**ALL projects MUST have a full admin panel for server configuration.**

## Design Principles
- **Pretty** - Clean, modern, professional design
- **Intuitive** - Self-explanatory, no manual needed
- **Easy to navigate** - Logical grouping, breadcrumbs, search
- **Follows all frontend rules** - Dracula theme (default), responsive, accessible
- **No default JS alerts** - Custom modals, toasts, confirmations
- **Real-time feedback** - Show save status, validation errors inline
- **Mobile-friendly** - Works on all screen sizes

## /admin (Web Interface)

**Authentication:**
- Login form with username/password
- Session cookie (30 days default, configurable)
- CSRF protection on all forms
- "Remember me" option
- Logout button always visible

**Sections:**
1. Overview/Dashboard
2. Server Settings
3. Web Settings
4. Security Settings
5. Database & Cache
6. Email & Notifications
7. SSL/TLS
8. Logs
9. Backup & Maintenance
10. System Info

## /api/v1/admin (REST API)

**Authentication:**
- Header: `Authorization: Bearer {token}`
- Token from `server.admin.token`

**Endpoints:**
```
GET    /api/v1/admin/config              # Get full config
PUT    /api/v1/admin/config              # Update full config
PATCH  /api/v1/admin/config              # Partial update
GET    /api/v1/admin/status              # Server status
GET    /api/v1/admin/health              # Detailed health
GET    /api/v1/admin/stats               # Statistics
GET    /api/v1/admin/logs/access         # Access logs
GET    /api/v1/admin/logs/error          # Error logs
POST   /api/v1/admin/backup              # Create backup
POST   /api/v1/admin/restore             # Restore backup
POST   /api/v1/admin/test/email          # Send test email
POST   /api/v1/admin/password            # Change password
POST   /api/v1/admin/token/regenerate    # Regenerate API token
```

---

# CLI Interface (Non-Negotiable)

**THESE COMMANDS CANNOT BE CHANGED. This is the complete command set.**

## Main Commands
```bash
--help                       # Show help (can be run by anyone)
--version                    # Show version (can be run by anyone)
--mode {production|development}  # Set application mode
--data {datadir}             # Set data dir
--config {etcdir}            # Set the config dir
--address {listen}           # Set listen address
--port {port}                # Set the port
--status                     # Show status and health
--service {start,restart,stop,reload,--install,--uninstall,--disable,--help}
--maintenance {backup,restore,update,mode} [optional-file-or-setting]
--update [check|yes|branch {stable|beta|daily}]  # Check/perform updates
```

**Note:** `--help`, `--version`, `--status`, and `--update check` can be run by anyone.

**Mode shortcuts:**
- `--mode dev` or `--mode development` → development mode
- `--mode prod` or `--mode production` → production mode (default)

## Display Rules (Non-Negotiable)
- **Never show `0.0.0.0`, `127.0.0.1`, `localhost`, etc. where applicable**
- User should see valid FQDN, host, or IP
- Show only one, the most relevant

---

# Docker (Non-Negotiable)

## Dockerfile
- **Alpine-based** (latest or version matching build version)
- All meta labels included
- For the scratch image: curl, bash, tini, and binary in `/usr/local/bin`
- **Use tini as init system**

## Container Configuration
- **Internal port**: 80
- **Data dir**: `/data`
- **Config dir**: `/config`
- **Log dir**: `/data/logs/{projectname}`
- **HEALTHCHECK**: `{binary} --status`

## Container Detection
- **Assume running in container if tini init system (PID 1) is detected**

## Tags (Non-Negotiable)
- **Release**: `ghcr.io/{projectorg}/{projectname}:latest`
- **Development**: `{projectname}:dev`

---

# Makefile (Non-Negotiable)

**DO NOT CHANGE THESE TARGETS.**

## Targets
| Target | Description |
|--------|-------------|
| `build` | Build all platforms to `./binaries` |
| `release` | GitHub release - production to `./releases` |
| `docker` | Docker release for ARM64/AMD64 |
| `test` | Run all tests |

## Binary Naming (Non-Negotiable)
- **Local/Testing**: `/tmp/{projectname}`
- **Host Build**: `./binaries/{projectname}`
- **Distribution**: `{projectname}-{os}-{arch}`
- **NEVER include `-musl` suffix** - binaries must be `{projectname}-{os}-{arch}` NOT `{projectname}-{os}-{arch}-musl`
- Example: `jokes-linux-amd64` NOT `jokes-linux-amd64-musl`

---

# GitHub Actions (Non-Negotiable)

**All projects MUST have GitHub Actions workflows for automated builds and releases.**

## Workflow Files

All workflow files in `.github/workflows/`:

| File | Trigger | Purpose |
|------|---------|---------|
| `release.yml` | Tag push (`v*`, `*.*.*`) | Production releases |
| `beta.yml` | Push to `beta` branch | Beta releases |
| `daily.yml` | Daily at 3am UTC + push to `main`/`master` | Daily/dev builds |

## Release Workflow (`release.yml`)

**Trigger:** Tag push with or without `v` prefix
```yaml
on:
  push:
    tags:
      - 'v*'      # v1.0.0, v1.0.0-rc1
      - '[0-9]*'  # 1.0.0, 1.0.0-rc1
```

**Version:** From tag (strip `v` prefix if present)

**Build Matrix:**
| OS | Arch | Binary Name |
|----|------|-------------|
| Linux | amd64 | `{projectname}-linux-amd64` |
| Linux | arm64 | `{projectname}-linux-arm64` |
| macOS | amd64 | `{projectname}-darwin-amd64` |
| macOS | arm64 | `{projectname}-darwin-arm64` |
| Windows | amd64 | `{projectname}-windows-amd64.exe` |
| Windows | arm64 | `{projectname}-windows-arm64.exe` |
| FreeBSD | amd64 | `{projectname}-freebsd-amd64` |
| FreeBSD | arm64 | `{projectname}-freebsd-arm64` |

**Release Process:**
1. Build static binaries (`CGO_ENABLED=0`, no `-musl` suffix)
2. Create source archive (exclude `.git`, `.github`, `binaries/`, `releases/`)
3. Delete existing release/tag if exists (using `gh release delete`)
4. Create new release with all binaries and source archive
5. Update `latest` tag to point to new release

## Beta Workflow (`beta.yml`)

**Trigger:** Push to `beta` branch

**Version Format:** `YYYYMMDDHHMM-beta` (e.g., `202512051430-beta`)

**Release Process:**
1. Build static binaries (`CGO_ENABLED=0`)
2. Create source archive
3. Delete existing beta release if exists
4. Create pre-release with tag `{version}`
5. Mark as pre-release in GitHub

## Daily Workflow (`daily.yml`)

**Trigger:** Daily at 3am UTC + push to `main`/`master`

**Version Format:** `YYYYMMDDHHMM` (e.g., `202512051430`)

**Release Process:**
1. Build static binaries (`CGO_ENABLED=0`)
2. Create source archive
3. Delete existing daily release with same date if exists
4. Create release with tag `{version}`
5. Mark as pre-release in GitHub
6. Keep only last 7 daily releases (cleanup old)

## Update Channel Mapping

The `--maintenance update branch` command maps to these releases:

| Branch | Release Type | Tag Pattern | Example |
|--------|--------------|-------------|---------|
| `stable` | Release | `v*`, `*.*.*` | `v1.0.0`, `1.0.0` |
| `beta` | Pre-release | `*-beta` | `202512051430-beta` |
| `daily` | Pre-release | `YYYYMMDDHHMM` | `202512051430` |

---

# Binary Requirements (Non-Negotiable)

## Single Static Binary
- **THE BINARY MUST BE A SINGLE STATIC BINARY**
- All assets embedded using Go's `embed` package
- No external dependencies at runtime
- **Must build with `CGO_ENABLED=0`**
- Use pure Go dependencies only

## Binary Default Behavior
- **Default (no arguments)**: Initialize (if needed) and start the server
- Auto-creates config file with defaults on first run
- Auto-creates required directories on first run
- **Must have proper signal handling** (SIGTERM, SIGINT, SIGHUP)
- **PID file support** (default: enabled)

## Embedded Assets
- **Templates**: `src/server/templates/`
- **Static files**: `src/server/static/`
- **Application data**: `src/data/` (JSON files)

## External Data Files (NOT Embedded)
- **GeoIP databases** - Download, update via scheduler
- **Blocklists** - Download, update via scheduler
- Any security-related databases

---

# Testing & Development (Non-Negotiable)

## Temporary Directory Structure (Non-Negotiable)
- **Format**: `/tmp/{tmpdir}/{projectname}/` (e.g., `/tmp/apimgr-build/{projectname}/`)
- **All temp files MUST be project-scoped** - never use shared temp directories
- **Cleanup required** - always clean up project temp files after use
- **Examples**:
  - Build output: `/tmp/apimgr-build/{projectname}/`
  - Test config: `/tmp/apimgr-test/{projectname}/`
  - Debug files: `/tmp/apimgr-debug/{projectname}/`
- **NEVER use `/tmp/{projectname}` directly** - always use subdirectory structure

## Container Usage
- **Use Docker/Incus/LXD** for building, testing, and debugging
- **Use `golang:latest`** (NOT `golang:alpine`) for build/test/debug containers
- Test binaries go in temp directories (e.g., `/tmp/apimgr-build/{projectname}/`)

## Build Command Example
```bash
docker run --rm -v /path/to/project:/build -w /build -e CGO_ENABLED=0 golang:latest go build -o /tmp/apimgr-build/{projectname}/{projectname} ./src
```

## Available Host Tools
- **jq** - Available on host for JSON parsing/manipulation

## Running and Testing
- **Always use Docker** for running/testing - never run binaries directly on the host
- Run tests in containers: `docker run --rm ... /tmp/apimgr-build/{projectname}/{projectname} --version`

## Process & Container Management (Non-Negotiable)
**All commands MUST be project-scoped. NEVER run global/broad commands.**

**Forbidden Commands (NEVER use):**
- `pkill -f {pattern}` - too broad, kills unrelated processes
- `docker rm $(docker ps -aq)` - removes ALL containers
- `docker rmi $(docker images -q)` - removes ALL images
- `docker system prune` - cleans ALL unused resources
- `killall {name}` - too broad
- Any command without explicit project scope

**Required: Project-Scoped Commands Only:**
- `docker stop {projectname}` - stop specific project container
- `docker rm {projectname}` - remove specific project container
- `docker rmi {projectorg}/{projectname}:tag` - remove specific project image
- `kill {specific-pid}` - kill exact PID only (verify first)
- `pkill -x {projectname}` - exact binary name match only

**Before Killing/Removing:**
1. List first: `docker ps | grep {projectname}` or `pgrep -la {projectname}`
2. Verify it's the correct process/container
3. Use the most specific command possible
4. Document what was killed and why

---

# Database Migrations (Non-Negotiable)

**ALL apps MUST have built-in AUTOMATIC database migration support.**

## Migration System
- **Fully automatic** - runs on startup
- Versioned migrations with timestamps
- Track applied migrations in `schema_migrations` table
- Auto-run pending migrations on startup
- Rollback on failure automatically

---

# Cluster Support (Non-Negotiable)

**ALL apps MUST support cluster mode (multiple instances).**

## Single Instance (Default - Auto-detected)
- No external cache/database configured
- Uses local file/SQLite for state

## Cluster Mode (Auto-detected)
- **Auto-enabled** when external cache or shared database detected
- **Primary election**: Only primary runs cluster-wide tasks
- **Distributed locks**: Prevent race conditions
- **Session sharing**: Store sessions in cache or database

---

# Application Lifecycle

## Graceful Shutdown
- **Handle termination signals properly** (SIGTERM, SIGINT)
- Stop accepting new requests
- Complete in-flight requests (with timeout)
- Close database connections gracefully
- Maximum shutdown time: 30 seconds

---

# Project-Specific API Endpoints

{Define your project's unique endpoints here}

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/{resource}` | GET | None | List resources |
| `/api/v1/{resource}/{id}` | GET | None | Get single resource |
| `/api/v1/{resource}/random` | GET | None | Get random resource |
| `/api/v1/{resource}/search` | GET | None | Search resources |

---

# Project-Specific Data Files

## Embedded Data (in binary)

| File | Location | Description |
|------|----------|-------------|
| `{data}.json` | `src/data/` | Main data file |

---

# Project-Specific Configuration

{Add any configuration options unique to this project}

```yaml
# Project-specific settings
{projectname}:
  # Custom settings here
```

---

# Security Headers (Non-Negotiable)

All responses MUST include appropriate security headers:

```
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

**In development mode**, these may be relaxed for testing.

---

# Logging (Non-Negotiable)

## Log Files

| Log | Purpose | Format |
|-----|---------|--------|
| `access.log` | HTTP requests | Apache combined |
| `server.log` | Application events | Text |
| `audit.log` | Security events | JSON |
| `security.log` | Fail2ban compatible | fail2ban |
| `debug.log` | Debug output (dev mode) | Text |

## Log Rotation
- Built-in log rotation support
- Configurable max size and retention
- Compress old logs

---

# Backup & Restore (Non-Negotiable)

## Backup Command
```bash
{projectname} --maintenance backup [filename]
```

## Backup Contents
- Configuration file
- Database (if applicable)
- Custom assets
- SSL certificates (optional, configurable)

## Backup Format
- Single `.tar.gz` file
- Includes manifest with version info
- Encrypted option available

## Restore Command
```bash
{projectname} --maintenance restore <backup-file>
```

## Update Command (--update)

```bash
{projectname} --update [command]
```

**Alias:** `--maintenance update` is an alias for `--update yes`

**Commands:**
- **`yes`** (default) - Check for update, if available perform in-place update with restart
  - Returns exit code 0 on successful update or no update available
  - Returns exit code 1 on error
  - HTTP 404 from GitHub API means no updates available (already current)
- **`check`** - Check for available updates without installing
  - Queries GitHub API for releases based on current branch
  - Shows current version, available version, and changelog
  - Returns exit code 0 if update available or already current, 1 on error
  - HTTP 404 from GitHub API means no updates available (already current)
  - Can be run by anyone (no privileges required)
- **`branch {stable|beta|daily}`** - Set update branch
  - **stable** (default): Tagged releases (e.g., `v1.0.0`, `1.0.0`)
  - **beta**: Beta releases (e.g., `202512051430-beta`)
  - **daily**: Daily/dev builds (e.g., `202512051430`)
  - Saved to config file for future updates

**Examples:**
```bash
# Check for updates - no privileges required
{projectname} --update check

# Perform update if available (these are equivalent)
{projectname} --update
{projectname} --update yes
{projectname} --maintenance update

# Switch to beta channel
{projectname} --update branch beta
```

---

# Health Checks (Non-Negotiable)

## /healthz (HTML)
Returns styled HTML page with:
- Status (healthy/unhealthy)
- Uptime
- Version
- Mode (production/development)
- System resources (optional)

## /api/v1/healthz (JSON)
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "mode": "production",
  "uptime": "2d 5h 30m",
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "database": "ok",
    "cache": "ok",
    "disk": "ok"
  }
}
```

---

# Versioning (Non-Negotiable)

## Version Format
- Semantic versioning: `MAJOR.MINOR.PATCH`
- Pre-release: `1.0.0-beta.1`
- Build metadata: `1.0.0+build.123`

## Version Sources
1. `release.txt` in project root
2. Git tag (if available)
3. Fallback: `dev`

## --version Output
```
{projectname} v1.0.0
Built: 2024-01-15T10:30:00Z
Go: 1.23
OS/Arch: linux/amd64
```

---

# Error Handling (Non-Negotiable)

## User-Facing Errors
- Clear, actionable messages
- No stack traces in production
- Appropriate HTTP status codes
- Consistent format

## Internal Errors
- Full stack traces in logs
- Context information
- Request ID for correlation

## Error Codes
Use consistent error codes:
- `ERR_VALIDATION` - Input validation failed
- `ERR_NOT_FOUND` - Resource not found
- `ERR_UNAUTHORIZED` - Authentication required
- `ERR_FORBIDDEN` - Permission denied
- `ERR_INTERNAL` - Server error
- `ERR_RATE_LIMIT` - Rate limit exceeded

---

# Caching (Non-Negotiable)

## Cache Drivers
- `memory` - In-memory (single instance)
- `redis` - Redis (cluster mode)
- `memcached` - Memcached (cluster mode)

## Cache Headers
- Static assets: `Cache-Control: max-age=31536000`
- API responses: `Cache-Control: no-cache` or as appropriate
- HTML pages: `Cache-Control: no-store`

---

# Internationalization (i18n)

## Support
- UTF-8 everywhere
- Accept-Language header respected
- Default: English (en)
- Extensible translation system

---

# Accessibility (a11y)

## Requirements
- WCAG 2.1 AA compliance
- Keyboard navigation
- Screen reader support
- Proper ARIA labels
- Color contrast ratios
- Focus indicators

---

# Notes

{Any additional notes, decisions, or context for this project}

---

**Remember: These are NON-NEGOTIABLE and NON-REPLACEABLE rules.**

**Because we are only defining the SPEC there is no need to show any code.**

**Ensure you read everything thoroughly.**

**The working directory is `.`**

**Any questions ask, never assume or guess.**

**You will always use a TODO.AI.md when doing more than 2 tasks.**

**You are to ALWAYS keep AI.md in sync with the project.**

---

# AI Assistant Rules

## Attribution
- **NEVER include AI attribution** in code, comments, commits, or documentation
- No "authored by Claude", "generated by AI", "Co-Authored-By: Claude", etc.
- Code should appear as if written by a human developer

## Tool Access
- Full access to all tools
- **EXCEPT**: `git push`, `git commit` - these require explicit user request
- Can stage files, create branches, check status, diff, etc.
- User must explicitly request commits and pushes
