# Security

## Authentication & Identity

CASCI ships local authentication only. No external identity provider is enabled in this build: no OIDC, no LDAP, no SAML. If that changes, this page and [Integrations](integrations.md) are where it will be documented.

### Credentials

| Credential | Used by | Transport |
|------------|---------|-----------|
| Session cookie | Browser sessions, including the admin panel | `Cookie` header, CSRF-protected |
| JWT | API clients after `POST /api/v1/auth/login` | `Authorization: Bearer` |
| API token | Long-lived automation | `Authorization: Bearer` |
| Node registration token | Build nodes joining the cluster | `POST /api/v1/nodes/register` |

Bearer tokens are the only accepted API transport. Tokens are not read from query parameters, because a URL-borne credential ends up in proxy logs, browser history, and referrer headers.

### Password Storage

Passwords are hashed with Argon2id and stored in the standard encoded form carrying version, memory cost, time cost, parallelism, salt, and digest:

```
$argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
```

The hash is never serialized into an API response.

### JWT

Access and refresh tokens are signed with HMAC-SHA256. Validation rejects any token whose signing method is not HMAC, which closes the algorithm-confusion path where an attacker swaps the algorithm header. Refresh tokens are valid for 24 hours.

## Authorization

The permission model is a single boolean on the user record, `is_admin`. There is no role table, no RBAC, and no service-account type.

- Administrators manage the server: users, nodes, security policy, audit retention, compliance mode, and metrics.
- Users see only their own projects, builds, workspaces, credentials, and logs.
- A request for another user's resource returns `404`, not `403`, so the API does not confirm that the resource exists.

The administrator flag governs which endpoints are reachable, not which rows are visible. It is not intended as a superuser over other people's project data, and admin-only routes deliberately exclude the per-user resource endpoints.

Admin-only endpoints run an admin check after authentication and return `403` for authenticated non-admins.

## CSRF Protection

Browser-driven state changes use a double-submit CSRF token bound to the session and held in the `csrf_token` cookie. Tokens expire and are swept periodically.

Four endpoints are exempt because they cannot present a session-bound token. Each is authenticated by another mechanism:

| Path | Authenticated by |
|------|------------------|
| `/webhook` | Provider signature verified against `server.security.webhook_secret` |
| `/api/v1/auth/login` | Establishes the session, so no session exists yet |
| `/api/v1/auth/register` | Public registration |
| `/api/v1/nodes/register` | Node registration token |

## CORS

The allowlist is `web.cors` in `server.yml` — a single origin, a comma-separated list of origins, or `*`.

!!! danger "The generated default is `*`"
    On first run `web.cors` is written as `*`. In that state CASCI answers every request with `Access-Control-Allow-Origin: *`, so any website can issue cross-origin requests to your instance. `Access-Control-Allow-Credentials` is **not** sent with `*`, which stops the browser from attaching cookies — but a bearer token held by a page-embedded script is still usable against your API.

    Set `web.cors` to your own origins before exposing the instance. Only an explicit, matched origin gets `Access-Control-Allow-Credentials: true`.

Behavior by value:

| `web.cors` | Response |
|------------|----------|
| Empty | No `Access-Control-Allow-Origin` header at all; browsers refuse cross-origin reads |
| `*` | `Access-Control-Allow-Origin: *`, no credentials |
| Comma-separated list | The request `Origin` is echoed back when it matches, with `Access-Control-Allow-Credentials: true` |

`Vary: Origin` is always sent so caches key on the origin. The `DOMAIN` environment variable sets the public hostname; it does **not** feed the CORS allowlist.

