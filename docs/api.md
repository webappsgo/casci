# API Reference

## Base URL

REST endpoints live under `/api/v1/`. Documentation examples use the official site; substitute your own host when self-hosting.

```
https://casci.casapps.us/api/v1/
```

The `v1` segment is not a literal. It comes from `server.api_version` in `server.yml` and defaults to `v1`. Every path on this page is written with the default; if you change that key, substitute your own segment.

## Authentication

Every authenticated request carries its credential in the `Authorization` header:

```
Authorization: Bearer <token>
```

Tokens are never accepted as a query parameter. A `?token=` style URL leaks the credential into proxy logs, browser history, and referrer headers, so it is not supported.

Two credential kinds are accepted:

| Kind | Obtained from | Lifetime |
|------|---------------|----------|
| JWT | `POST /api/v1/auth/login` | Short-lived; renew with `/api/v1/auth/refresh` |
| API token | `POST /api/v1/users/me/token` | Long-lived until regenerated |

### Register

The first account created on a fresh install becomes the administrator.

```bash
curl -q -LSsf -X POST https://casci.casapps.us/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"administrator","email":"admin@example.com","password":"CHANGE_ME"}'
```

### Login

```bash
curl -q -LSsf -X POST https://casci.casapps.us/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"administrator","password":"CHANGE_ME"}'
```

The response carries a JWT. Use it as the bearer token on subsequent requests.

### Refresh

```bash
curl -q -LSsf -X POST https://casci.casapps.us/api/v1/auth/refresh \
  -H "Authorization: Bearer $JWT"
```

### Regenerate the API Token

```bash
curl -q -LSsf -X POST https://casci.casapps.us/api/v1/users/me/token \
  -H "Authorization: Bearer $JWT"
```

Regenerating invalidates the previous API token immediately.

## Health

Health endpoints are public and unauthenticated.

| Endpoint | Method | Response | Description |
|----------|--------|----------|-------------|
| `/server/healthz` | GET | HTML | Human-readable health page |
| `/api/v1/server/healthz` | GET | JSON | Machine-readable health check |
| `/readyz` | GET | JSON | Readiness probe |
| `/livez` | GET | JSON | Liveness probe |

`/healthz` at the root is not mounted by default. Setting `server.healthz.root.enabled` to `true` in `server.yml` mounts it to the same handler as `/server/healthz` — an alias, not a redirect.

```bash
curl -q -LSsf https://casci.casapps.us/api/v1/server/healthz
```

Locally, `casci --status` performs the same check and exits `0` when healthy and `1` when not, which makes it usable directly as a container health check.

## Users

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/users/me` | GET | User | Current user profile |
| `/api/v1/users/me/token` | POST | User | Regenerate the API token |
| `/api/v1/users` | GET | Admin | List all users |
| `/api/v1/users/{id}` | GET | User | Fetch a user |
| `/api/v1/users/{id}` | PUT, PATCH | User | Update a user |
| `/api/v1/users/{id}` | DELETE | Admin | Delete a user |

Reading and updating an individual user requires authentication; only deletion requires the admin flag at the routing layer.

```bash
curl -q -LSsf https://casci.casapps.us/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

## Projects

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/projects` | GET | User | List the caller's projects |
| `/api/v1/projects/` | POST | User | Create a project |
| `/api/v1/projects/{id}` | GET | User | Fetch a project |
| `/api/v1/projects/{id}` | PUT, PATCH | User | Update a project |
| `/api/v1/projects/{id}` | DELETE | User | Delete a project |
| `/api/v1/projects/{id}/builds` | GET | User | List the project's builds |
| `/api/v1/projects/{id}/builds` | POST | User | Trigger a build |
| `/api/v1/projects/{id}/builds/stats` | GET | User | Build statistics |
| `/api/v1/projects/{id}/notifications` | GET | User | Project notification configs |
| `/api/v1/projects/{id}/credentials` | GET | User | List project credentials |
| `/api/v1/projects/{id}/credentials` | POST | User | Create a project credential |

Users can only see and act on their own projects.

```bash
# Trigger a build
curl -q -LSsf -X POST https://casci.casapps.us/api/v1/projects/1/builds \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"branch":"main"}'
```

## Builds

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/builds/{id}` | GET | User | Fetch a build |
| `/api/v1/builds/{id}/log` | GET | User | Build log |
| `/api/v1/builds/{id}/cancel` | POST | User | Cancel a running build |
| `/api/v1/builds/{id}/restart` | POST | User | Restart a build |
| `/api/v1/builds/{id}/security` | GET | User | Security reports for the build |
| `/api/v1/builds/{id}/security/scan` | POST | User | Trigger a security scan |
| `/api/v1/builds/{id}/notifications` | GET | User | Notification delivery log |

