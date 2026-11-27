# API reference

This page documents the HTTP routes currently registered by the server. The
route set is still changing, and the OpenAPI document is incomplete; verify a
route against the running build before depending on it.

## Base path and authentication

The default API prefix is `https://casci.casapps.us/api/v1`. The `v1` segment
comes from `server.api_version`. Some machine routes have unversioned aliases.

## Frontend response formats

Frontend pages normally return server-rendered HTML. Requests from recognized
text browsers (Lynx, w3m, Links, ELinks, Browsh, Carbonyl, and NetSurf) receive
the same server-rendered content with script elements removed; navigation and
forms remain usable through links and standard form submissions. Requests from
recognized command-line HTTP tools, or requests that explicitly accept
`text/plain`, receive an 80-column text rendering of the page. API paths and
static assets keep their own response formats.

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
| GET | `/api/v1/server/security` | Public | Security contact channels, policy and acknowledgment URLs, and PGP availability |
| GET | `/api/v1/server/security/policy` | Public | Built-in coordinated disclosure policy in Markdown |
| GET | `/api/v1/server/security/thanks` | Public | Credited researcher acknowledgments |
| GET, POST | `/api/v1/server/preferences` | Public | Preference API handler |
| GET, POST | `/api/v1/server/preferences/sync` | Public | Preference sync API handler |
| GET, POST | `/api/v1/server/preferences/sync/import` | Public | Preference import API handler |
| POST | `/api/v1/server/auth/register` | Public when multi-user mode is enabled and registration is open | Register a regular-user account |
| POST | `/api/v1/server/auth/login` | Public | Authenticate a server administrator or enabled regular user; sets the matching cookie and returns a bearer JWT for regular users |
| POST | `/api/v1/server/auth/refresh` | Authenticated | Refresh a JWT |
| GET | `/api/v1/server/{admin_path}/setup` | Public during first run | Report whether primary administrator setup is required |
| POST | `/api/v1/server/{admin_path}/setup/verify` | Setup token | Verify the one-time first-run setup token |
| POST | `/api/v1/server/{admin_path}/setup/account` | Setup token | Create the primary administrator and return the one-time `adm_` API token |
| POST | `/api/v1/server/{admin_path}/setup/complete` | Public | Report whether primary administrator creation is complete |
| GET | `/api/v1/server/{admin_path}/{admin_username}/profile` | Administrator bearer token | Read the authenticated administrator profile |
| PATCH | `/api/v1/server/{admin_path}/{admin_username}/profile` | Administrator bearer token | Update the account email after current-password verification |
| POST | `/api/v1/server/{admin_path}/{admin_username}/profile/password` | Administrator bearer token | Change the administrator password after current-password verification |
| GET | `/api/v1/server/{admin_path}/{admin_username}/profile/security` | Administrator session | Read TOTP, passkey, and recovery-code status |
| POST | `/api/v1/server/{admin_path}/{admin_username}/profile/security/totp/setup` | Administrator session | Start TOTP enrollment after current-password verification; return the secret and provisioning URI |
| POST | `/api/v1/server/{admin_path}/{admin_username}/profile/security/totp/enable` | Administrator session | Verify TOTP and return one-time recovery codes; a separate acknowledgment enables TOTP |
| POST | `/server/{admin_path}/{admin_username}/profile/security` | Administrator session + CSRF | Disable TOTP after verifying a current TOTP or recovery code |
| POST | `/api/v1/server/{admin_path}/{admin_username}/profile/security/passkey/setup` | Administrator session | Verify the current password and begin WebAuthn registration |
| POST | `/api/v1/server/{admin_path}/{admin_username}/profile/security/passkey/register` | Administrator session | Finish WebAuthn registration; first enrollment returns one-time recovery codes pending acknowledgment; the security page also posts rename/revoke actions here |
| POST | `/api/v1/server/auth/passkey/challenge` | Public | Request a primary-login or partial-session passkey challenge |
| POST | `/api/v1/server/auth/passkey/verify` | Public | Verify a passkey and issue an administrator session |
| POST | `/api/v1/server/auth/2fa` | Partial administrator session | Complete login with a TOTP code |
| POST | `/api/v1/server/auth/recovery/use` | Partial administrator session | Complete login with a one-time recovery code |

The HTML login page is `/server/auth/login`; administrator accounts can use a
passkey as a primary credential. Accounts requiring a second factor continue
at `/server/auth/2fa`, which accepts TOTP, a passkey, or a one-time recovery
code. Passkey ceremonies use short-lived, one-use server-side challenges. The
`/healthz` root alias is
disabled unless `server.healthz.root.enabled` is true.

Administrator API credentials can be managed from
`/server/{admin_path}/security/tokens`. The page creates named credentials,
shows each new value once, lists safe metadata, and revokes credentials with a
CSRF-protected form. The API supports listing, revocation, and one-time
rotation:

| Method | Path | Access | Behavior |
|--------|------|--------|----------|
| GET | `/api/{api_version}/server/{admin_path}/security/tokens` | Administrator bearer token | List safe metadata for active administrator, user, and organization tokens; secret values are never returned |
| DELETE | `/api/{api_version}/server/{admin_path}/security/tokens/{id}` | Administrator bearer token | Revoke a non-agent token across account types |
| POST | `/api/{api_version}/server/{admin_path}/security/tokens/{id}/rotate` | Administrator bearer token | Replace a token secret, preserving owner, name, scope, and expiry; returns the new value once |
| GET | `/api/{api_version}/server/{admin_path}/{admin_username}/profile/token` | Administrator bearer token | Read masked metadata for the setup-issued token belonging to the signed-in administrator |
| POST | `/api/{api_version}/server/{admin_path}/{admin_username}/profile/token` | Administrator bearer token | Regenerate the setup-issued token; returns its new value once |

