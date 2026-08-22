# casci

CASCI is a complete CI/CD platform delivered as a single static Go binary. It runs existing Jenkins, GitHub Actions, GitLab CI, Travis CI, CircleCI, Azure Pipelines, Bitbucket Pipelines, Drone CI, and Buildkite pipelines without rewriting them, needs zero configuration on first run, and ships with security scanning, a credential vault, and compliance auditing built in.

## Quick Start

```bash
# Docker
docker run --name "casci-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" -p 172.17.0.1:64580:80 ghcr.io/webappsgo/casci:latest

# Binary
./casci-linux-amd64
```

The binary needs no flags. On first run it creates its config, data, cache, log, and backup directories, selects a random port in the `64000-64999` range, and persists that port to `server.yml` so it stays stable across restarts. Inside a container the port defaults to `80`.

## Features

- Multi-format pipeline engine covering Jenkinsfile, GitHub Actions, GitLab CI, CircleCI, Travis CI, Azure Pipelines, Bitbucket Pipelines, Drone CI, and Buildkite
- Project auto-detection for Go, JavaScript, Python, Java, Rust, Ruby, and PHP when no pipeline file is present
- Jenkins-compatible REST surface at `/api/json` and `/crumbIssuer/api/json`
- Container-isolated build execution with per-build workspace cleanup
- Build queue with dispatch to local execution or registered remote build nodes
- Artifact storage and retrieval per build
- Git clone, branch/tag checkout, and commit metadata capture
- GitHub and GitLab webhook signature verification, with Gitea reusing the GitHub payload format
- Security scanning by delegating to `trivy`, `grype`, `gitleaks`, `semgrep`, and `syft` on the host
- Encrypted credential vault scoped per user and per project
- Notifications across chat, email, issue trackers, commit status, and paging providers
- Compliance modes for HIPAA, SOX, PCI-DSS, GDPR, FedRAMP, and ISO 27001
- Audit log with admin-managed retention and cleanup
- System, build, node, security, and API metrics
- REST API under `/api/v1/` and Swagger UI with OpenAPI output
- WebSocket build log streaming
- SQLite, PostgreSQL, and MySQL storage backends

## Current Limitations

CASCI is under active development. These parts are mounted but not yet functional, and are called out again on the page that covers each one:

- **GraphQL** resolvers are stubs; the endpoint and playground exist but return nothing
- **Webhooks** verify signatures but cannot yet resolve a repository URL to a project, so they do not queue builds
- **Bitbucket** webhook events are acknowledged and discarded
- **S3, GCS, and Azure** artifact backends return "not yet implemented"; local disk storage is the working backend
- **`--service`, `--maintenance`, `--update`, and `--shell`** report that they are not implemented; use the platform installers in [Installation](installation.md)

## Documentation

- [Installation](installation.md) - How to install and run
- [Configuration](configuration.md) - All configuration options
- [API Reference](api.md) - REST API, Swagger, GraphQL
- [CLI Reference](cli.md) - Server binary flags
- [Admin Panel](admin.md) - WebUI administration
- [Security](security.md) - Auth, public endpoints, reporting, and hardening behavior
- [Integrations](integrations.md) - Webhooks, notifications, discovery, and platform integrations
- [Development](development.md) - Contributing guide

## Links

- [Repository](https://github.com/webappsgo/casci)
- [Official Site](https://casci.casapps.us)
- [API Documentation](https://casci.casapps.us/server/docs/swagger) (Swagger UI)
- [GraphQL Playground](https://casci.casapps.us/server/docs/graphql)

## License

MIT - See [LICENSE.md](https://github.com/webappsgo/casci/blob/main/LICENSE.md)
