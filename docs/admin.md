# Admin panel

The default admin root is `https://casci.casapps.us/server/administration`; `server.admin_path` changes
that segment. Registered admin routes include:

| Page | Default path | Purpose |
|------|-------------|---------|
| Dashboard | `https://casci.casapps.us/server/administration` | Server overview |
| Profile | `https://casci.casapps.us/server/administration/{admin_username}/profile` | Change the signed-in administrator's email or password |
| Account security | `https://casci.casapps.us/server/administration/{admin_username}/profile/security` | Enroll and manage TOTP, passkeys, and recovery codes |
| Settings | `https://casci.casapps.us/server/administration/settings` | Read the settings currently exposed by the server configuration model |
| Branding | `https://casci.casapps.us/server/administration/branding` | Edit the application name and tagline |
| SSL/TLS | `https://casci.casapps.us/server/administration/ssl` | Review the active TLS configuration |
| Email | `https://casci.casapps.us/server/administration/email` | Configure SMTP sender/connection details, send a test message, and save credentials to encrypted secrets storage |
| Email templates | `https://casci.casapps.us/server/administration/email/templates` | Edit, preview, send a sample message, or reset embedded email templates |
| Logs | `https://casci.casapps.us/server/administration/logs` | Filter recent access, error, security, or audit logs; download or clear non-audit logs |
| Audit log | `https://casci.casapps.us/server/administration/logs/audit` | Review, filter, paginate, and export retained audit events as JSON or CSV |
| Security | `https://casci.casapps.us/server/administration/security/{auth,tokens,ratelimit,firewall,allowlist}` | Security and access-control section pages |
| Disclosure reports | `https://casci.casapps.us/server/administration/security/reports` and `https://casci.casapps.us/server/administration/security/reports/{tracking_id}` | Filter and sort encrypted researcher reports; view decrypted details in an authenticated, audited session; update status, severity, researcher comment, and expected disclosure date |
| Network | `https://casci.casapps.us/server/administration/network/{tor,i2p,geoip,blocklists}` | GeoIP country policy and freshness, external blocklist settings and source management, update controls, and IP checks |
| User operations | `https://casci.casapps.us/server/administration/moderation/users`, `https://casci.casapps.us/server/administration/users/invites` | Review moderation status and create/revoke expiring user or server-admin invitations |
| Scheduler | `https://casci.casapps.us/server/administration/scheduler` | View, edit, and trigger scheduled tasks |
| Backup | `https://casci.casapps.us/server/administration/backup` | Configure schedule and retention, create, list, download, and verify archives; restore an archive after entering its password (when encrypted) and typing its exact filename to confirm; the server restarts through its supervisor |
| Maintenance | `https://casci.casapps.us/server/administration/maintenance` | Current status and guidance for critical errors; startup database failures keep health and maintenance guidance available while configured retries run; manual mode controls remain incomplete |
| Updates | `https://casci.casapps.us/server/administration/updates` | Review status and change the persisted release channel |
| Cluster | `https://casci.casapps.us/server/administration/nodes`, `https://casci.casapps.us/server/administration/nodes/add` | Node and enrollment section pages |
| Metrics | `https://casci.casapps.us/server/administration/metrics` | System, build, node, security, and API measurements |
| Server info | `https://casci.casapps.us/server/administration/info` | Runtime, version, process, and database driver details |
| Pages | `https://casci.casapps.us/server/administration/pages` | Edit the About, Privacy, Terms, Footer, Contact, and Help page settings |
| Page editor | `https://casci.casapps.us/server/administration/pages/{about,privacy,terms,footer,contact,help}` | Edit public page overrides or contact form settings; Footer uses the server-rendered form only; CAPTCHA secrets are encrypted and write-only; preview and reset supported content overrides |
| Users | `https://casci.casapps.us/server/administration/users` | View accounts and update an account's email or password |
| Agents | `https://casci.casapps.us/server/administration/agents`, `https://casci.casapps.us/server/administration/agents/add`, `https://casci.casapps.us/server/administration/agents/remove` | Create scoped agent credentials, review reports, rotate credentials, and confirm removals |

