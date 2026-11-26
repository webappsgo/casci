# API reference

This page documents the HTTP routes currently registered by the server. The
route set is still changing, and the OpenAPI document is incomplete; verify a
route against the running build before depending on it.

## Base path and authentication

The default API prefix is `https://casci.casapps.us/api/v1`. The `v1` segment
comes from `server.api_version`. Some machine routes have unversioned aliases.

Authenticated handlers accept a bearer JWT in the `Authorization` header:

```http
Authorization: Bearer <bearer-token>
```

Browser middleware also accepts the CASCI authentication cookie for routes
that use the shared user middleware. MCP accepts only the bearer header.
Login is public. Registration is mounted only when `server.users.enabled` is
`true` and `server.users.registration.mode` is `open`; otherwise the route
returns 404. User, admin, and node handlers apply different checks. The API
does not yet implement the complete token types,
scopes, rate limits, request IDs, or response-envelope rules from AI.md.

## Health, help, preferences, and auth

| Method | Path | Access | Behavior |
|--------|------|--------|----------|
| GET | `/api/v1/server/healthz` | Public | JSON health response |
| GET | `/api/healthz` | Public alias | Same health handler |
| GET | `/readyz` | Public | Readiness handler |
| GET | `/livez` | Public | Liveness handler |
| GET | `/api/v1/server/help` | Public | Help JSON and registered help pages |
| GET | `/api/v1/server/help/mcp` | Public | MCP connection details, tools, and project text |
| GET, POST | `/api/v1/server/preferences` | Public | Preference API handler |
| GET, POST | `/api/v1/server/preferences/sync` | Public | Preference sync API handler |
| GET, POST | `/api/v1/server/preferences/sync/import` | Public | Preference import API handler |
| POST | `/api/v1/server/auth/register` | Public when multi-user mode is enabled and registration is open | Register a regular-user account |
| POST | `/api/v1/server/auth/login` | Public | Authenticate a server administrator or enabled regular user; sets the matching cookie and returns a bearer JWT for regular users |
| POST | `/api/v1/server/auth/refresh` | Authenticated | Refresh a JWT |
| GET | `/api/v1/server/{admin_path}/setup` | Public during first run | Report whether primary administrator setup is required |
| POST | `/api/v1/server/{admin_path}/setup/verify` | Setup token | Verify the one-time first-run setup token |
| POST | `/api/v1/server/{admin_path}/setup/account` | Setup token + CSRF for browser forms | Create the primary administrator and return the one-time `adm_` API token |
| POST | `/api/v1/server/{admin_path}/setup/complete` | Public | Report whether primary administrator creation is complete |

The HTML login page is `/server/auth/login`. The `/healthz` root alias is
disabled unless `server.healthz.root.enabled` is true.

## Users

| Method | Path | Access |
|--------|------|--------|
| GET, PATCH | `/api/v1/users` | Authenticated user; reads or updates the caller's own profile |
| GET | `/api/v1/users/me` | Authenticated user; compatibility alias for the caller profile |
| POST | `/api/v1/users/me/token` | Authenticated user; regenerates the legacy API token |
| GET | `/api/v1/server/administration/users` | Administrator; lists accounts |
| GET, PUT, PATCH, DELETE | `/api/v1/server/administration/users/{id}` | Administrator; manages one account |

User management routes are under the configured admin path in a running
instance. Regular users cannot use the admin API path or read another user's
profile through `/api/v1/users`.

The server-rendered profile page is `/users`; it uses the same self-scoped
account data and accepts a no-JavaScript form POST for email or password
updates.

## Projects and builds

| Method | Path | Access | Handler behavior |
|--------|------|--------|------------------|
| GET | `/api/v1/projects` | Authenticated | List projects |
| POST | `/api/v1/projects/` | Authenticated | Create project through the trailing-slash subroute |
| GET, PUT, PATCH, DELETE | `/api/v1/projects/{id}` | Authenticated | Project CRUD |
| GET, POST | `/api/v1/projects/{id}/builds` | Authenticated | List or trigger builds |
| GET | `/api/v1/projects/{id}/builds/stats` | Authenticated | Project build statistics |
| GET | `/api/v1/projects/{id}/notifications` | Authenticated | Project notification configurations |
| GET, POST | `/api/v1/projects/{id}/credentials` | Authenticated | List or create project credentials |
| GET | `/api/v1/builds/{id}` | Authenticated | Read a build |
| GET | `/api/v1/builds/{id}/log` | Authenticated | Read a build log |
| POST | `/api/v1/builds/{id}/cancel` | Authenticated | Cancel a build |
| POST | `/api/v1/builds/{id}/restart` | Authenticated | Restart a build |
| GET | `/api/v1/builds/{id}/security` | Authenticated | Read build security reports |
| POST | `/api/v1/builds/{id}/security/scan` | Authenticated | Trigger a scan |
| GET | `/api/v1/builds/{id}/notifications` | Authenticated | Read build notification data |

Authorization and ownership checks are in the handlers/services and are not
fully represented by the routing table. Review the current source before
exposing these operations to users.

## Nodes

| Method | Path | Access |
|--------|------|--------|
| GET | `/api/v1/nodes` | Administrator |
| POST | `/api/v1/nodes/token` | Administrator |
| POST | `/api/v1/nodes/register` | Registration token |
| GET, PUT, PATCH, DELETE | `/api/v1/nodes/{id}` | Administrator |
| POST | `/api/v1/nodes/{id}/heartbeat` | Node handler |
| POST | `/api/v1/nodes/{id}/drain` | Administrator |

