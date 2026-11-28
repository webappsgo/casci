# API reference

This page describes the current HTTP API and its server-rendered counterparts.
The generated OpenAPI document provides the machine-readable route inventory;
authorization and handler-specific behavior are summarized here.

## Base path and authentication

The API version segment comes from `server.api_version` (default `v1`).
`GET https://casci.casapps.us/api/{api_version}` returns the discovery envelope. The slash form
`GET https://casci.casapps.us/api/{api_version}/` returns the API version information envelope.
`GET https://casci.casapps.us/api/autodiscover` provides the public discovery document. Some machine
routes have unversioned aliases.

## Frontend response formats

Frontend pages normally return server-rendered HTML. Requests from recognized
text browsers (Lynx, w3m, Links, ELinks, Browsh, Carbonyl, and NetSurf) receive
the same server-rendered content with script elements removed; navigation and
forms remain usable through links and standard form submissions. Requests from
recognized command-line HTTP tools, or requests that explicitly accept
`text/plain`, receive an 80-column text rendering of the page. API paths and
static assets keep their own response formats.

Native API handlers accept an account's opaque bearer API token in the
`Authorization` header:

```http
Authorization: Bearer <bearer-token>
```

Browser middleware accepts the matching CASCI session cookie on browser routes.
API login returns a revocable, role-scoped `adm_` or `usr_` token and does not
set a cookie. MCP accepts only the bearer header.
Login is public. Registration is mounted only when `server.users.enabled` is
`true` and `server.users.registration.mode` is `open`; otherwise the route
returns 404. User, admin, and node handlers apply different checks. Bearer
tokens are opaque values, not JWTs. One-time link codes identify a single
time-limited action and are never accepted as bearer credentials.

## Admin page management

Page management uses `https://casci.casapps.us/api/{api_version}/server/{admin_path}/pages` and
requires an administrator bearer token. `GET https://casci.casapps.us/api/{api_version}/server/{admin_path}/pages` lists the API-managed pages.
Each page can be read with `GET https://casci.casapps.us/api/{api_version}/server/{admin_path}/pages/{page}`; supported page names are
`about`, `privacy`, `terms`, `contact`, and `help`. Footer is managed only by its
server-rendered editor and is not accepted by this API. `PATCH` updates page
content with `{ "format": "md", "body": "..." }`, or contact form settings
with `{ "settings": { "enabled": true, "captcha": "simple" } }`.
`DELETE https://casci.casapps.us/api/{api_version}/server/{admin_path}/pages/{page}` removes the About, Privacy, Terms, or Help override and
restores its built-in content. Footer reset is available from its CSRF-protected
server-rendered editor; Contact settings do not have a reset operation.
`POST https://casci.casapps.us/api/{api_version}/server/{admin_path}/pages/{page}/preview` accepts the same `format` and `body` fields and
returns the raw input, sanitized HTML, and whether sanitization changed the
content without saving it. Page bodies are limited to
256 KiB. API content updates accept the documented `md`, `html`, or `txt`
formats where enabled. About, Privacy, Terms, and Footer overrides are stored in
`yaml.d/branding.yml`; Help content and Contact settings are stored in
`yaml.d/pages.yml`. Contact captcha secrets are never returned by the API.
Contact settings may select `simple`, `recaptcha`, or `hcaptcha`. Supply a new
write-only `captcha_secret_key` in contact settings to replace the encrypted
provider secret, or set `clear_captcha_secret: true` to remove it. Reads return
only `captcha_secret_configured`. Public contact descriptors identify the
selected provider and include a `fallback` simple challenge for provider-backed
forms. API submissions may send one provider `captcha_response`, or the fallback
`captcha_id` and `captcha_answer`; CAPTCHA validation errors include a fresh
descriptor in `details.captcha`.

## Public contact submissions

`GET https://casci.casapps.us/api/{api_version}/server/contact` returns the offered types, field
constraints, a CAPTCHA descriptor, and a signed `form_token`. Post the selected
type's fields, the token, and its CAPTCHA proof back to the same path. A valid
first post returns a review summary, `review_token`, a fresh `form_token`, and
`confirm_after_seconds`; it does not deliver an email. Confirm by posting the
same fields with `step: "confirm"`, the review token, and `confirmed: true`.
The review token binds the submitted fields and is single-use. Post
`step: "edit"` to receive the fields with a fresh token and CAPTCHA challenge.

