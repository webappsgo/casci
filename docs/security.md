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
(7-day default). API login returns opaque, revocable `adm_` or `usr_` bearer
API tokens and sets no cookie. Administrator API routes use hashed `adm_` tokens; first-run setup displays the
initial token once. Administrator accounts use the separate `admins` table
and their own session issuer. The signing secret is
loaded from configuration and must be at least 32 bytes. Administrator login
locks the account for 15 minutes after five failed password attempts.
Both browser and API login also return HTTP 429 after five failed attempts
from the same direct client address within 15 minutes. The handlers do not
trust `X-Forwarded-For` without an explicit trusted-proxy policy.

Server administrators can enroll TOTP or multiple WebAuthn passkeys from the
profile security page. Passkey registration requires the current password;
registration and login ceremonies use expiring, single-use server-side
challenges. First-factor passkey login and passkey second-factor login both
issue the normal administrator session. Enabling the first security factor
requires acknowledging one-time recovery codes before it becomes active.
Recovery codes are stored as hashes and can be consumed only once. Regular users
can enroll TOTP and passkeys from `/users/security`; registration and login use
single-use challenges, and recovery-code acknowledgment is required before a
factor becomes active. User login completes a pending TOTP challenge before
issuing the session token. `/users/security/sessions` lists and revokes the
caller's API-token sessions. Native API tokens are not JWTs.

Password recovery is available through `/server/auth/password/forgot` and
`/server/auth/password/reset`, with matching API endpoints. Recovery tokens
are one-time, time-limited, and stored only as SHA-256 digests. The flow
supports administrators and regular users. When SMTP works and email
verification is required, new regular users must verify the current account
email before login; verification codes are bound to that address.

Regular users can create named `usr_` API tokens from `/users/tokens` or
`POST /api/v1/users/tokens`. Token values are shown once; the database stores
the SHA-256 digest and display prefix. Users can list, inspect, and revoke only
their own tokens. The supported scopes are `global`, `read-write`, and `read`;
read-only tokens cannot make state-changing requests, and read-write tokens
cannot delete resources or use server administration routes. The deprecated
`X-API-Token` header is ignored. The server does not currently expose OIDC,
LDAP, or SAML sign-in flows; these are required by AI.md and remain an
implementation gap until the provider routes are present.

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

Builds run in per-build Docker containers with no network, a read-only root filesystem, dropped Linux capabilities, no-new-privileges, a non-root UID/GID, PID and memory limits, and a private temporary filesystem. The build workspace is mounted for pipeline output. The queue resolves project owners, bounds global and per-user pending builds, allows one running build per user, and dispatches owners round-robin. It does not provide per-user CPU or memory budgets, a VM boundary, or a signed artifact chain. Pipeline commands are repository-controlled code and can alter mounted workspace contents; isolate the host and review pipeline access accordingly.

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

The MCP endpoint requires a bearer API token on every request, ignores browser
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
- `GET /server/security`, `/server/security/policy`, and `/server/security/thanks`
- `GET /api/v1/server/security`, `/api/v1/server/security/policy`, and `/api/v1/server/security/thanks`
- `GET /.well-known/llms.txt` and `GET /llms.txt`
- `GET /api/autodiscover` (filters non-public hosts and publishes only binaries with matching checksums)
- `POST /api/{version}/server/reports/{type}` (anonymous browser telemetry, bounded, sanitized, and rate-limited)
- `GET /server/healthz`, `GET /readyz`, and `GET /livez`
- Public sign-in and registration routes under `/api/{version}/server/auth/`
- `POST /api/{version}/server/mcp`, which still requires bearer auth

The security contact link in `security.txt` carries a rotating 48-hour
`security_id`; the form accepts the current and previous window and requires a
review confirmation before storage. A review token binds the submitted fields,
must be at least two seconds old, expires after 15 minutes, and can be consumed
once per server process. The form's arithmetic CAPTCHA answer is stored as a
hash, consumed on the first verification attempt, and expires after ten
minutes. Each security report body and its
researcher contact details are encrypted separately with AES-256-GCM before
they enter the database. Server administrators can
decrypt reports from the authenticated security reports page, and every report
view is audited without recording report content. The researcher status page
is available at `/server/security/report/{tracking_id}` and its token is
single-use per UTC day. When SMTP is configured, the server can send a researcher acknowledgment with a status-only tracking link and can send status-change messages without including report contents. The acknowledgment can be PGP-encrypted when a valid researcher public key is configured. Report-at-rest encryption currently uses AES-256-GCM; maintainer notification with per-admin PGP encryption and the AI.md-defined PGP-at-rest policy remain incomplete.

## Reporting a vulnerability

Use the repository's [GitHub private vulnerability reporting page](https://github.com/webappsgo/casci/security/advisories/new)
for a source-code vulnerability. For a deployed instance, check that
instance's `/server/security` page for its operator's reporting channel and
`/server/security/policy` for the coordinated-disclosure terms. For instance
reports, use the security contact form linked from that page. Do not include
live credentials or user data in a report submitted over an untrusted channel.