The admin path segment is configurable with `server.admin_path`. Guests are
redirected to `https://casci.casapps.us/server/auth/login`; authenticated regular users are redirected
to `https://casci.casapps.us/users`. Administrators use a separate account table and session cookie.
The Email, Email templates, and Invitations pages use server-rendered forms with CSRF protection. Email templates use embedded defaults and store administrator overrides in `{config_dir}/template/email/` with owner-only file permissions. The editor lists account and operational messages, validates template variables, previews sample data, sends test mail to the entered address (defaulting to the signed-in administrator email), and restores a default after explicit confirmation.
Several section routes remain informational rather than full management interfaces. The Authentication section links to external provider configuration; Tor, I2P, and some cluster controls are still incomplete. The GeoIP page edits country deny/allow lists, saves and loads operator-authored presets, and displays database availability and freshness. Country filtering fails open when lookup data is unavailable. The blocklists page manages the built-in FireHOL feed and configured HTTPS IP/domain sources, toggles enforcement and scheduled updates, runs updates, and checks IP addresses. Feed status reports rule counts, last update, errors, and 72-hour freshness. The allowlist page supports adding and removing trusted IP addresses and CIDR ranges, and checking an address against current entries; changes are audited, and allowlisted addresses still require authentication.

The audit log page is `https://casci.casapps.us/server/administration/logs/audit`. It supports
category, severity, result, date, actor, and IP filters, 50-event pages,
JSON/CSV exports for the current page, all matching events, or the full log,
and displays the configured database retention period. Category and severity
are persisted and filterable in the database. Each newly logged event is also
appended as JSONL to `{log_dir}/audit.log`; the file chains entry hashes and
uses a protected tail checkpoint and rotation anchor. Its rotation interval,
size limit, retention, and compression follow server.logs.audit; defaults are
daily rotation with no retained archives. Event-category controls are not
configurable yet. `audit.log` is authoritative; the database
table is a best-effort query mirror, and mirror failures create an
`audit.mirror_failed` file event without blocking the original audit write.
The admin viewer and exports read the database mirror. The active log stores a
tail checkpoint to detect edits to its final record.

The logs page shows access, error, and security files, supports type/search/page
filters, optional 30-second refresh, downloads, and CSRF-protected clearing for
non-audit files. Audit records remain append-only and link to the dedicated
viewer. The API log routes require an administrator Bearer token. Access,
error, and security log formats, enable flags, rotation intervals, size limits,
retention, and compression are read from server.logs; access supports Apache,
Nginx, and JSON output, errors support text or JSON, and security supports
fail2ban, syslog, CEF, JSON, or text. Application events are written to
server.log and errors are also written to error.log. debug.log receives debug
events only when explicitly enabled. File-log search is bounded to the most
recent 10,000 lines.

Both the API and browser login flows record admin and user successes and
failures. Failed user events omit the attempted username; admin failures retain
the attempted admin name. Login and logout events include the request IP and
user agent, and logout records the session duration when the cookie is valid.
Rate-limit rejections are recorded as security events.

## Sign in

The browser sign-in page is `https://casci.casapps.us/server/auth/login`. It authenticates server
administrators first and, when regular users are enabled, regular users next.
Administrator credentials are stored in the `admins` table; regular-user
registration never grants server-administrator access. On first startup, the
server prints a one-time setup token directly to the console and serves
`https://casci.casapps.us/server/{admin_path}/setup`; it does not persist the plaintext token in a log
file. That page creates the primary administrator with the default username
`steward` (`administrator` is reserved), and stores only the token hash in
`{config_dir}/setup_token.txt`. Its no-JavaScript form submits to a
CSRF-protected server-rendered route. The current setup flow does not yet
implement all setup wizard steps required by AI.md.
The setup response displays the initial `adm_` API token once; the database
stores only its hash.