Malformed fields and CAPTCHA failures return `VALIDATION_FAILED`; CAPTCHA
failures include a fresh descriptor in `details.captcha`, and form-token
failures include a fresh `details.form_token`. Rate-limit responses use HTTP
429 and `Retry-After`. Honeypot and duplicate submissions use the normal review
and success responses but are silently discarded. Delivery requires configured
SMTP. Security reports use the separate `https://casci.casapps.us/server/contact?security_id=...`
workflow linked from the public security policy.

## Health, help, preferences, and auth

| Method | Path | Access | Behavior |
|--------|------|--------|----------|
| GET | `https://casci.casapps.us/api/{api_version}` | Public | API discovery envelope; same data as `https://casci.casapps.us/api/autodiscover` |
| GET | `https://casci.casapps.us/api/v1/server/healthz` | Public | JSON health response |
| GET | `https://casci.casapps.us/api/healthz` | Public alias | Same health handler |
| GET | `https://casci.casapps.us/readyz` | Public | Readiness handler |
| GET | `https://casci.casapps.us/livez` | Public | Liveness handler |
| GET | `https://casci.casapps.us/api/v1/server/help` | Public | Help JSON and registered help pages |
| GET | `https://casci.casapps.us/api/v1/server/help/mcp` | Public | MCP connection details, tools, and project text |
| GET | `https://casci.casapps.us/api/v1/server/security` | Public | Security contact channels, policy and acknowledgment URLs, and PGP availability |
| GET | `https://casci.casapps.us/api/v1/server/security/policy` | Public | Built-in coordinated disclosure policy in Markdown |
| GET | `https://casci.casapps.us/api/v1/server/security/thanks` | Public | Credited researcher acknowledgments |
| GET, POST | `https://casci.casapps.us/api/v1/server/preferences` | Public | Preference API handler |
| GET | `https://casci.casapps.us/api/v1/server/preferences/sync` | Public | Export the current preference snapshot |
| GET | `https://casci.casapps.us/api/v1/server/preferences/sync/import` | Public | Import preferences from the URL query parameters |
| POST | `https://casci.casapps.us/api/v1/server/auth/register` | Public when multi-user mode is enabled and registration is open | Register a regular-user account |
| POST | `https://casci.casapps.us/api/v1/server/auth/login` | Public | Authenticate a server administrator or enabled regular user; returns a revocable 24-hour `adm_` or `usr_` Bearer API token named `login` and sets no cookie |
| GET | `https://casci.casapps.us/api/v1/server/auth/check` | Public; Bearer token required | Validate an administrator or user credential and return `data.kind` (`operator` or `user`) without returning the token; invalid or missing credentials return 401 |
| POST | `https://casci.casapps.us/api/v1/server/auth/logout` | Bearer token | Revoke the login API token |
| POST | `https://casci.casapps.us/api/v1/server/auth/2fa` | Partial login token | Complete TOTP or recovery-code verification |
| POST | `https://casci.casapps.us/api/v1/server/auth/refresh` | Authenticated | Refresh a supported signed session token |
| POST | `https://casci.casapps.us/api/v1/server/auth/password/forgot` | Public, rate-limited | Request password reset instructions without revealing whether the email exists |
| POST | `https://casci.casapps.us/api/v1/server/auth/password/reset` | One-time reset token | Set a new password and consume the reset token |
| POST | `https://casci.casapps.us/api/v1/server/auth/username/forgot` | Public, rate-limited | Request a username reminder without revealing whether the email exists |
| POST | `https://casci.casapps.us/api/v1/server/auth/verify` | One-time verification token | Verify the account email address |
| GET | `https://casci.casapps.us/api/v1/server/auth/oidc/{provider}` | Public | Begin browser OIDC authorization with state, nonce, and PKCE |
| GET | `https://casci.casapps.us/server/auth/oidc/{provider}/callback` | OIDC provider callback | Validate state/nonce, exchange the code, and establish or complete external account linking |
| POST | `https://casci.casapps.us/api/v1/server/auth/oidc/{provider}/callback` | Public, one-time request ID | Poll an API-mode OIDC login transaction and consume its session result |
| GET, POST | `https://casci.casapps.us/server/auth/ldap/{provider}` | Public; browser POST uses CSRF | Authenticate through a configured LDAP provider and establish or link an account |
| POST | `https://casci.casapps.us/api/v1/server/auth/ldap/{provider}` | Public | Authenticate through a configured LDAP provider using JSON |
| GET | `https://casci.casapps.us/server/auth/saml/{provider}` | Public | Start browser SAML sign-in |
| POST | `https://casci.casapps.us/server/auth/saml/{provider}/acs` | SAML IdP response | Validate the signed assertion and establish or link an account |
| GET | `https://casci.casapps.us/server/auth/saml/{provider}/metadata` | Public | Publish this provider's SAML service-provider metadata |
| GET, POST | `https://casci.casapps.us/server/auth/saml/{provider}/slo/callback` | SAML IdP logout response | Complete service-provider initiated SAML logout |
| GET | `https://casci.casapps.us/api/v1/server/auth/saml/{provider}` | Public | Start API-mode SAML sign-in |
| POST | `https://casci.casapps.us/api/v1/server/auth/saml/{provider}/exchange` | Public, one-time request ID | Poll an API-mode SAML login transaction and consume its session result |
| POST | `https://casci.casapps.us/api/v1/server/auth/external/username` | One-time pending identity | Confirm the local username for an external account link |
| GET, POST | `https://casci.casapps.us/api/{api_version}/server/auth/invite/user/{token}` | One-time invite code | Validate or accept a regular-user invitation when user accounts are enabled |
| GET, POST | `https://casci.casapps.us/api/{api_version}/server/auth/invite/server/{token}` | One-time invite code | Validate or accept a server-administrator invitation |
| POST | `https://casci.casapps.us/api/v1/server/reports/{type}` | Public, anonymous, rate-limited | Accept allow-listed browser telemetry, sanitize it, and write only to operational logs; success returns 204 |
| GET | `https://casci.casapps.us/api/v1/server/{admin_path}/setup` | Public | Report setup requirement and primary-account recovery status |
| POST | `https://casci.casapps.us/api/v1/server/{admin_path}/setup/verify` | Setup token | Verify the one-time first-run setup token |
| POST | `https://casci.casapps.us/api/v1/server/{admin_path}/setup/account` | Setup token | Create the primary administrator and return the one-time `adm_` API token |
| POST | `https://casci.casapps.us/api/v1/server/{admin_path}/setup/complete` | Administrator session | Complete the first-run setup after the required steps are saved |
| GET | `https://casci.casapps.us/api/v1/server/{admin_path}/{admin_username}/profile` | Administrator bearer token | Read the authenticated administrator profile |
| PATCH | `https://casci.casapps.us/api/v1/server/{admin_path}/{admin_username}/profile` | Administrator bearer token | Update the account email after current-password verification |
| POST | `https://casci.casapps.us/api/v1/server/{admin_path}/{admin_username}/profile/password` | Administrator bearer token | Change the administrator password after current-password verification |
| GET | `https://casci.casapps.us/api/v1/server/{admin_path}/{admin_username}/profile/security` | Administrator session | Read TOTP, passkey, and recovery-code status |
| POST | `https://casci.casapps.us/api/v1/server/{admin_path}/{admin_username}/profile/security/totp/setup` | Administrator session | Start TOTP enrollment after current-password verification; return the secret and provisioning URI |
| POST | `https://casci.casapps.us/api/v1/server/{admin_path}/{admin_username}/profile/security/totp/enable` | Administrator session | Verify TOTP and return one-time recovery codes; a separate acknowledgment enables TOTP |
| POST | `https://casci.casapps.us/server/{admin_path}/{admin_username}/profile/security` | Administrator session + CSRF | Disable TOTP after verifying a current TOTP or recovery code |
| POST | `https://casci.casapps.us/api/v1/server/{admin_path}/{admin_username}/profile/security/passkey/setup` | Administrator session | Verify the current password and begin WebAuthn registration |
| POST | `https://casci.casapps.us/api/v1/server/{admin_path}/{admin_username}/profile/security/passkey/register` | Administrator session | Finish WebAuthn registration; first enrollment returns one-time recovery codes pending acknowledgment; the security page also posts rename/revoke actions here |
| POST | `https://casci.casapps.us/api/v1/server/auth/passkey/challenge` | Public | Request a primary-login or partial-session passkey challenge |
| POST | `https://casci.casapps.us/api/v1/server/auth/passkey/verify` | Public | Verify a passkey and issue an administrator session |
| POST | `https://casci.casapps.us/api/v1/server/auth/2fa` | Partial administrator session | Complete login with a TOTP code |
| POST | `https://casci.casapps.us/api/v1/server/auth/recovery/use` | Partial administrator session | Complete login with a one-time recovery code |

