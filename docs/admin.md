# Admin panel

The current web admin surface has six registered pages:

| Page | Default path | Purpose |
|------|-------------|---------|
| Dashboard | `/server/administration` | Server overview |
| Settings | `/server/administration/settings` | Read the settings currently exposed by the server configuration model |
| Metrics | `/server/administration/metrics` | System, build, node, security, and API measurements |
| Server info | `/server/administration/info` | Runtime, version, process, and database driver details |
| Pages | `/server/administration/pages/help` | Edit the MCP help page's project guidance |
| Users | `/server/administration/users` | View accounts and update an account's email or password |

The admin path segment is configurable with `server.admin_path`. Guests are
redirected to `/server/auth/login`; authenticated regular users are redirected
to `/users`. Administrators use a separate account table and session cookie.
The admin page set is not yet the complete settings and operations panel
described by AI.md.

## Sign in

The browser sign-in page is `/server/auth/login`. It authenticates server
administrators first and, when regular users are enabled, regular users next.
Administrator credentials are stored in the `admins` table; regular-user
registration never grants server-administrator access. On first startup, the
server prints a one-time setup token and serves `/server/{admin_path}/setup`.
That page creates the primary administrator and stores only the token hash in
`{config_dir}/setup_token.txt`. The current setup flow does not yet implement
all setup wizard steps required by AI.md.
The setup response displays the initial `adm_` API token once; the database
stores only its hash.

Successful browser logins set role-specific HTTP-only cookies. The API login
route is `POST /api/{api_version}/server/auth/login`; it sets the matching
cookie. Regular-user login also returns a bearer JWT. Administrator API
requests use an `adm_` API token; first-run setup displays the initial token
once. The administrator cookie is
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
