# CASCI

[![CI](https://github.com/webappsgo/casci/actions/workflows/ci.yml/badge.svg)](https://github.com/webappsgo/casci/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/webappsgo/casci)](https://github.com/webappsgo/casci/releases)
[![License](https://img.shields.io/github/license/webappsgo/casci)](LICENSE.md)

## About

CASCI is an open-source CI/CD application server written in Go. The repository
contains a server-rendered web interface, a versioned REST API, a GraphQL
endpoint, pipeline parsers, build and node services, and supporting packages.
The project is actively under development; the feature list below describes
implemented code and does not imply complete compatibility with the systems
CASCI aims to support.

## Official Site

[https://casci.casapps.us](https://casci.casapps.us)

## Features

- Server-rendered home, health, account sign-in, preferences, help, legal, and
  security pages
- Installable PWA shell with build-stamped assets, public-page offline fallback,
  and user-controlled service worker updates
- SQLite, PostgreSQL, and MySQL database drivers
- User registration and login, named per-user `usr_` API tokens with one-time
  display and scoped revocation, project CRUD, and build lifecycle API handlers
- First-run server administrator setup, separate administrator sessions and
  `adm_` tokens, administrator token management, TOTP, and WebAuthn passkeys
- Configurable OIDC, LDAP, and SAML sign-in, with provider account linking and
  explicit account-onboarding and role-mapping flows
- Encrypted security-report intake and review, scheduled and manual backups,
  user-managed infrastructure connection metadata, and best-effort budgets
- Container executor and build queue, with local build-node management routes
- Parsers for several common CI configuration formats; parser support does not
  mean full runtime compatibility with those CI systems
- Credential encryption, notification configuration, security scan models and
  handlers, audit logging, compliance configuration, and metrics packages
- Versioned REST endpoints, Swagger UI, a generated OpenAPI endpoint, GraphQL
  endpoint, and authenticated Streamable HTTP MCP endpoint
- Companion `casci-cli` for health checks, owner-scoped project list/create,
  login, and authenticated versioned API requests
- Docker image and Compose configurations, plus service installer scripts

## Production

### Docker

The Compose file publishes the application on the Docker bridge at
`172.17.0.1:64580`; place a reverse proxy in front if the service must be
reachable outside the host.

```bash
curl -q -LSsf -o docker-compose.yml \
  https://raw.githubusercontent.com/webappsgo/casci/main/docker/docker-compose.yml
docker compose -f docker-compose.yml up -d
```

The Compose setup also starts Valkey. The application currently uses its
configured SQLite database by default; supported database drivers are SQLite,
PostgreSQL, and MySQL. Review [Configuration](docs/configuration.md) before
selecting an external database.

### Binary

Build the current-platform server and client with `make dev` or the release
platform set with `make build`. Builds run in Docker using
`casjaysdev/go:latest`. The platform targets are Linux, macOS, Windows, and
FreeBSD on amd64 and arm64. Build output from `make build` is written to the
ignored `binaries/` directory.

### Service installers

The server binary can install and control supported native service managers
with `--service --install`, `start`, `stop`, `restart`, `reload`, `--disable`,
and `--uninstall`. Review [Service management](docs/services.md) before using
system-level actions; uninstall removes application data after confirmation.

## Client

The companion `casci-cli` supports health checks, owner-scoped project listing
and creation, user login, and authenticated versioned API requests. Build it
with `make dev` for the current platform or `make build` for release targets.
For example, `casci-cli --server https://casci.casapps.us status` checks the
official server. See the [CLI reference](docs/cli.md) for authentication,
configuration, and command details.

## Configuration

CASCI loads `server.yml` and `yaml.d/` overrides from its resolved
configuration directory. Current configuration covers server address and
port, mode, API and admin path segments, branding and page content, TLS,
database connections, CORS, security contacts and policies, user registration,
identity providers, SMTP, scheduler tasks, backups, updates, and per-kind log
formats and rotation. Some server settings remain read-only in the admin UI;
the complete database-backed, hot-reloadable configuration system described
in the product specification is not implemented.

See [Configuration](docs/configuration.md) for current paths, environment
variables, and supported settings.

## API

The default API prefix is `https://casci.casapps.us/api/v1`. The version segment
is configurable. Current route groups include authentication and external
identity, users, projects, builds, nodes, credentials, notifications,
infrastructure connections and budgets, security reports, audit, compliance,
metrics, server health/help/preferences, and MCP. The OpenAPI
document is currently incomplete; use the route list in the
[API Reference](docs/api.md) as the current implemented surface.

- Health: `GET https://casci.casapps.us/api/v1/server/healthz`
- Help: `GET https://casci.casapps.us/api/v1/server/help`
- MCP: `POST https://casci.casapps.us/api/v1/server/mcp`
- OpenAPI JSON: `GET https://casci.casapps.us/api/v1/server/swagger`
- Swagger UI: `GET https://casci.casapps.us/server/docs/swagger`
- GraphQL: `https://casci.casapps.us/api/v1/server/graphql` (resolver coverage
  is incomplete)

## Other

- [Installation and service setup](docs/installation.md)
- [Configuration reference](docs/configuration.md)
- [REST and GraphQL API reference](docs/api.md)
- [CLI reference](docs/cli.md)
- [Agent management](docs/agents.md)
- [Admin panel](docs/admin.md)
- [Security model and current limitations](docs/security.md)
- [Integrations and protocol support](docs/integrations.md)
- [Troubleshooting](docs/index.md#current-limitations)

## Development

Go builds and tests run in Docker. The Makefile provides these primary targets:

```bash
make dev
make build
make test
```

`make test` runs `go test ./...` with coverage enforcement and currently
requires at least 60% total statement coverage. See
[Development](docs/development.md) for toolchain, test, and contribution notes.
Run `scripts/verify-licenses.sh` inside `casjaysdev/go:latest` to scan dependency
licenses and print the temporary report.

## Disclaimer

This software is provided “as is,” without warranty of any kind. Use it at your
own risk and review the implementation before production deployment.

- **No warranty:** The authors are not responsible for damages, data loss, or
  other issues arising from use of this software.
- **No professional advice:** CASCI does not provide legal, financial, medical,
  or other professional advice.
- **Third-party services:** External services have their own terms and privacy
  policies.
- **Security:** No software can be guaranteed free of vulnerabilities. Assess
  the current implementation and its limitations before exposing an instance.
- **Production use:** Test the deployment, backups, access controls, and
  recovery process before relying on it.

By using this software, you acknowledge that you have read and understood this
disclaimer.

## License

CASCI is licensed under the MIT License. See [LICENSE.md](LICENSE.md).