Allowed request headers are `Content-Type`, `Authorization`, `X-Request-ID`, and `X-CSRF-Token`. Allowed methods are `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, and `OPTIONS`.

## Secrets & Credential Vault

Stored credentials are encrypted with AES-256-GCM. The data key is derived from the master key with PBKDF2-HMAC-SHA256 at 100,000 iterations, and per-user keys use a per-user salt so one user's credentials cannot be decrypted with another user's derived key.

The master key lives at `server.security.encryption_key` in `server.yml`, with a companion `server.security.encryption_key_version` that starts at `1` and increments on every rotation. It is the canonical at-rest key for all sensitive server data, including 2FA secrets. There is no separate `encryption.key` file and no `ENCRYPTION_KEY` environment variable.

On first run the key is generated from 32 bytes of `crypto/rand`, base64-encoded, and written to `server.yml`, which is created with mode `0600` because it holds this key and the webhook secret. At startup the key must decode from base64 to exactly 32 bytes; anything else is a fatal error. If real entropy cannot be sourced or persisted, startup fails rather than continuing with a predictable key.

Rotate the key from the admin panel. Rotation re-prompts for the admin password, records `security.encryption_key_rotated` in the audit log, and increments the key version.

The key is never returned by any API response, even to an administrator, and never written to a log.

Operational rules that hold throughout the build pipeline:

- Secrets are injected just in time and cleaned up automatically after use
- Secrets never appear in build logs or build artifacts
- Secret values are never returned in an API response body
- The encryption key and the database connection string are stripped before any value reaches a template

Losing the master key makes every stored credential unrecoverable and forces every user to re-enroll 2FA. Back up `server.yml` with the database.

## Build Isolation

Each build runs in its own Docker container with a clean per-build workspace bind-mounted at `/workspace`, and the container is removed when the build finishes. Resource limits are applied: 2 GB of memory and 2 CPUs per build container.

!!! warning "Containers are not hardened yet"
    Build containers currently run on the default Docker bridge network with full outbound access, as root inside the container, with a writable root filesystem and the default capability set. There is no network isolation, no user-namespace remapping, no seccomp or AppArmor profile beyond the daemon default, and no per-user build quota or fair-scheduling policy.

    A build therefore executes arbitrary repository-supplied commands with whatever the Docker daemon grants it. Run CASCI on a host dedicated to it, and treat every build as untrusted code with network access.

The stronger isolation properties — network isolation, per-tenant storage separation, and quota enforcement — are design goals, not current behavior. Do not run untrusted third-party pipelines on a shared instance until they land.

## Build Supply Chain

| Scan type | Finds | Implemented by |
|-----------|-------|----------------|
| `vulnerability` | Known CVEs in dependencies and images | `trivy`, `grype` |
| `sast` | Static analysis findings in source | `semgrep` |
| `secret` | Credentials committed to the repository | `gitleaks` |
| `sbom` | Software bill of materials for the build | `syft` |
| `license` | Dependency licenses that violate policy | Built in |
| `malware` | Malicious content in dependencies or artifacts | Scan type defined; no scanner wired up |

Scanning is delegated, not embedded. CASCI invokes these tools as external commands, so each one must be installed and on the `PATH` of the process running the scan. A missing tool makes that scan fail rather than pass.

Findings are classified `critical`, `high`, `medium`, `low`, or `info`. Block-on-critical is configurable at `/api/v1/security/config`.

Artifact signing is **not** implemented. There is no cosign integration and no signature is produced for build outputs, so an artifact cannot currently be verified independently of the instance that produced it.

## Public Endpoints

These require no authentication:

| Path | Purpose |
|------|---------|
| `/server/healthz` | Human-readable health page |
| `/api/v1/server/healthz` | Machine-readable health check |
| `/readyz` | Readiness probe |
| `/livez` | Liveness probe |
| `/server/about` | Instance information |
| `/server/privacy` | Privacy policy |
| `/server/contact` | Contact form |
| `/server/help` | Help and documentation |
| `/api/v1/auth/register` | Account registration |
| `/api/v1/auth/login` | Authentication |
| `/api/v1/nodes/register` | Node registration, secured by registration token |
| `/webhook` | SCM webhook receiver, secured by provider signature |
| `/server/docs/swagger` | Swagger UI |
| `/api/v1/server/swagger`, `/api/swagger` | OpenAPI specification, JSON |
| `/api/v1/server/swagger.yaml`, `/api/swagger.yaml` | OpenAPI specification, YAML |
| `/server/docs/graphql` | GraphQL playground |
| `/.well-known/security.txt` | RFC 9116 vulnerability reporting document |
| `/api/json`, `/crumbIssuer/api/json` | Jenkins compatibility surface |
| `/static/` | Embedded CSS, JS, and images |

The API documentation surface is public. If you do not want your endpoint schema readable by anyone who can reach the instance, block `/server/docs/`, `/api/swagger*`, and `/api/v1/server/swagger*` at your reverse proxy — there is no configuration switch that turns them off.

Health responses are intentionally thin. They report whether the service is up without disclosing version detail, dependency topology, or internal addresses to an unauthenticated caller.

Metrics are not public: `/metrics`, `/metrics/json`, and every `/api/v1/metrics/*` endpoint require an admin credential.

## Well-Known Namespace

One entry is served: `/.well-known/security.txt`, generated per request in RFC 9116 form.

```
Contact: https://github.com/webappsgo/casci/security/advisories/new
Contact: mailto:security@casci.casapps.us
Policy: https://casci.casapps.us/server/security
Expires: 2027-08-22T00:00:00Z
```

The primary contact is `web.security.report_url`; the CC mailbox is `web.security.contact`, which falls back to `security@<request host>` when unset. The `Expires` stamp is rewritten one year out whenever it is missing or already past, since an expired document is invalid.

Forwarding headers are not consulted when building these URLs. That is deliberate: without a trusted-proxy allowlist, honoring `X-Forwarded-*` would let any client publish a contact URL of its choosing. The document therefore reflects the host the client actually addressed.

Every other `/.well-known/` path returns `404` and is never redirected — no `pgp-key.asc`, `change-password`, WebFinger, OpenID provider metadata, Android app links, Apple app-site association, or MTA-STS.

!!! note "The published policy URL is not yet routed"
    `security.txt` advertises `Policy: /server/security`, but no handler is mounted at that path, so it currently answers `404`. Use the reporting channels below.

## Security Reporting

Report a suspected vulnerability privately. Do not open a public issue containing exploit detail.

- Use the contact form at `https://casci.casapps.us/server/contact`
- Or open a private advisory at [github.com/webappsgo/casci/security/advisories](https://github.com/webappsgo/casci/security/advisories)

Include the affected version, reproduction steps, and the impact you observed. A running instance publishes the same channels at `/.well-known/security.txt`.

## Transport Security

Set `server.ssl.enabled` to `true` with `server.ssl.cert` and `server.ssl.key` to serve HTTPS directly. `server.ssl.min_version` pins the floor; the server refuses anything below TLS 1.2 regardless.

When TLS terminates at a reverse proxy, leave `server.ssl.enabled` off. Note that CASCI does not read `X-Forwarded-Proto` or any other forwarding header, so set `DOMAIN` (or `server.fqdn`) to the public hostname explicitly rather than expecting the proxy to convey it.

Bind to a loopback or Docker bridge address rather than a public interface when a reverse proxy fronts the service. The shipped Compose file binds `172.17.0.1:64580` for exactly this reason.

## Audit & Compliance

Audit events are queryable at `/api/v1/audit/events`, with admin-managed retention and cleanup.

Selecting a compliance mode applies that standard's expectations rather than requiring each control to be set by hand:

| Mode | Notable expectation |
|------|---------------------|
| `hipaa` | Six-year audit retention |
| `sox` | Seven-year audit retention |
| `pci-dss` | One year of access logging |
| `gdpr` | Data minimization, right to deletion, data portability |
| `fedramp` | US federal control baseline |
| `iso27001` | ISMS control baseline |
| `none` | No compliance preset applied |

Set the mode at `/api/v1/compliance/mode` and verify current state with `/api/v1/compliance/check`.

## Operator Hardening Checklist

- **Change `web.cors` from its generated `*` default to your own origins before exposing the instance**
- Set `DOMAIN` (or `server.fqdn`) so the instance knows its own public hostname; no forwarding header will tell it
- Back up `server.yml`; it holds the encryption key and the webhook secret, and neither is recoverable
- Verify `server.yml` is mode `0600` and owned by the service account
- Set `server.security.webhook_secret` before exposing `/webhook`
- Terminate TLS, either in CASCI or at a proxy, and never serve plain HTTP publicly
- Bind to a private interface and let a reverse proxy handle public exposure
- Change the default `administrator` username
- Run as the non-privileged service account created by the platform installer
- Keep `/metrics` admin-only; do not proxy it to an unauthenticated scrape target
- Choose a compliance mode if you have a retention obligation, before events start aging out