The HTML login page is `https://casci.casapps.us/server/auth/login`; the registration page is
`https://casci.casapps.us/server/auth/register` when regular-user accounts are enabled and public
registration is open. Administrator accounts can use a passkey as a primary
credential. Accounts requiring a second factor continue
at `https://casci.casapps.us/server/auth/2fa`, which accepts TOTP, a passkey, or a one-time recovery
code. Passkey ceremonies use short-lived, one-use server-side challenges. The
`https://casci.casapps.us/healthz` root alias is
disabled unless `server.healthz.root.enabled` is true.
Password recovery and invitation acceptance also have no-JavaScript browser
pages at `https://casci.casapps.us/server/auth/password/forgot`, `https://casci.casapps.us/server/auth/password/reset`,
`https://casci.casapps.us/server/auth/invite/user/{code}`, and `https://casci.casapps.us/server/auth/invite/server/{code}`.
Recovery and invite codes are stored as hashes, expire, and are redacted from
request logs. Email-dependent actions require a working SMTP configuration.

Administrator API credentials can be managed from
`https://casci.casapps.us/server/{admin_path}/security/tokens`. The page creates named credentials,
shows each new value once, lists safe metadata, and revokes credentials with a
CSRF-protected form. The API supports listing, revocation, and one-time
rotation:

