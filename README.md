# casci

CASCI (CI/CD Application Server for Continuous Integration) is a complete
CI/CD platform delivered as a single static Go binary. It replaces Jenkins,
GitHub Actions, GitLab CI, CircleCI, Travis CI, and similar tools for teams
and individuals who want one deployable that runs existing CI/CD pipelines
unchanged, requires no configuration files, and scales from a laptop to a
cluster of 1000+ build nodes.

[![Build](https://github.com/webappsgo/casci/actions/workflows/ci.yml/badge.svg)](https://github.com/webappsgo/casci/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/webappsgo/casci)](https://github.com/webappsgo/casci/releases/latest)
[![License](https://img.shields.io/github/license/webappsgo/casci)](LICENSE.md)
[![Documentation](https://readthedocs.org/projects/webappsgo-casci/badge/?version=latest)](https://webappsgo-casci.readthedocs.io)

---

## 📦 Install

Download the latest release from [GitHub Releases](https://github.com/webappsgo/casci/releases/latest).

### Linux
| Arch | Binary |
|------|--------|
| amd64 | `casci-linux-amd64` |
| arm64 | `casci-linux-arm64` |

```bash
curl -LSsf https://github.com/webappsgo/casci/releases/latest/download/casci-linux-amd64 \
  -o /usr/local/bin/casci && chmod +x /usr/local/bin/casci
```

### macOS
| Arch | Binary |
|------|--------|
| Intel (x86_64) | `casci-darwin-amd64` |
| Apple Silicon (arm64) | `casci-darwin-arm64` |

```bash
# Detect arch automatically
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -LSsf "https://github.com/webappsgo/casci/releases/latest/download/casci-darwin-${ARCH}" \
  -o /usr/local/bin/casci && chmod +x /usr/local/bin/casci
# Remove macOS quarantine flag
xattr -d com.apple.quarantine /usr/local/bin/casci 2>/dev/null || true
```

### Windows
| Arch | Binary |
|------|--------|
| amd64 | `casci-windows-amd64.exe` |
| arm64 | `casci-windows-arm64.exe` |

Download and add to `%PATH%`.

### FreeBSD
| Arch | Binary |
|------|--------|
| amd64 | `casci-freebsd-amd64` |
| arm64 | `casci-freebsd-arm64` |

```bash
curl -LSsf https://github.com/webappsgo/casci/releases/latest/download/casci-freebsd-amd64 \
  -o /usr/local/bin/casci && chmod +x /usr/local/bin/casci
```

On first run CASCI selects a random port in the `64000-64999` range and
persists it to `server.yml`, so the port stays stable across restarts.

---

## 🐳 Docker

```bash
docker run -d \
  --name casci \
  -p 172.17.0.1:64580:80 \
  -v ./volumes/config:/config:z \
  -v ./volumes/data:/data:z \
  ghcr.io/webappsgo/casci:latest
```

Inside a container the port defaults to `80`.

### Docker Compose

```bash
curl -q -LSsf -O https://raw.githubusercontent.com/webappsgo/casci/main/docker/docker-compose.yml
docker compose up -d
```

---

## Features

- Single static binary with no required external services; embedded SQLite by default, external PostgreSQL or MySQL opt-in
- Zero configuration on first run — no CLI flags and no hand-edited config files for normal operation
- Multi-format pipeline engine: Jenkinsfile (declarative and scripted, Blue Ocean and multibranch), GitHub Actions, GitLab CI, Travis CI, CircleCI, Azure Pipelines, Bitbucket Pipelines, Drone CI, Buildkite, Tekton, Woodpecker, and CodeFresh parsed into one internal pipeline model, with auto-generated pipelines for repositories that ship none of the above
- Jenkins-compatible REST surface at `/api/json` and `/crumbIssuer/api/json`, including Blue Ocean API and `jenkins-cli.jar`-style CLI operations
- Project auto-detection for Go, JavaScript, Python, Java, Rust, Ruby, and PHP when a repository ships no pipeline file
- Container-isolated build execution with per-build workspace cleanup and no storage shared between users' builds
- Build queue dispatching to local execution or registered remote build nodes, scaling to 1000+ agents via token-based join
- On-demand build capacity on AWS, Google Cloud, Azure, Oracle Cloud, Vultr, Hetzner Cloud, DigitalOcean, and Linode, with instances destroyed as soon as a build completes or times out
- Artifact storage and retrieval per build, with every build signed, SBOM-tracked, and scanned for vulnerabilities, secrets, and license issues
- Git clone, branch and tag checkout, and commit metadata capture
- GitHub, GitLab, Bitbucket, and Gitea webhook triggers with signature verification
- Security scanning built into the binary: vulnerability, SAST, secret, license, SBOM, and malware scans with severity classification
- Encrypted credential vault scoped per user and per project
- Notifications to Slack, Discord, Teams, Telegram, Matrix, IRC, Mattermost, email, Jira, GitHub/GitLab issues, Linear, Asana, GitHub/GitLab/Bitbucket/Gitea commit status, PagerDuty, OpsGenie, VictorOps, generic webhooks, and scripts
- Compliance modes for HIPAA, SOX, PCI-DSS, GDPR, FedRAMP, and ISO 27001, with per-mode audit retention
- Audit log with admin-managed retention and cleanup
- System, build, node, security, and API metrics with Prometheus-style and JSON output
- Web UI, REST API, and GraphQL API over the same data, plus Swagger UI and OpenAPI JSON/YAML
- WebSocket build log streaming
- Self-healing operation: automatic error recovery and graceful degradation

---

## ⌨️ CLI Usage

`casci` is a single binary — the same executable serves the application and
its own command-line interface.

```bash
casci --help
casci --version
casci --status
```

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Show help |
| `--version`, `-v` | Show version |
| `--status` | Show server status and health (exit `0` healthy, `1` unhealthy) |
| `--shell {completions,init,--help} [SHELL]` | Shell integration |
| `--mode {production,development,debug}` | Application mode |
| `--config DIR` | Config directory |
| `--data DIR` | Data directory |
| `--cache DIR` | Cache directory |
| `--log DIR` | Log directory |
| `--backup DIR` | Backup directory |
| `--pid FILE` | PID file path |
| `--address ADDR` | Listen address |
| `--port PORT` | Listen port |
| `--baseurl PATH` | URL path prefix |
| `--daemon` | Run as daemon (detach from terminal) |
| `--debug` | Enable debug mode |
| `--color {auto,yes,no}` | Color output |
| `--lang CODE` | Language for output |
| `--service {start,restart,stop,reload,--install,--uninstall,--disable,--help}` | Service management |
| `--maintenance {backup,restore,update,mode,setup,--help}` | Maintenance operations |
| `--update [check\|yes\|branch {stable,beta,daily}\|--help]` | Check or perform updates |

---

## ⚙️ Configuration

Configuration is auto-generated on first run and managed from the admin
panel (`/server/administration` on your instance).

Key settings:

- `server.port` — Listen port (default: random port in `64000-64999`, persisted to `server.yml` after first run; `80` in a container)
- `server.mode` — `production`, `development`, or `debug`

### Paths

| Purpose | Linux/BSD (root) | Linux/BSD (user) |
|---------|------------------|------------------|
| Config | `/etc/casapps/casci/` | `~/.config/casapps/casci/` |
| Data | `/var/lib/casapps/casci/` | `~/.local/share/casapps/casci/` |
| Cache | `/var/cache/casapps/casci/` | `~/.cache/casapps/casci/` |
| Logs | `/var/log/casapps/casci/` | `~/.local/log/casapps/casci/` |

The config file is `server.yml` inside the config directory. It is generated
on first run, written with mode `0600` because it holds the encryption key
and the webhook secret, and is the persisted source of truth for everything
not listed as an environment variable below.

### Environment Variables

Environment variables are unprefixed. The list is closed — anything not on
it is a `server.yml` key, and setting it in the environment has no effect.

**Runtime — re-read on every start:**

| Variable | Description | Default |
|----------|-------------|---------|
| `NO_COLOR` | Disable ANSI color and emoji output when set and non-empty | unset |
| `TERM` | Terminal type; `TERM=dumb` disables all ANSI escapes | inherited |
| `DOMAIN` | Public FQDN, also the default CORS origin | unset (same-origin only) |
| `MODE` | `production`, `development`, or `debug` | `production` |
| `DATABASE_DRIVER` | `sqlite`, `libsql`, `postgres`, `mysql`, `mssql`, or `mongodb` | `sqlite` |
| `DATABASE_URL` | Driver-specific DSN | embedded SQLite file |
| `SMTP_HOST` | SMTP server hostname | autodetected |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_USERNAME` | SMTP username | unset |
| `SMTP_PASSWORD` | SMTP password | unset |
| `SMTP_FROM_NAME` | Sender name | application title |
| `SMTP_FROM_EMAIL` | Sender address | `no-reply@<fqdn>` |
| `SMTP_TLS` | `auto`, `starttls`, `tls`, or `none` | `auto` |

**Init-only — read once on first run, persisted to `server.yml`, ignored afterwards:**

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_DIR` | Configuration directory | `/etc/casapps/casci` |
| `DATA_DIR` | Data directory | `/var/lib/casapps/casci` |
| `LOG_DIR` | Log directory | `/var/log/casapps/casci` |
| `DATABASE_DIR` | Embedded-database directory | under `DATA_DIR` |
| `BACKUP_DIR` | Backup output directory | under `DATA_DIR` |
| `PORT` | Listen port | random 64000-64999, then persisted |
| `LISTEN` | Bind address | `0.0.0.0` |
| `APPLICATION_NAME` | Display name | `casci` |
| `APPLICATION_TAGLINE` | Display tagline | unset |

Everything else is configured in `server.yml` or from the admin panel:
`server.cors_origins`, `server.tls.*`, `server.security.encryption_key`,
`server.security.webhook_secret`, `server.footer.*`, and
`server.database.max_conns` / `max_idle` / `replica_urls`. See
[Configuration](https://webappsgo-casci.readthedocs.io/en/latest/configuration/).

---

## 🔌 API

| Endpoint | Description |
|----------|-------------|
| `GET /server/healthz` | Health check |
| `GET /api/v1/server/healthz` | Health check (JSON) |
| `POST /api/v1/auth/login` | Obtain a JWT |
| `GET /api/v1/projects` | List projects |
| `GET /api/v1/builds/` | Build operations |
| `POST /webhook` | SCM webhook receiver (GitHub, GitLab, Bitbucket, Gitea) |
| `POST /graphql` | GraphQL endpoint |

### Examples

```bash
# Health check
curl -q -LSsf https://your-instance.example.com/server/healthz

# API request (requires auth)
curl -q -LSsf -H "Authorization: Bearer TOKEN" https://your-instance.example.com/api/v1/projects
```

Authentication uses the `Authorization: Bearer` header. Tokens are never
accepted as a query parameter.

---

## 🩺 Troubleshooting & Documentation

- Check logs: `docker logs casci`
- Health check: `curl -q -LSsf https://your-instance.example.com/server/healthz`
- Local status: `casci --status`
- Full documentation: https://webappsgo-casci.readthedocs.io

---

## ⚠️ Disclaimer

This software is provided "as is" without warranty of any kind. Use at your
own risk.

- **No Warranty**: The authors are not responsible for any damages, data loss, or issues arising from use of this software
- **Not Professional Advice**: This software does not constitute legal, financial, medical, or other professional advice
- **Third-Party Services**: If this software connects to external APIs or services, their terms of service apply separately
- **Security**: While we strive to follow security best practices, no software is guaranteed to be free of vulnerabilities
- **Production Use**: Evaluate thoroughly before deploying in production environments

By using this software, you acknowledge that you have read and understood
this disclaimer.

---

## 🛠️ Development

**Development instructions are for contributors only.**

### Prerequisites

- Go (latest stable)
- Docker (for containerized builds)

### Build

```bash
# Clone
git clone https://github.com/webappsgo/casci
cd casci

# Quick dev build (outputs to OS temp dir)
make dev

# Full build (all platforms, outputs to binaries/)
make build

# Test
make test
```

| Target | Description |
|--------|-------------|
| `make dev` | Quick local build to a temp dir |
| `make local` | Local-arch build to `binaries/` |
| `make build` | Full 8-platform build to `binaries/` |
| `make release` | Local build plus a GitHub release |
| `make docker` | Multi-arch Docker image build (no push) |
| `make test` | Run tests with coverage enforcement |

### Project Structure

```
src/           # Source code
tests/         # Test files
docker/        # Dockerfile and docker-compose.yml
docs/          # Documentation
binaries/      # Built binaries (gitignored)
```

### 🐳 Docker build

```bash
docker buildx build \
  -f docker/Dockerfile \
  --platform linux/amd64,linux/arm64 \
  -t ghcr.io/webappsgo/casci:latest \
  .
```

`make docker` runs the equivalent multi-arch build locally without pushing.

---

## 📄 License

MIT — see [LICENSE.md](LICENSE.md)
