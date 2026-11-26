# CASCI documentation

CASCI is a Go CI/CD application server under active development. These pages
describe the implementation in this repository; they distinguish available
routes and components from planned product capabilities.

## Start here

- [Install CASCI](installation.md) with the provided Docker Compose files or a
  locally built binary.
- Review [configuration](configuration.md) before exposing an instance.
- Use the [API reference](api.md) for the currently mounted HTTP routes.
- Read the [security notes](security.md) before running pipelines from
  untrusted repositories.

## Current application surface

The server exposes a public web shell, sign-in, user profile, preference and
help pages, an admin dashboard, settings, Pages → Help and user management,
versioned API handlers, health and readiness endpoints, Swagger UI, GraphQL,
MCP, webhooks, and Jenkins-shaped compatibility endpoints. The browser shell
includes an installable PWA with versioned static assets, offline fallback for
eligible public pages, and a user-controlled service worker update prompt.
Several subsystems are incomplete; consult the limitations below and the
relevant reference pages before relying on them.

## Current limitations

- The project is not yet a production-complete replacement for Jenkins or the
  other CI systems whose formats it parses.
- The OpenAPI document is currently empty/incomplete, and GraphQL resolver
  coverage is incomplete.
- The contact form currently validates a basic form but cannot deliver reports.
- MCP tool routing covers only a subset of the API and does not yet implement
  all required token scopes, rate limits, or audit integration.
- Build isolation, multi-user isolation, workload fairness, provider
  provisioning, and several security controls remain incomplete.
- Shell completion, native service lifecycle commands, backup and restore
  maintenance commands, and staged update operations are implemented. Service
  manager support varies by operating system; see the [CLI](cli.md) and
  [service](services.md) references.
- The configured build test gate requires at least 60% statement coverage.
  The current suite has not yet met that threshold.

## Documentation pages

- [Installation](installation.md)
- [Configuration](configuration.md)
- [API reference](api.md)
- [API quick reference](api-quickref.md)
- [CLI reference](cli.md)
- [Admin panel](admin.md)
- [Security](security.md)
- [Integrations](integrations.md)
- [Service installers](services.md)
- [Development](development.md)
- [Translation guidance](translating.md)

The official site is [https://casci.casapps.us](https://casci.casapps.us).