```bash
curl -q -LSsf https://casci.casapps.us/api/v1/builds/42/log \
  -H "Authorization: Bearer $TOKEN"
```

Live log output is available over WebSocket at `/ws`.

## Nodes

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/nodes` | GET | Admin | List build nodes |
| `/api/v1/nodes/register` | POST | Node token | Register a build node |
| `/api/v1/nodes/token` | POST | Admin | Issue a node registration token |
| `/api/v1/nodes/{id}` | GET | Admin | Fetch a node |
| `/api/v1/nodes/{id}` | PUT, PATCH | Admin | Update a node |
| `/api/v1/nodes/{id}` | DELETE | Admin | Remove a node |
| `/api/v1/nodes/{id}/heartbeat` | POST | Node | Node heartbeat |
| `/api/v1/nodes/{id}/drain` | POST | Admin | Drain a node before removal |

Node registration is unauthenticated by user credentials and secured by the registration token issued from `/api/v1/nodes/token`.

## Credentials

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/credentials/user` | GET | User | List the caller's credentials |
| `/api/v1/credentials/user` | POST | User | Create a user credential |
| `/api/v1/credentials/user/{id}` | GET | User | Fetch a user credential |
| `/api/v1/credentials/user/{id}` | PUT, PATCH | User | Update a user credential |
| `/api/v1/credentials/user/{id}` | DELETE | User | Delete a user credential |
| `/api/v1/credentials/project/{id}` | GET | User | Fetch a project credential |
| `/api/v1/credentials/project/{id}` | PUT, PATCH | User | Update a project credential |
| `/api/v1/credentials/project/{id}` | DELETE | User | Delete a project credential |

Secret values are encrypted at rest and are never returned in a response body.

## Notifications

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/notifications` | GET, POST | User | List or create notification configs |
| `/api/v1/notifications/{id}` | GET | User | Fetch a config |
| `/api/v1/notifications/{id}` | PUT, PATCH | User | Update a config |
| `/api/v1/notifications/{id}` | DELETE | User | Delete a config |
| `/api/v1/notifications/test` | POST | User | Send a test notification |
| `/api/v1/notifications/types` | GET | User | Supported provider types |
| `/api/v1/notifications/events` | GET | User | Supported event types |

## Security

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/security/reports` | GET | User | List security reports |
| `/api/v1/security/reports/{id}` | GET | User | Fetch a report |
| `/api/v1/security/statistics` | GET | User | Aggregate findings |
| `/api/v1/security/config` | GET, PUT | Admin | Scan policy configuration |

Scan types are `vulnerability`, `sast`, `secret`, `license`, `sbom`, and `malware`. Findings are classified `critical`, `high`, `medium`, `low`, or `info`.

## Audit

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/audit/events` | GET | User | Query audit events |
| `/api/v1/audit/status` | GET | Admin | Audit subsystem status |
| `/api/v1/audit/config` | PUT, PATCH | Admin | Update retention configuration |
| `/api/v1/audit/cleanup` | POST | Admin | Purge events past retention |

`/api/v1/audit/config` is write-only. A `GET` against it returns `405`; read the current state from `/api/v1/audit/status`.

## Compliance

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/compliance/modes` | GET | Admin | List available modes |
| `/api/v1/compliance/mode` | POST, PUT | Admin | Set the active mode |
| `/api/v1/compliance/preset` | GET | Admin | Settings a mode applies |
| `/api/v1/compliance/config` | GET, PUT, PATCH | Admin | Compliance configuration |
| `/api/v1/compliance/check` | POST | Admin | Run a compliance check |