## Users

| Method | Path | Access |
|--------|------|--------|
| GET, PATCH | `/api/v1/users` | Authenticated user; reads or updates the caller's own profile |
| GET, POST | `/api/v1/users/tokens` | Authenticated user; list token metadata or create a named `usr_` token (the value is returned once) |
| GET, DELETE | `/api/v1/users/tokens/{token_id}` | Authenticated owner; inspect metadata or immediately revoke a token |
| GET | `/api/v1/server/administration/users` | Administrator; lists accounts |
| GET, PUT, PATCH, DELETE | `/api/v1/server/administration/users/{id}` | Administrator; manages one account |

User management routes are under the configured admin path in a running
instance. Regular users cannot use the admin API path or read another user's
profile through `/api/v1/users`.

The server-rendered profile page is `/users`; it uses the same self-scoped
account data and accepts a no-JavaScript form POST for email or password
updates. `/users/tokens` and, when regular-user accounts are enabled,
`/server/preferences/tokens` list the caller's tokens and support no-JavaScript
create and revoke forms. Token values are returned only at creation; the
database stores a SHA-256 digest and the first eight characters for display.

## Projects and builds

| Method | Path | Access | Handler behavior |
|--------|------|--------|------------------|
| GET | `/api/v1/projects` | Authenticated | List projects |
| POST | `/api/v1/projects` | Authenticated | Create project |
| GET, PATCH, DELETE | `/api/v1/projects/{id}` | Authenticated owner | Read, update, or delete one owned project |
| GET | `/api/v1/projects/{id}/pipelines` | Authenticated owner | List configured pipelines or detect workflows from an HTTPS repository |
| GET | `/api/v1/projects/{id}/pipelines/{pipeline_id}` | Authenticated owner | Read one configured pipeline or detected workflow definition |
| GET, POST | `/api/v1/projects/{id}/builds` | Authenticated | List or trigger builds |
| GET | `/api/v1/projects/{id}/builds/stats` | Authenticated | Project build statistics |
| GET | `/api/v1/projects/{id}/notifications` | Authenticated | Project notification configurations |
| GET, POST | `/api/v1/projects/{id}/credentials` | Authenticated | List or create project credentials |
| GET | `/api/v1/builds/{id}` | Authenticated | Read a build |
| GET | `/api/v1/builds/{id}/logs` | Authenticated | Read a build log |
| POST | `/api/v1/builds/{id}/cancel` | Authenticated | Cancel a build |
| GET | `/api/v1/builds/{id}/artifacts` | Authenticated owner | List safe metadata for the build's artifacts |
| GET | `/api/v1/builds/{id}/artifacts/{artifact_id}` | Authenticated owner | Download an owned artifact with attachment and `nosniff` headers |
| GET | `/api/v1/builds/{id}/security` | Authenticated | Read build security reports |
| POST | `/api/v1/builds/{id}/security/scan` | Authenticated | Trigger a scan |
| GET | `/api/v1/builds/{id}/notifications` | Authenticated | Read build notification data |

Authorization and ownership checks are in the handlers/services and are not
fully represented by the routing table. Review the current source before
exposing these operations to users.

## Infrastructure connections and budgets

These endpoints require an authenticated regular-user account and operate
only on that account's records. Successful JSON responses use the standard
`ok`/`data` envelope. Connection secrets are accepted during create or update,
encrypted at rest, and never included in responses. Budget enforcement is
best effort; provider spend is returned only when a value has been recorded in
CASCI, because provider billing synchronization is not currently available.

| Method | Path | Handler behavior |
|--------|------|------------------|
| GET, POST | `/api/v1/infrastructure/connections` | List or create cloud, private-server, or storage connection metadata |
| GET, PATCH, DELETE | `/api/v1/infrastructure/connections/{connection_id}` | Read, update, or delete one caller-owned connection |
| GET, PUT | `/api/v1/budgets` | Read or replace the caller's monthly and per-provider budget limits and available spend data |
| POST | `/api/v1/infrastructure/hourly-instances/destroy-all` | Returns `503 MAINTENANCE` while no provider cleanup backend is configured |

The corresponding no-JavaScript pages are `/users/infrastructure` and
`/users/budgets`. Their state-changing forms use CSRF tokens.

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
| GET | `/api/v1/server/{admin_path}/logs` | Administrator Bearer token; list log files |
| GET | `/api/v1/server/{admin_path}/logs/{type}` | Administrator Bearer token; read entries with `limit`, `offset`, and `search` |
| GET | `/api/v1/server/{admin_path}/logs/{type}/download` | Administrator Bearer token; download a log file |
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
| POST | `/api/v1/webhooks/{provider}` | Provider-signature verified webhook; path provider must match provider headers |
| POST | `/webhook` | Legacy provider webhook handler |
| GET | `/ws` | WebSocket endpoint |

OpenAPI uses JSON only. There is no YAML endpoint.

## Error and response formats

Response envelopes vary by package. Some handlers return `ok`, `error`, and
`message`; others use package-specific JSON or `http.Error`. Do not assume a
single error schema across the API. Validation, pagination, compatibility,
and generated-schema behavior are still being standardized.
