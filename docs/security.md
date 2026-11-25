# Security

CASCI is under active development. This page describes current controls and
known gaps; it is not a security certification. Do not run untrusted pipelines
on a shared or sensitive host until build isolation and tenant boundaries meet
your requirements.

## Authentication

Regular-user passwords and local administrator passwords are hashed with
Argon2id. The shared login endpoint checks administrators first and then,
when enabled, regular users. It sets an HTTP-only `admin_session` cookie for
administrators (30-day default) or `user_session` cookie for regular users
(7-day default). Regular-user API login returns a signed JWT. Administrator
API routes use hashed `adm_` bearer tokens; first-run setup displays the
initial token once. Administrator accounts use the separate `admins` table
and their own session issuer. The signing secret is
loaded from configuration and must be at least 32 bytes. Administrator login
locks the account for 15 minutes after five failed password attempts.
Both browser and API login also return HTTP 429 after five failed attempts
from the same direct client address within 15 minutes. The handlers do not
trust `X-Forwarded-For` without an explicit trusted-proxy policy.

The middleware also supports a legacy per-user API token through
`X-API-Token`; the database stores its SHA-256 digest. This token is distinct
from the bearer JWT path used by MCP. There are no named token records, granular
scopes, service-account tokens, MFA, OAuth, OIDC, or LDAP integration in the
current user/auth implementation.

## Authorization and tenant boundaries

Legacy user records with an administrator flag are migrated into the separate
administrator table, and regular-user registration never grants administrator
access. Administrator sessions are checked against the administrator record;
regular-user sessions do not authorize admin pages. Project and build handlers
are intended to scope records to their owner. The full multi-user design
requires stronger project, credential, workspace, artifact, queue, quota, and
encryption boundaries than are currently implemented. Treat user isolation as
incomplete.

## Browser protections

Server-rendered forms include CSRF tokens. The CSRF manager binds issued tokens
to a browser session identifier and expires them. Login, registration, node
registration, and the signed webhook route are exempt from the general CSRF
check because they use other entry conditions or create the session.

CORS is configured with `web.cors`, which accepts a single origin, a
comma-separated list, or `*`. The generated first-run default is `*`; change
it to trusted origins before exposing the service. Credentialed CORS is only
returned for an explicit matching origin. The current CORS allowed-header list
does not include all MCP protocol headers, so browser-based MCP clients may
fail preflight even when their origin is otherwise accepted.

The server includes request security headers and a Content Security Policy.
Review the live middleware before relying on a policy value as a deployment
guarantee.

## Encryption and secrets

Credential values use AES-256-GCM with a PBKDF2-HMAC-SHA256 derived key. The
master key is held in `server.security.encryption_key` in `server.yml`. Protect
that file and keep a secure backup; losing the key makes encrypted credentials
unreadable. The current key rotation and per-user key model do not meet all
requirements in IDEA.md.

Do not put secrets in pipeline files, documentation, source control, or build
logs. Secret injection, cleanup, and redaction are not yet complete end to end.

## Build execution

Builds use a container executor and a per-build workspace. The executor does
not currently provide the complete isolated VM/container boundary, network
policy, clean host guarantees, secret handling, per-user quotas, fair
scheduling, or signed artifact chain required by the product design. Pipeline
commands are repository-controlled code. Run CASCI only on a host where that
execution risk is acceptable.

## MCP

The MCP endpoint requires a bearer JWT on every request, ignores browser
cookies, accepts POST only, checks the request Origin, and enforces a 1 MiB
request-body limit. It currently exposes only a subset of API operations. MCP
requests do not yet pass through a complete shared API limiter/audit pipeline,
and named token scopes are not implemented. See
[MCP help](https://casci.casapps.us/server/help/mcp).

## Webhooks

The webhook route is `/webhook`. GitHub uses HMAC signature verification and
GitLab uses a shared token. Provider support and build triggering are
incomplete; a successful response is not proof that a build was queued. Keep
the webhook secret private and configure the provider to send its signature.

## Public endpoints and data

The current public machine-readable endpoints include:

- `GET /.well-known/security.txt`
- `GET /.well-known/llms.txt` and `GET /llms.txt`
- `GET /server/healthz`, `GET /readyz`, and `GET /livez`
- Public sign-in and registration routes under `/api/{version}/server/auth/`
- `POST /api/{version}/server/mcp`, which still requires bearer auth

The contact form does not currently deliver reports. It is not a substitute for
the private vulnerability reporting URL published by `security.txt`.

## Reporting a vulnerability

Use the repository's [GitHub private vulnerability reporting page](https://github.com/webappsgo/casci/security/advisories/new)
for a source-code vulnerability. For a deployed instance, check that
instance's `/server/security` page for its operator's reporting channel. Do not
include live credentials or user data in a report submitted over an untrusted
channel.