| Method | Path | Access | Behavior |
|--------|------|--------|----------|
| GET | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/security/tokens` | Administrator bearer token | List safe metadata for active administrator, user, and organization tokens; secret values are never returned |
| DELETE | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/security/tokens/{id}` | Administrator bearer token | Revoke a non-agent token across account types |
| POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/security/tokens/{id}/rotate` | Administrator bearer token | Replace a token secret, preserving owner, name, scope, and expiry; returns the new value once |
| GET | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/{admin_username}/profile/token` | Administrator bearer token | Read masked metadata for the setup-issued token belonging to the signed-in administrator |
| POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/{admin_username}/profile/token` | Administrator bearer token | Regenerate the setup-issued token; returns its new value once |

## Users

| Method | Path | Access |
|--------|------|--------|
| GET, PATCH | `https://casci.casapps.us/api/v1/users` | Authenticated user; reads or updates the caller's own profile |
| GET, POST | `https://casci.casapps.us/api/v1/users/tokens` | Authenticated user; list token metadata or create a named `usr_` token (the value is returned once) |
| GET, DELETE | `https://casci.casapps.us/api/v1/users/tokens/{token_id}` | Authenticated owner; inspect metadata or immediately revoke a token |
| POST | `https://casci.casapps.us/api/v1/users/security/password` | Authenticated user; verify the current password and change the caller's password |
| GET | `https://casci.casapps.us/api/v1/users/security/sessions` | Authenticated user; list the caller's API-token sessions |
| DELETE | `https://casci.casapps.us/api/v1/users/security/sessions/{session_id}` | Authenticated owner; revoke one API-token session |
| GET | `https://casci.casapps.us/api/v1/users/security/2fa` | Authenticated user; read TOTP, passkey, and recovery-code status |
| POST | `https://casci.casapps.us/api/v1/users/security/2fa/enable` | Authenticated user; begin or confirm TOTP enrollment and return one-time recovery codes |
| POST | `https://casci.casapps.us/api/v1/users/security/2fa/disable` | Authenticated user; disable TOTP after verifying a current TOTP or recovery code |
| GET | `https://casci.casapps.us/api/v1/users/security/recovery` | Authenticated user; read remaining recovery-code status |
| POST | `https://casci.casapps.us/api/v1/users/security/recovery/regenerate` | Authenticated user; replace recovery codes and return them once |
| GET, POST | `https://casci.casapps.us/api/v1/users/security/passkeys` | Authenticated user; list passkeys or begin registration |
| DELETE | `https://casci.casapps.us/api/v1/users/security/passkeys/{passkey_id}` | Authenticated owner; revoke one passkey |
| GET, POST | `https://casci.casapps.us/users/consents` | User session; POST requires CSRF and supports a standard HTML form |
| GET, PATCH | `https://casci.casapps.us/api/v1/users/consents` | User bearer token; read or update consent settings |
| GET | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/users` | Administrator; lists accounts |
| POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/users` | Administrator; creates a pending user record, emails or returns its single-use activation URL, and lets the user set their password (multi-user mode) |
| GET, PUT, PATCH, DELETE | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/users/{id}` | Administrator; manages one account |
| GET, POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/users/invites` | Administrator; list or create user/admin invitations (the raw code is returned once) |
| GET, DELETE | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/users/invites/{id}` | Administrator; inspect safe metadata or revoke an invitation |
| GET, PATCH | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/email` | Administrator; read safe SMTP settings or save settings (password is write-only) |
| POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/email/test` | Administrator; send a `[TEST]` message using a selected template and recipient (defaults to the administrator email and `test` template) |
| GET | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/email/templates` | Administrator; list template names, subjects, account classification, and default/custom status |
| GET, PUT | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/email/templates/{name}` | Administrator; read or save a template's plain-text subject and body |
| POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/email/templates/{name}/preview` | Administrator; render a subject and body with sample data |
| POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/email/templates/{name}/reset` | Administrator; remove the custom file and return the embedded default |
| GET, POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/security/auth/{oidc,ldap,saml}/providers` | Administrator | List safe provider settings or create a provider |
| GET, PATCH, DELETE | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/security/auth/{oidc,ldap,saml}/providers/{name}` | Administrator | Read safe settings, update, or remove an identity provider |
| POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/security/auth/{oidc,ldap,saml}/providers/{name}/test` | Administrator | Test provider connectivity without returning credentials |
| GET, POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/backup` | Administrator | List safe archive metadata or create and verify a full backup; the optional password is never returned |
| GET, PUT | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/backup/settings` | Administrator | Read or update scheduled backup enablement, hourly incrementals, and validated retention settings |
| GET, DELETE | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/backup/{id}` | Administrator | Read archive metadata or delete it when no retained incremental depends on it |
| GET | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/backup/{id}/download` | Administrator | Download a backup archive as an attachment |
| POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/backup/test` | Administrator | Verify an existing backup archive using its filename and optional password |
| POST | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/backup/restore` | Administrator | Verify and restore an archive, then restart the server; encrypted archives require a password and all restores require typing the exact filename in `confirmation` |

