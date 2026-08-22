# Admin Panel

## Access

| Property | Value |
|----------|-------|
| URL | `https://casci.casapps.us/server/administration` |
| Authentication | Session-based, administrator flag required |
| First run | The first account created on a fresh install becomes the administrator |

Non-admin sessions receive `403`; unauthenticated requests receive `401`. Every admin page is behind the same admin check, so there is no partially protected surface. An unmatched path below the admin root is also admin-gated, so probing for admin pages reveals nothing to an anonymous caller.

The default administrator username is `administrator` and can be changed.

!!! note "`administration` is not a fixed literal"
    The path segment comes from `server.admin_path` in `server.yml`. If you change it, substitute your own value everywhere `administration` appears on this page. Moving it is obscurity, not access control — the admin check runs either way.

## Roles

CASCI has a deliberately simple permission model: a single `is_admin` boolean on the user record. There is no role table, no general-purpose RBAC, and no service-account type.

| Role | Can do | Cannot do |
|------|--------|-----------|
| Administrator | Manage server configuration, nodes, users, security policy, audit retention, and compliance mode | Read another user's projects, credentials, or logs |
| Regular user | Create and run their own projects and builds, manage their own credentials and infrastructure connections | See any other user's resources, or reach any admin endpoint |

The administrator flag decides which endpoints are reachable, not which rows are visible. An admin manages the server, not the contents of other people's projects.

## Pages

| Page | Path | Purpose |
|------|------|---------|
| Dashboard | `/server/administration` | Project, build, user, and node counts plus system metrics |
| Settings | `/server/administration/config/settings` | Server configuration |

Those two are the whole panel. Everything else administrative is API-only, listed under [Admin API](#admin-api) below.

Secret values are never rendered into a template. The encryption key and the database connection string are stripped from the settings page before it is served, so a template bug or a cached page cannot leak them.

## What Is Configurable

Settings that can be managed from the panel:

- Server address, port, and mode
- Application title and tagline
- Admin path segment and API version segment
- Database driver, connection string, and pool sizing
- TLS enablement, certificate and key paths, and minimum version
- CORS allowlist
- Webhook shared secret
- Encryption key rotation
- Security contact fields published in `security.txt`
- Footer branding, analytics tracking ID, and cookie consent text
- Security scan policy, including block-on-critical behavior
- Audit retention and cleanup
- Compliance mode selection
- Notification providers and event routing
- Build node registration and draining

Anything settable through the panel is also settable in `server.yml`. Only a small, closed set has an environment variable, and several of those are read once on first run and ignored afterwards. See [Configuration](configuration.md) for the full list and the precedence rules.

## Admin API

Every admin capability is also reachable programmatically with bearer-token authentication:

| Area | Endpoint prefix |
|------|-----------------|
| Users | `/api/v1/users` |
| Nodes | `/api/v1/nodes` |
| Security policy | `/api/v1/security/config` |
| Audit | `/api/v1/audit/config`, `/api/v1/audit/status`, `/api/v1/audit/cleanup` |
| Compliance | `/api/v1/compliance/mode`, `/api/v1/compliance/config`, `/api/v1/compliance/check` |
| Metrics | `/metrics`, `/metrics/json`, `/api/v1/metrics/*` |

```bash
curl -q -LSsf https://casci.casapps.us/api/v1/nodes \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

See [API Reference](api.md) for the complete endpoint list and required roles.

## Public Server Pages

These pages sit alongside the admin panel under `/server/` but are public:

| Path | Purpose |
|------|---------|
| `/server/healthz` | Human-readable health page |
| `/server/about` | About the instance |
| `/server/privacy` | Privacy policy |
| `/server/contact` | Contact form |
| `/server/help` | Help and documentation |

## Audit Trail

Administrative actions are recorded in the audit log. Query it from `/api/v1/audit/events`, check subsystem state at `/api/v1/audit/status`, and manage retention at `/api/v1/audit/config`.

Retention is driven by the active compliance mode: HIPAA expects six years, SOX seven, and PCI-DSS one year of access logging. Selecting a mode applies its retention expectation rather than requiring it to be configured by hand. See [Security](security.md).