## Agents

Agent management is available to administrators under the configured admin
path and to signed-in users under `/api/v1/users/agents`. Management requests
require the matching account credential; user routes require a user bearer JWT.
Machine registration, heartbeat, report, and binary download requests require
an agent-scoped `adm_` or `usr_` bearer token. The full route behavior and
payload examples are in the [Agent guide](agents.md).

| Method | Path | Access |
|--------|------|--------|
| GET, POST | `/api/v1/server/{admin_path}/agents` | Administrator |
| GET, PATCH, DELETE | `/api/v1/server/{admin_path}/agents/{id}` | Administrator |
| POST | `/api/v1/server/{admin_path}/agents/{id}/token` | Administrator |
| GET | `/api/v1/server/{admin_path}/agents/{id}/metrics` | Administrator |
| POST | `/api/v1/server/{admin_path}/agents/{register,heartbeat,report}` | Admin-scoped agent token |
| GET, POST | `/api/v1/users/agents` | Authenticated user |
| GET, PATCH, DELETE | `/api/v1/users/agents/{id}` | Authenticated user; owner scoped |
| POST | `/api/v1/users/agents/{id}/token` | Authenticated user; owner scoped |
| GET | `/api/v1/users/agents/{id}/metrics` | Authenticated user; owner scoped |
| POST | `/api/v1/users/agents/{register,heartbeat,report}` | User-scoped agent token |
| GET, HEAD | `/agent/binaries/{name}` | Any valid agent-scoped bearer token |

Agent management API paths return a JSON `ok`/`data` envelope. Agent tokens
are stored as SHA-256 hashes and are shown only when created or rotated.
Organization-scoped agent routes are not currently implemented.

## Credentials and notifications

| Method | Path | Access |
|--------|------|--------|
| GET, POST | `/api/v1/credentials/user` | Authenticated |
| GET, PUT, PATCH, DELETE | `/api/v1/credentials/user/{id}` | Authenticated |
| GET, PUT, PATCH, DELETE | `/api/v1/credentials/project/{id}` | Authenticated |
| GET, POST | `/api/v1/notifications` | Authenticated |
| GET, PUT, PATCH, DELETE | `/api/v1/notifications/{id}` | Authenticated |
| POST | `/api/v1/notifications/test` | Authenticated |
| GET | `/api/v1/notifications/types` | Authenticated |
| GET | `/api/v1/notifications/events` | Authenticated |

The credential endpoints are sensitive. Do not expose their responses to
untrusted clients or model contexts.

## Security, audit, compliance, and metrics

| Method | Path | Access |
|--------|------|--------|
| GET | `/api/v1/security/reports` | Authenticated |
| GET | `/api/v1/security/reports/{id}` | Authenticated |
| GET | `/api/v1/security/statistics` | Authenticated |
| GET, PUT | `/api/v1/security/config` | Administrator |
| GET | `/api/v1/audit/events` | Authenticated |
| GET | `/api/v1/audit/status` | Administrator |
| PUT, PATCH | `/api/v1/audit/config` | Administrator |
| POST | `/api/v1/audit/cleanup` | Administrator |
| GET, PUT, PATCH | `/api/v1/compliance/config` | Administrator |
| GET | `/api/v1/compliance/modes` | Administrator |
| GET | `/api/v1/compliance/preset` | Administrator |
| POST, PUT | `/api/v1/compliance/mode` | Administrator |
| POST | `/api/v1/compliance/check` | Administrator |
| GET | `/metrics`, `/metrics/json` | Administrator |
| GET | `/api/v1/metrics/{system,builds,nodes,security,api}` | Administrator |

## MCP

The canonical endpoint is `POST /api/v1/server/mcp`; direct aliases are
`POST /api/mcp` and `POST /mcp`. Requests require a bearer token and the MCP
protocol headers. This implementation currently exposes a subset of the API
operation list as tools. See [MCP help](https://casci.casapps.us/server/help/mcp)
for the current tool list and request format.

## Documentation and compatibility routes

| Method | Path | Behavior |
|--------|------|----------|
| GET | `/api/v1/server/swagger` | OpenAPI JSON (generated route inventory is incomplete) |
| GET | `/api/swagger` | Unversioned OpenAPI JSON alias |
| GET | `/server/docs/swagger` | Swagger UI |
| GET, POST | `/api/v1/server/graphql` | GraphQL handler; resolver coverage is incomplete |
| GET, POST | `/api/graphql` | Unversioned GraphQL alias |
| GET | `/server/docs/graphql` | GraphQL playground |
| GET | `/api/json` | Jenkins-shaped server response |
| GET | `/crumbIssuer/api/json` | Jenkins-shaped crumb response |
| POST | `/webhook` | Provider webhook handler |
| GET | `/ws` | WebSocket endpoint |

OpenAPI uses JSON only. There is no YAML endpoint.

## Error and response formats

Response envelopes vary by package. Some handlers return `ok`, `error`, and
`message`; others use package-specific JSON or `http.Error`. Do not assume a
single error schema across the API. Validation, pagination, compatibility,
and generated-schema behavior are still being standardized.