User management routes are under the configured admin path in a running
instance. Regular users cannot use the admin API path or read another user's
profile through `https://casci.casapps.us/api/v1/users`.

The server-rendered profile page is `https://casci.casapps.us/users`; it uses the same self-scoped
account data and accepts a no-JavaScript form POST for email or password
updates. `https://casci.casapps.us/users/tokens` and, when regular-user accounts are enabled,
`https://casci.casapps.us/server/preferences/tokens` list the caller's tokens and support no-JavaScript
create and revoke forms. Token values are returned only at creation; the
database stores a SHA-256 digest and the first eight characters for display.
The `https://casci.casapps.us/users/security` page supports password changes, TOTP enrollment,
passkey management, and recovery-code acknowledgment. `https://casci.casapps.us/users/security/sessions`
lists API-token sessions and provides CSRF-protected revoke forms. Both pages
use no-store/noindex headers.
`https://casci.casapps.us/users/consents` accepts both a CSRF-protected POST form and the PATCH JSON
action. The administrator invitation page is
`https://casci.casapps.us/server/{admin_path}/users/invites`; SMTP settings and template editing are
available at `https://casci.casapps.us/server/{admin_path}/email` and
`https://casci.casapps.us/server/{admin_path}/email/templates` with CSRF-protected forms.