Available modes are `none`, `hipaa`, `sox`, `pci-dss`, `gdpr`, `fedramp`, and `iso27001`.

`/api/v1/compliance/mode` is write-only. Read the currently active mode from `/api/v1/compliance/config`.

## Metrics

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/metrics` | GET | Admin | Prometheus-format metrics |
| `/metrics/json` | GET | Admin | Same metrics as JSON |
| `/api/v1/metrics/system` | GET | Admin | Host and runtime metrics |
| `/api/v1/metrics/builds` | GET | Admin | Build throughput and outcomes |
| `/api/v1/metrics/nodes` | GET | Admin | Per-node capacity and health |
| `/api/v1/metrics/security` | GET | Admin | Scan findings over time |
| `/api/v1/metrics/api` | GET | Admin | API request metrics |

## Webhooks

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/webhook` | POST | Signature | Inbound SCM webhook receiver |

One receiver handles GitHub, GitLab, Bitbucket, and Gitea. The provider is identified from the request headers, and the payload signature is verified against `server.security.webhook_secret` in `server.yml`.

!!! warning "Webhooks do not trigger builds yet"
    Signature verification works for GitHub and GitLab, but the receiver cannot yet map an incoming repository URL to a project — the lookup returns an error unconditionally, so a verified push responds `404 Project not found` instead of queueing a build. Bitbucket events are acknowledged with `200` and discarded. See [Integrations](integrations.md) for the per-provider detail.

## Jenkins Compatibility

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/json` | GET | Jenkins-style server description |
| `/crumbIssuer/api/json` | GET | Jenkins CSRF crumb issuer |

These exist so existing Jenkins clients and tooling keep working against CASCI without modification.

## Swagger and OpenAPI

| Endpoint | Description |
|----------|-------------|
| `/server/docs/swagger` | Swagger UI |
| `/api/v1/server/swagger` | OpenAPI specification, JSON |
| `/api/swagger` | Unversioned alias for the JSON specification |
| `/api/v1/server/swagger.yaml` | OpenAPI specification, YAML |
| `/api/swagger.yaml` | Unversioned alias for the YAML specification |

The generated OpenAPI document describes request and response schemas; this page describes the surface and its access rules.

## GraphQL

| Endpoint | Description |
|----------|-------------|
| `/api/v1/server/graphql` | GraphQL endpoint |
| `/api/graphql` | Unversioned alias for the endpoint |
| `/server/docs/graphql` | Interactive playground |

```bash
curl -q -LSsf -X POST https://casci.casapps.us/api/v1/server/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ projects { id name } }"}'
```

!!! warning "GraphQL is not usable yet"
    The endpoint and the playground are mounted and the schema is generated, but every resolver in `src/pkg/graphql/resolver.go` is still a `panic("not implemented")` stub. No query or mutation returns data. Use the REST API described above; this section documents the route layout only.

## WebSocket

| Endpoint | Description |
|----------|-------------|
| `/ws` | Live build log and status stream |

## Error Responses

Errors are returned as JSON with a message describing the failure.

| Status | Meaning |
|--------|---------|
| `400` | Malformed request or invalid path |
| `401` | Missing or invalid credential |
| `403` | Authenticated but not permitted, including admin-only endpoints |
| `404` | Resource does not exist, or is not visible to the caller |
| `405` | Method not allowed for this endpoint |
| `409` | Conflict, such as a duplicate name |
| `422` | Request understood but semantically invalid |
| `429` | Rate limited |
| `500` | Server error |

A resource belonging to another user returns `404` rather than `403`, so the API does not confirm that it exists.
