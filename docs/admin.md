# Admin panel

The default admin root is `/server/administration`; `server.admin_path` changes
that segment. Registered admin routes include:

| Page | Default path | Purpose |
|------|-------------|---------|
| Dashboard | `/server/administration` | Server overview |
| Profile | `/server/administration/{admin_username}/profile` | Change the signed-in administrator's email or password |
| Account security | `/server/administration/{admin_username}/profile/security` | Enroll and manage TOTP, passkeys, and recovery codes |
| Settings | `/server/administration/settings` | Read the settings currently exposed by the server configuration model |
| Branding | `/server/administration/branding` | Edit the application name and tagline |
| SSL/TLS | `/server/administration/ssl` | Review the active TLS configuration |
| Email | `/server/administration/email` | SMTP configuration status |
| Logs | `/server/administration/logs` | Filter recent access, error, security, or audit logs; download or clear non-audit logs |
| Audit log | `/server/administration/logs/audit` | Review, filter, paginate, and export retained audit events as JSON or CSV |
| Security | `/server/administration/security/{auth,tokens,ratelimit,firewall,allowlist}` | Security and access-control section pages |
| Disclosure reports | `/server/administration/security/reports` and `/server/administration/security/reports/{tracking_id}` | Filter and sort encrypted researcher reports; view decrypted details in an authenticated, audited session; update status, severity, researcher comment, and expected disclosure date |
| Network | `/server/administration/network/{tor,i2p,geoip,blocklists}` | Overlay and network section pages |
| User operations | `/server/administration/moderation/users`, `/server/administration/users/invites` | Moderation and invitation section pages |
| Scheduler | `/server/administration/scheduler` | View, edit, and trigger scheduled tasks |
| Backup | `/server/administration/backup` | View retention and archives, trigger a scheduled backup |
| Maintenance | `/server/administration/maintenance` | Show maintenance-mode availability |
| Updates | `/server/administration/updates` | Review status and change the persisted release channel |
| Cluster | `/server/administration/nodes`, `/server/administration/nodes/add` | Node and enrollment section pages |
| Metrics | `/server/administration/metrics` | System, build, node, security, and API measurements |
| Server info | `/server/administration/info` | Runtime, version, process, and database driver details |
| Pages | `/server/administration/pages/help` | Edit the MCP help page's project guidance |
| Users | `/server/administration/users` | View accounts and update an account's email or password |
| Agents | `/server/administration/agents`, `/server/administration/agents/add`, `/server/administration/agents/remove` | Create scoped agent credentials, review reports, rotate credentials, and confirm removals |

The admin path segment is configurable with `server.admin_path`. Guests are
redirected to `/server/auth/login`; authenticated regular users are redirected
to `/users`. Administrators use a separate account table and session cookie.
The admin panel is still incomplete. Branding, scheduler, backup, and update
pages expose server-rendered actions; the other listed section routes currently
provide status or overview content and do not implement the full management
flows required by AI.md.

The audit log page is `/server/administration/logs/audit`. It supports
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
tail checkpoint to detect edits to its final record; the missing checkpoint
requirement in AI.md is recorded in `~/Projects/github/webappsgo/TODO.SERVER.md`.

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

The browser sign-in page is `/server/auth/login`. It authenticates server
administrators first and, when regular users are enabled, regular users next.
Administrator credentials are stored in the `admins` table; regular-user
registration never grants server-administrator access. On first startup, the
server prints a one-time setup token directly to the console and serves
`/server/{admin_path}/setup`; it does not persist the plaintext token in a log
file. That page creates the primary administrator with the default username
`steward` (`administrator` is reserved), and stores only the token hash in
`{config_dir}/setup_token.txt`. Its no-JavaScript form submits to a
CSRF-protected server-rendered route. The current setup flow does not yet
implement all setup wizard steps required by AI.md.
The setup response displays the initial `adm_` API token once; the database
stores only its hash.

The Agents page is also available to signed-in users at `/users/agents` for
their own agents. Admin and user agent management APIs are described in the
[Agent guide](agents.md). Organization agent management is not currently
registered.

Successful browser logins set role-specific HTTP-only cookies. The API login
route is `POST /api/{api_version}/server/auth/login`; it returns a revocable,
role-scoped `adm_` or `usr_` bearer API token and sets no cookie. First-run
setup displays the initial administrator token once. The administrator cookie is
`admin_session` (30-day default); the regular-user cookie is `user_session`
(7-day default). Admin API routes require an administrator bearer token.

## Current settings page

The settings page is read-only and reflects only fields currently carried by
the server config model. It is not a full editor for application configuration,
and settings do not generally hot-reload. The metrics and server information
pages display live collector values and process metadata. Consult
[Configuration](configuration.md) for the current configuration source and
supported keys.

The Pages form currently edits only the Markdown project guidance on the MCP
help page. It is saved under `server.pages.help.mcp` in
`{config_dir}/yaml.d/pages.yml`, limited to 64 KiB, and rendered through
Goldmark and the bluemonday allow-list policy with raw HTML disabled.

The API user-management routes are also scoped below the configured admin path.
The public `/api/{api_version}/users` endpoint addresses only the authenticated
caller's profile, shown on `/users`.

## Admin API routes

The following API route groups apply administrator checks in the current
router:

- `/api/v1/server/{admin_path}/users/**`
- `GET /api/v1/nodes` and node CRUD/drain routes
- `POST /api/v1/nodes/token`
- `GET` and `PUT /api/v1/security/config`
- `/api/v1/audit/status`, `/api/v1/audit/config`, and
  `/api/v1/audit/cleanup`
- `/api/v1/compliance/**`
- `/metrics`, `/metrics/json`, and `/api/v1/metrics/**`

See [API Reference](api.md) for methods and route details. API behavior and
authorization are still being completed; do not treat the presence of an admin
page or route as proof that every server-management function is implemented.

## Roles

The current user model has an `is_admin` flag. The complete product design has
additional tenant isolation requirements that are not yet fully implemented.
Administrators should not be assumed to have safe access boundaries over
other users' projects, credentials, or build data until the relevant
multi-user work is complete.