## Projects and builds

| Method | Path | Access | Handler behavior |
|--------|------|--------|------------------|
| GET | `https://casci.casapps.us/api/v1/projects` | Authenticated | List projects |
| POST | `https://casci.casapps.us/api/v1/projects` | Authenticated | Create project |
| GET, PATCH, DELETE | `https://casci.casapps.us/api/v1/projects/{id}` | Authenticated owner | Read, update, or delete one owned project |
| GET | `https://casci.casapps.us/api/v1/projects/{id}/pipelines` | Authenticated owner | List configured pipelines or detect workflows from an HTTPS repository |
| GET | `https://casci.casapps.us/api/v1/projects/{id}/pipelines/{pipeline_id}` | Authenticated owner | Read one configured pipeline or detected workflow definition |
| GET, POST | `https://casci.casapps.us/api/v1/projects/{id}/builds` | Authenticated | List or trigger builds |
| GET | `https://casci.casapps.us/api/v1/projects/{id}/builds/stats` | Authenticated | Project build statistics |
| GET | `https://casci.casapps.us/api/v1/projects/{id}/notifications` | Authenticated | Project notification configurations |
| GET, POST | `https://casci.casapps.us/api/v1/projects/{id}/credentials` | Authenticated | List or create project credentials |
| GET | `https://casci.casapps.us/api/v1/builds/{id}` | Authenticated | Read a build |
| GET | `https://casci.casapps.us/api/v1/builds/{id}/logs` | Authenticated | Read a build log |
| POST | `https://casci.casapps.us/api/v1/builds/{id}/cancel` | Authenticated | Cancel a build |
| GET | `https://casci.casapps.us/api/v1/builds/{id}/artifacts` | Authenticated owner | List safe metadata for the build's artifacts |
| GET | `https://casci.casapps.us/api/v1/builds/{id}/artifacts/{artifact_id}` | Authenticated owner | Download an owned artifact with attachment and `nosniff` headers |
| GET | `https://casci.casapps.us/api/v1/builds/{id}/security` | Authenticated | Read build security reports |
| POST | `https://casci.casapps.us/api/v1/builds/{id}/security/scan` | Authenticated | Trigger a scan |
| GET | `https://casci.casapps.us/api/v1/builds/{id}/notifications` | Authenticated | Read build notification data |

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
| GET, POST | `https://casci.casapps.us/api/v1/infrastructure/connections` | List or create cloud, private-server, or storage connection metadata |
| GET, PATCH, DELETE | `https://casci.casapps.us/api/v1/infrastructure/connections/{connection_id}` | Read, update, or delete one caller-owned connection |
| GET, PUT | `https://casci.casapps.us/api/v1/budgets` | Read or replace the caller's monthly and per-provider budget limits and available spend data |
| POST | `https://casci.casapps.us/api/v1/infrastructure/hourly-instances/destroy-all` | Returns `503 MAINTENANCE` while no provider cleanup backend is configured |

The corresponding no-JavaScript pages are `https://casci.casapps.us/users/infrastructure` and
`https://casci.casapps.us/users/budgets`. Their state-changing forms use CSRF tokens.

## Nodes

| Method | Path | Access |
|--------|------|--------|
| GET | `https://casci.casapps.us/api/v1/nodes` | Administrator |
| POST | `https://casci.casapps.us/api/v1/nodes/token` | Administrator |
| POST | `https://casci.casapps.us/api/v1/nodes/register` | Registration token |
| GET, PUT, PATCH, DELETE | `https://casci.casapps.us/api/v1/nodes/{id}` | Administrator |
| POST | `https://casci.casapps.us/api/v1/nodes/{id}/heartbeat` | Node handler |
| POST | `https://casci.casapps.us/api/v1/nodes/{id}/drain` | Administrator |

