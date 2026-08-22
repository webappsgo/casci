# casci

[![Build](https://github.com/webappsgo/casci/actions/workflows/ci.yml/badge.svg)](https://github.com/webappsgo/casci/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/webappsgo/casci)](https://github.com/webappsgo/casci/releases/latest)
[![License](https://img.shields.io/github/license/webappsgo/casci)](LICENSE.md)
[![Documentation](https://readthedocs.org/projects/webappsgo-casci/badge/?version=latest)](https://webappsgo-casci.readthedocs.io)

## About

CASCI is a complete CI/CD platform delivered as a single static Go binary. It runs existing Jenkins, GitHub Actions, GitLab CI, CircleCI, Travis CI, and Drone pipelines unchanged, needs no configuration files, and scales from a laptop to a cluster of 1000+ build nodes.

## Official Site

https://casci.casapps.us

## Features

- Single static binary with no required external services; embedded SQLite by default, external PostgreSQL or MySQL opt-in
- Zero configuration on first run — no CLI flags and no hand-edited config files for normal operation
- Multi-format pipeline engine: Jenkinsfile, GitHub Actions, GitLab CI, CircleCI, Travis CI, Azure Pipelines, Bitbucket Pipelines, Drone CI, and Buildkite parsed into one internal pipeline model
- Project auto-detection for Go, JavaScript, Python, Java, Rust, Ruby, and PHP when a repository ships no pipeline file
- Jenkins-compatible REST surface at `/api/json` and `/crumbIssuer/api/json`
- Container-isolated build execution with per-build workspace cleanup
- Build queue dispatching to local execution or registered remote build nodes
- Distributed build nodes with token-based join, scaling to 1000+ agents
- Artifact storage and retrieval per build
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

## Production

### Docker (Recommended)

```bash
docker run -d \
  --name casci \
  -p 172.17.0.1:64580:80 \
  -v ./volumes/config:/config:z \
  -v ./volumes/data:/data:z \
  ghcr.io/webappsgo/casci:latest
```

### Docker Compose

```bash
curl -q -LSsf -O https://raw.githubusercontent.com/webappsgo/casci/main/docker/docker-compose.yml
docker compose up -d
```

### Binary

```bash
# Download latest release
curl -q -LSsf -O https://github.com/webappsgo/casci/releases/latest/download/casci-linux-amd64

# Make executable and run
chmod +x casci-linux-amd64
./casci-linux-amd64
```

On first run CASCI selects a random port in the `64000-64999` range and persists it to `server.yml`, so the port stays stable across restarts. Inside a container the port defaults to `80`.

## Configuration

Configuration is auto-generated on first run. Edit it via the admin panel at
`https://casci.casapps.us/server/administration`.

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

The config file is `server.yml` inside the config directory. It is generated on first run, written with mode `0600` because it holds the encryption key and the webhook secret, and is the persisted source of truth for everything not listed as an environment variable below.

### Environment Variables

Environment variables are unprefixed. The list is closed — anything not on it is a `server.yml` key, and setting it in the environment has no effect.

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

Everything else is configured in `server.yml` or from the admin panel: `server.cors_origins`, `server.tls.*`, `server.security.encryption_key`, `server.security.webhook_secret`, `server.footer.*`, and `server.database.max_conns` / `max_idle` / `replica_urls`. See [Configuration](https://webappsgo-casci.readthedocs.io/en/latest/configuration/).

## CLI

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

## API

API documentation is available at `https://casci.casapps.us/api/v1/` when running.

| Endpoint | Description |
|----------|-------------|
| `GET https://casci.casapps.us/server/healthz` | Health check |
| `GET https://casci.casapps.us/api/v1/server/healthz` | Health check (JSON) |
| `POST https://casci.casapps.us/api/v1/auth/login` | Obtain a JWT |
| `GET https://casci.casapps.us/api/v1/projects` | List projects |
| `GET https://casci.casapps.us/api/v1/builds/` | Build operations |
| `POST https://casci.casapps.us/webhook` | SCM webhook receiver (GitHub, GitLab, Bitbucket, Gitea) |
| `POST https://casci.casapps.us/graphql` | GraphQL endpoint |

### Examples

```bash
# Health check
curl -q -LSsf https://casci.casapps.us/server/healthz

# API request (requires auth)
curl -q -LSsf -H "Authorization: Bearer TOKEN" https://casci.casapps.us/api/v1/projects
```

Authentication uses the `Authorization: Bearer` header. Tokens are never accepted as a query parameter.

## Other

### Troubleshooting

- Check logs: `docker logs casci`
- Health check: `curl -q -LSsf https://casci.casapps.us/server/healthz`
- Local status: `casci --status`

### Documentation

Full documentation: https://webappsgo-casci.readthedocs.io

## Development

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

### Project Structure

```
src/           # Source code
tests/         # Test files
docker/        # Dockerfile and docker-compose.yml
docs/          # Documentation
binaries/      # Built binaries (gitignored)
```

## Disclaimer

This software is provided "as is" without warranty of any kind. Use at your own risk.

- **No Warranty**: The authors are not responsible for any damages, data loss, or issues arising from use of this software
- **Not Professional Advice**: This software does not constitute legal, financial, medical, or other professional advice
- **Third-Party Services**: If this software connects to external APIs or services, their terms of service apply separately
- **Security**: While we strive to follow security best practices, no software is guaranteed to be free of vulnerabilities
- **Production Use**: Evaluate thoroughly before deploying in production environments

By using this software, you acknowledge that you have read and understood this disclaimer.

## License

MIT - See [LICENSE.md](LICENSE.md)