The Agents page is also available to signed-in users at `https://casci.casapps.us/users/agents` for
their own agents. Admin and user agent management APIs are described in the
[Agent guide](agents.md). Organization agent management is not currently
registered.

Successful browser logins set role-specific HTTP-only cookies. The API login
route is `POST https://casci.casapps.us/api/{api_version}/server/auth/login`; it returns a revocable,
role-scoped `adm_` or `usr_` bearer API token and sets no cookie. First-run
setup displays the initial administrator token once. The administrator cookie is
`admin_session` (30-day default); the regular-user cookie is `user_session`
(7-day default). Admin API routes require an administrator bearer token.

## Current settings page

The settings page is read-only and reflects only fields currently carried by
the server config model. It is not a full editor for application configuration.
Settings do not generally hot-reload. The metrics and server information pages
display live collector values and process metadata. Consult
[Configuration](configuration.md) for the current configuration source and
supported keys.

The Pages editor stores About, Privacy, Terms, and Footer content in
`{config_dir}/yaml.d/branding.yml` under `server.branding.content`. Help page
content is stored in `server.pages.help.content` in
`{config_dir}/yaml.d/pages.yml`; MCP-specific guidance remains in the sibling
`server.pages.help.mcp` setting. Contact page settings are stored under
`server.pages.contact` in `pages.yml`. Page bodies are limited to 256 KiB,
written atomically with owner-only permissions, and sanitized before preview
and public rendering. The editor uses standard POST forms with CSRF tokens,
so page editing and preview work without JavaScript. The API editor requires
an administrator bearer token.

SMTP settings are available at `https://casci.casapps.us/server/{admin_path}/email` and the matching
`https://casci.casapps.us/api/{api_version}/server/{admin_path}/email` routes. Passwords are encrypted
in `secrets.yml` and never returned. Invitation management is available at
`https://casci.casapps.us/server/{admin_path}/users/invites` and its API routes; each raw invite code is
shown only at creation.

With multi-user mode enabled, `POST https://casci.casapps.us/api/{api_version}/server/{admin_path}/users`
creates a pending user record from the username and optional email, then returns
a single-use activation URL. The Users page provides the same flow with a
CSRF-protected form; when SMTP is configured, the activation link is emailed
automatically if an email address was supplied, and otherwise the page shows a
link for manual delivery. The
record cannot be accessed with a password until the new user chooses one by
redeeming the link. The link expires after seven days.

The API user-management routes are also scoped below the configured admin path.
The public `https://casci.casapps.us/api/{api_version}/users` endpoint addresses only the authenticated
caller's profile, shown on `https://casci.casapps.us/users`.

## Admin API routes

The following API route groups apply administrator checks in the current
router:

- `https://casci.casapps.us/api/v1/server/{admin_path}/users/**`
- `https://casci.casapps.us/api/v1/server/{admin_path}/email` and `https://casci.casapps.us/api/v1/server/{admin_path}/email/test`
- `https://casci.casapps.us/api/v1/server/{admin_path}/users/invites/**`
- `GET https://casci.casapps.us/api/v1/nodes` and node CRUD/drain routes
- `POST https://casci.casapps.us/api/v1/nodes/token`
- `GET` and `PUT https://casci.casapps.us/api/v1/security/config`
- `https://casci.casapps.us/api/v1/audit/status`, `https://casci.casapps.us/api/v1/audit/config`, and
  `https://casci.casapps.us/api/v1/audit/cleanup`
- `https://casci.casapps.us/api/v1/compliance/**`
- `https://casci.casapps.us/metrics`, `https://casci.casapps.us/metrics/json`, and `https://casci.casapps.us/api/v1/metrics/**`

See [API Reference](api.md) for methods and route details. API behavior and
authorization are still being completed; do not treat the presence of an admin
page or route as proof that every server-management function is implemented.

## Roles

Server administrators and regular users are separate account types. The
complete product design has additional tenant-isolation requirements that are
not yet fully implemented. Administrators should not be assumed to have safe
access boundaries over other users' projects, credentials, or build data until
the relevant multi-user work is complete.