## Agents

Agent management is available to administrators under the configured admin
path and to signed-in users under `https://casci.casapps.us/api/v1/users/agents`. Management requests
require the matching account credential; user routes require a user bearer token.
Machine registration, heartbeat, report, and binary download requests require
an agent-scoped `adm_` or `usr_` bearer token. The full route behavior and
payload examples are in the [Agent guide](agents.md).

| Method | Path | Access |
|--------|------|--------|
| GET, POST | `https://casci.casapps.us/api/v1/server/{admin_path}/agents` | Administrator |
| GET, PATCH, DELETE | `https://casci.casapps.us/api/v1/server/{admin_path}/agents/{id}` | Administrator |
| POST | `https://casci.casapps.us/api/v1/server/{admin_path}/agents/{id}/token` | Administrator |
| GET | `https://casci.casapps.us/api/v1/server/{admin_path}/agents/{id}/metrics` | Administrator |
| POST | `https://casci.casapps.us/api/v1/server/{admin_path}/agents/{register,heartbeat,report}` | Admin-scoped agent token |
| GET, POST | `https://casci.casapps.us/api/v1/users/agents` | Authenticated user |
| GET, PATCH, DELETE | `https://casci.casapps.us/api/v1/users/agents/{id}` | Authenticated user; owner scoped |
| POST | `https://casci.casapps.us/api/v1/users/agents/{id}/token` | Authenticated user; owner scoped |
| GET | `https://casci.casapps.us/api/v1/users/agents/{id}/metrics` | Authenticated user; owner scoped |
| POST | `https://casci.casapps.us/api/v1/users/agents/{register,heartbeat,report}` | User-scoped agent token |
| GET, HEAD | `https://casci.casapps.us/agent/binaries/{name}` | Any valid agent-scoped bearer token |

Agent management API paths return a JSON `ok`/`data` envelope. Agent tokens
are stored as SHA-256 hashes and are shown only when created or rotated.
Organization-scoped agent routes are not currently implemented.

## Credentials and notifications

| Method | Path | Access |
|--------|------|--------|
| GET, POST | `https://casci.casapps.us/api/v1/credentials/user` | Authenticated |
| GET, PUT, PATCH, DELETE | `https://casci.casapps.us/api/v1/credentials/user/{id}` | Authenticated |
| GET, PUT, PATCH, DELETE | `https://casci.casapps.us/api/v1/credentials/project/{id}` | Authenticated |
| GET, POST | `https://casci.casapps.us/api/v1/notifications` | Authenticated |
| GET, PUT, PATCH, DELETE | `https://casci.casapps.us/api/v1/notifications/{id}` | Authenticated |
| POST | `https://casci.casapps.us/api/v1/notifications/test` | Authenticated |
| GET | `https://casci.casapps.us/api/v1/notifications/types` | Authenticated |
| GET | `https://casci.casapps.us/api/v1/notifications/events` | Authenticated |

The credential endpoints are sensitive. Do not expose their responses to
untrusted clients or model contexts.

## Security, audit, compliance, and metrics

