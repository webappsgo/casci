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

Regular users can create named `usr_` API tokens from `/users/tokens` or
`POST /api/v1/users/tokens`. Token values are shown once; the database stores
the SHA-256 digest and display prefix. Users can list, inspect, and revoke only
their own tokens. The supported scopes are `global`, `read-write`, and `read`;
read-only tokens cannot make state-changing requests, and read-write tokens
cannot delete resources or use server administration routes. The deprecated
`X-API-Token` header is ignored. MFA, OAuth, OIDC, and LDAP integration are not
implemented in the current user/auth system.

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
to a browser session identifier and expires them. CSRF validation applies only
to frontend form actions that use browser cookies; `/api/**` routes and
compatibility routes never require CSRF tokens. Bearer-authenticated frontend
actions also skip the check because browsers do not attach Bearer credentials
automatically. API login and registration, node registration, and signed
webhooks are outside the frontend CSRF scope and use their own entry conditions.
Administrator and user session cookies are `HttpOnly`, `Secure`, and
`SameSite=Strict`; CSRF tokens are bound to the corresponding signed session
cookie and cannot be reused between user and administrator sessions.

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

Infrastructure connection credentials are encrypted with the configured
credential encryptor and scoped to the owning user's key context. API and HTML
responses expose only whether credentials are configured; replacement
credentials are write-only. Budget limits are user scoped. Provider spend is
not synchronized and is available only if a spend value has already been
recorded locally.

Do not put secrets in pipeline files, documentation, source control, or build
logs. Secret injection, cleanup, and redaction are not yet complete end to end.

## Build execution

Builds use a container executor and a per-build workspace. The executor does
not currently provide the complete isolated VM/container boundary, network
policy, clean host guarantees, secret handling, per-user quotas, fair
scheduling, or signed artifact chain required by the product design. Pipeline
commands are repository-controlled code. Run CASCI only on a host where that
execution risk is acceptable.

Build security scanning is enabled by default for vulnerability, SAST, secret,
license, and SBOM checks. Scanners run before the executor reports a build as
successful. A missing scanner executable, malformed scanner output, or report
storage error fails the build closed. The scanner adapters currently expect
`trivy`, `semgrep`, `gitleaks`, `go-licenses`, and `syft` in the server
process's `PATH`; the minimal production image does not currently bundle all
of these tools. Provisioning the complete scanner toolchain in every runtime
is still required for production use. Secret matches are not persisted in the
scanner result, and Gitleaks report output is redacted. Manual scans require
the caller to own the build's project and an existing build workspace; scan
requests cannot override the administrator's scan policy. Queue saturation is
reported as unavailable instead of silently dropping a scan.

## MCP

The MCP endpoint requires a bearer JWT on every request, ignores browser
cookies, accepts POST only, checks the request Origin, and enforces a 1 MiB
request-body limit. It currently exposes only a subset of API operations. MCP
requests do not yet pass through a complete shared API limiter/audit pipeline,
and named token scopes are not implemented. See
[MCP help](https://casci.casapps.us/server/help/mcp).

## Webhooks

The webhook route is `POST /webhook`; it caps request bodies at 1 MiB. GitHub
uses `X-Hub-Signature-256`, GitLab uses `X-GitLab-Token`, Bitbucket uses
`X-Hub-Signature` with a SHA-256 HMAC, and Gitea uses `X-Gitea-Signature` with
a raw SHA-256 HMAC digest. Every provider fails closed if the shared
`server.security.webhook_secret` is unset or its credential is invalid. Keep
that secret private and configure the provider with the same value. A
`202 Accepted` response means a build was queued; it does not mean execution
completed.

## Public endpoints and data

The current public machine-readable endpoints include:

- `GET /.well-known/security.txt`
- `GET /.well-known/llms.txt` and `GET /llms.txt`
- `GET /api/autodiscover` (filters non-public hosts and publishes only binaries with matching checksums)
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