| Method | Path | Access |
|--------|------|--------|
| GET | `https://casci.casapps.us/api/v1/security/reports` | Authenticated |
| GET | `https://casci.casapps.us/api/v1/security/reports/{id}` | Authenticated |
| GET | `https://casci.casapps.us/api/v1/security/statistics` | Authenticated |
| GET, PUT | `https://casci.casapps.us/api/v1/security/config` | Administrator |
| GET | `https://casci.casapps.us/api/v1/server/{admin_path}/security/locked-accounts` | Administrator Bearer token; list locked user and administrator accounts with `kind`, `id`, and `locked_until` |
| DELETE | `https://casci.casapps.us/api/v1/server/{admin_path}/security/locked-accounts/{kind}/{id}` | Administrator Bearer token; unlock a `user` or `admin` account |
| GET, POST | `https://casci.casapps.us/api/v1/server/{admin_path}/security/allowlist` | Administrator Bearer token; list entries or add an IP/CIDR with an optional description |
| DELETE | `https://casci.casapps.us/api/v1/server/{admin_path}/security/allowlist/{id}` | Administrator Bearer token; remove an entry by its ID |
| GET | `https://casci.casapps.us/api/v1/server/{admin_path}/security/allowlist/check/{ip}` | Administrator Bearer token; check an IPv4 or IPv6 address against configured entries |
| GET | `https://casci.casapps.us/api/v1/server/{admin_path}/logs` | Administrator Bearer token; list log files |
| GET | `https://casci.casapps.us/api/v1/server/{admin_path}/logs/{type}` | Administrator Bearer token; read entries with `limit`, `offset`, and `search` |
| GET | `https://casci.casapps.us/api/v1/server/{admin_path}/logs/{type}/download` | Administrator Bearer token; download a log file |
| GET | `https://casci.casapps.us/api/v1/audit/events` | Authenticated |
| GET | `https://casci.casapps.us/api/v1/audit/status` | Administrator |
| PUT, PATCH | `https://casci.casapps.us/api/v1/audit/config` | Administrator |
| POST | `https://casci.casapps.us/api/v1/audit/cleanup` | Administrator |
| GET, PUT, PATCH | `https://casci.casapps.us/api/v1/compliance/config` | Administrator |
| GET | `https://casci.casapps.us/api/v1/compliance/modes` | Administrator |
| GET | `https://casci.casapps.us/api/v1/compliance/preset` | Administrator |
| POST, PUT | `https://casci.casapps.us/api/v1/compliance/mode` | Administrator |
| POST | `https://casci.casapps.us/api/v1/compliance/check` | Administrator |
| GET | `https://casci.casapps.us/server/metrics/{prometheus,grafana,loki}` | Per-service bearer token; internal Prometheus, Grafana-compatible, or bounded Loki data |
| GET | `https://casci.casapps.us/api/v1/server/metrics/{prometheus,grafana,loki}` | Per-service bearer token; versioned alias |
| GET | `https://casci.casapps.us/api/metrics/{service}` | Per-service bearer token; unversioned API alias |
| GET | `https://casci.casapps.us/metrics/{prometheus,grafana,loki}` | Per-service bearer token; optional root alias |

## MCP

The canonical endpoint is `POST https://casci.casapps.us/api/v1/server/mcp`; direct aliases are
`POST https://casci.casapps.us/api/mcp` and `POST https://casci.casapps.us/mcp`. Requests require a bearer token and the MCP
protocol headers. This implementation currently exposes a subset of the API
operation list as tools. See [MCP help](https://casci.casapps.us/server/help/mcp)
for the current tool list and request format.

## Documentation and compatibility routes

| Method | Path | Behavior |
|--------|------|----------|
| GET | `https://casci.casapps.us/api/v1/server/swagger` | Generated OpenAPI JSON |
| GET | `https://casci.casapps.us/api/swagger` | Unversioned OpenAPI JSON alias |
| GET | `https://casci.casapps.us/server/docs/swagger` | Swagger UI |
| GET, POST | `https://casci.casapps.us/api/v1/server/graphql` | GraphQL handler; resolver coverage is incomplete |
| GET, POST | `https://casci.casapps.us/api/graphql` | Unversioned GraphQL alias |
| GET | `https://casci.casapps.us/server/docs/graphql` | GraphQL playground |
| GET | `https://casci.casapps.us/api/json` | Jenkins-shaped server response |
| GET | `https://casci.casapps.us/crumbIssuer/api/json` | Jenkins-shaped crumb response |
| POST | `https://casci.casapps.us/api/v1/webhooks/{provider}` | Provider-signature verified webhook; path provider must match provider headers |
| POST | `https://casci.casapps.us/webhook` | Legacy provider webhook handler |
| GET | `https://casci.casapps.us/ws` | WebSocket endpoint |

OpenAPI uses JSON only. There is no YAML endpoint.

## Error and response formats

Response envelopes vary by package. Some handlers return `ok`, `error`, and
`message`; others use package-specific JSON or `http.Error`. Do not assume a
single error schema across the API. Validation, pagination, compatibility,
and generated-schema behavior are still being standardized.
