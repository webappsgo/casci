# Configuration

CASCI reads `server.yml` from its resolved configuration directory and reads
area settings from `yaml.d/` files, including MCP help from `pages.yml` and
CLI binary distribution settings from `maintenance.yml`. The current file-backed
configuration holds server, web, database, TLS, branding, footer,
security-contact, encrypted SMTP credentials, privacy/report sampling, encryption-key, webhook-secret, and log-output values. This
implementation does not yet provide the full database-backed configuration
system or hot reload described in the product specification.

## Configuration file

The main blocks are `server` and `web`. A first-run file has defaults filled
in by the config loader. This is a partial example of the current schema:

```yaml
server:
  address: "[::]"
  port: 64580
  mode: production
  admin_path: administration
  api_version: v1
  users:
    enabled: false
    registration:
      mode: open
  healthz:
    root:
      enabled: false
  branding:
    title: casci
    tagline: ""
  ssl:
    enabled: false
    cert: ""
    key: ""
    min_version: TLS1.2
  security:
    encryption_key: ""
    encryption_key_version: 1
    webhook_secret: ""
  footer:
    tracking_id: ""
    custom_html: ""
    cookie_consent: true
    cookie_message: "This site uses cookies for functionality and analytics."
    cookie_policy_url: /server/privacy
  database:
    driver: sqlite
    url: ""
    max_conns: 25
    max_idle: 5
    replica_urls: []
  notifications:
    email:
      smtp:
        host: ""
        port: 587
        username: ""
        tls: auto
      from:
        name: casci
        email: ""
web:
  cors: "*"
  robots:
    allow: ["/", "/api"]
    deny: ["/server/administration"]
  security:
    report_url: https://github.com/webappsgo/casci/security/advisories/new
    contact: ""
    expires: ""
  reports:
    rate_limit_per_minute: 60
    rate_limit_per_ip_burst: 10
  csp:
    reports_sample_rate: 1.0
privacy:
  hash_report_ips: true
```

SMTP passwords are stored encrypted in `secrets.yml`; they are omitted from
`server.yml`. `SMTP_*` environment variables override file values at startup,
including the sender name and address. The defaults are port 587, automatic TLS
mode, the application name as the sender name, and `no-reply@{fqdn}` as the
sender address when a public FQDN is configured. Email remains disabled until
the SMTP connection check succeeds.

TLS mode `auto` uses implicit TLS on port 465 and otherwise starts in plain SMTP
and upgrades with STARTTLS when the server offers it. `starttls` requires that
upgrade; `tls` uses implicit TLS on any port; `none` explicitly uses plaintext.
Authentication is never sent without TLS. Certificates are verified normally
for hostnames and public IPs; only configured private IP literals use the
documented relaxed certificate policy. `web.reports.rate_limit_per_minute`
defaults to 60 per client IP, `web.reports.rate_limit_per_ip_burst` defaults to
10, and `web.csp.reports_sample_rate` defaults to 1.0. Report IP hashing is
enabled by default and can be changed with `privacy.hash_report_ips`.

The example omits generated key material. Protect `server.yml` because it can
contain the encryption and webhook secrets. Do not commit real configuration
or credentials.

## Logging

server.logs.level accepts debug, info, warn, or error; when empty it follows
the mode (info in production and debug in development/debug).
server.logs.defaults.rotate and server.logs.defaults.keep apply to each log
kind unless that kind overrides them. Rotation accepts a time interval, a size
limit such as 50MB, or both (for example weekly,50MB). Retention accepts none,
an archive count, a duration such as 14d, or forever.

The supported file kinds are access, server, error, audit, security, and
debug. Each has an enabled switch, format, rotation, and retention. audit
supports JSON only and can optionally compress retained archives. debug is
disabled by default. Access formats are Apache, Nginx, JSON, or custom;
server/error/debug formats are text or JSON; security formats are fail2ban,
syslog, CEF, JSON, or text. custom access output uses the format string from
server.logs.defaults.custom or server.logs.access.custom.

Server log files use ASCII output. Successful health checks and successful
--status polls are omitted from request logs. Audit records are written to the
authoritative audit.log; database rows are its best-effort query mirror.

Regular-user accounts are disabled by default. Set `server.users.enabled` to
`true` to enable them. `server.users.registration.mode` accepts `open` or
`private` and defaults to `open`. In open mode, `/server/auth/register` renders
the public registration form and creates a signed-in account after successful
submission. In private mode, public self-registration is unavailable; existing
users can still sign in.

## Directories and environment

The path resolver supports `CONFIG_DIR`, `DATA_DIR`, `CACHE_DIR`, `LOG_DIR`,
`DATABASE_DIR`, and `BACKUP_DIR`. Defaults vary by operating system and
privilege level. The flags `--config`, `--data`, `--cache`, `--log`, and
`--backup` are also accepted by the CLI.

Current configuration-related environment variables:

| Variable | Effect |
|----------|--------|
| `LISTEN` | Seeds `server.address` when the file has no address |
| `PORT` | Seeds the port on first run; the generated port is persisted |
| `APPLICATION_NAME` | Seeds `server.branding.title` |
| `APPLICATION_TAGLINE` | Seeds `server.branding.tagline` |
| `MODE` | Overrides the configured mode at startup |
| `DEBUG` | Truthy values enable debug profiling and detailed logging; `DEBUG=false` overrides the `debug` mode default |
| `DOMAIN` | Overrides the configured public FQDN at startup |
| `DATABASE_DRIVER` | Overrides the database driver at startup |
| `DATABASE_URL` | Overrides the database connection string at startup |
| `SMTP_HOST` | Overrides `server.notifications.email.smtp.host` |
| `SMTP_PORT` | Overrides the SMTP port (default `587`) |
| `SMTP_USERNAME` | Overrides the SMTP username |
| `SMTP_PASSWORD` | Overrides the SMTP password without persisting it |
| `SMTP_TLS` | Overrides TLS mode (`auto`, `starttls`, `tls`, or `none`) |
| `SMTP_FROM_NAME` | Overrides the sender display name |
| `SMTP_FROM_EMAIL` | Overrides the sender address |
| `NO_COLOR` | Disables color and emoji output when non-empty |
| `TERM` | `dumb` disables ANSI and emoji output |

The CLI also accepts `--address`, `--port`, `--mode`, `--debug`, and
`--color`. See [CLI Reference](cli.md). The explicit `--debug` flag overrides
`DEBUG`; a `DEBUG` value overrides the debug default selected by `MODE=debug`.

## Database drivers

The current config loader accepts SQLite, PostgreSQL, and MySQL driver names
(including their documented aliases). SQLite is the default. Connection
pooling and replica fields are part of the config model, but cluster mode and
database-backed configuration are not implemented as a complete system.

Use a credential-safe DSN format for your selected driver. Do not place real
passwords in source-controlled YAML or documentation examples.

## CORS

`web.cors` accepts one origin, a comma-separated origin list, or `*`. The
first-run default is `*`; restrict it to trusted origins before exposing the
instance. MCP performs its own Origin check and does not treat `*` as permission
for arbitrary MCP origins.

## TLS

`server.ssl.enabled` enables the server's TLS listener using `cert` and `key`.
The minimum TLS version is resolved from `server.ssl.min_version`; unsupported
or empty values fall back to TLS 1.2. When TLS terminates at a reverse proxy,
leave direct server TLS disabled. The current code does not manage certificate
issuance or renewal.

## Public URL resolution

The server builds public URLs from the request host and connection scheme. If a
non-empty `DOMAIN` list is configured, only listed request hosts are used; an
unlisted or invalid host falls back to the first configured domain. Forwarded
host headers are ignored because trusted-proxy support is not implemented.
Serve the app at the root of its host rather than a URL subdirectory.

`/robots.txt`, `/sitemap.xml`, `/feed.xml`, and `/rss.xml` are always served; `/favicon.ico` serves the embedded site icon.
Robots rules come from `web.robots.allow` and `web.robots.deny`; the defaults
allow `/` and `/api` and disallow the configured admin path. The sitemap lists
public pages allowed by those rules, with the generation timestamp as `lastmod`. Atom and RSS feeds are valid empty feeds
while this application has no public per-item resource. All four routes support
GET and HEAD, conditional requests, and five-minute public caching.

## Well-known files

`/.well-known/security.txt` is generated from `web.security.report_url`,
`web.security.contact`, and `web.security.expires`. `/.well-known/llms.txt`
and `/llms.txt` are also served. Other paths under `/.well-known/` return
`404` unless implemented by a handler.

## CLI binary downloads

The server publishes only CLI binaries found in `{data_dir}/cli` whose hashes
match the colocated `sha256.txt` file. `version.txt` supplies the release
version. `/api/autodiscover` includes the verified platform, version, and
SHA-256 values; `/cli/binaries/{project}-cli-{os}-{arch}` streams a matching
binary. The server does not build or modify release files.

These settings belong in `{config_dir}/yaml.d/maintenance.yml`:

```yaml
cli:
  binary_download:
    dir: "{data_dir}/cli"
    require_auth: false
```

The default download is public. Set `require_auth: true` to require an
`Authorization: Bearer` token. A relative `dir` is resolved under
`{data_dir}`; the `{data_dir}` placeholder can also be used in an explicit
path.

## Admin settings

The admin settings page is mounted at `/server/{admin_path}/settings` and is
read-only. It displays only fields currently carried by the server config
model; it is not a complete configuration editor.

The admin Pages form is mounted at `/server/{admin_path}/pages/help`. It edits
`server.pages.help.mcp`, the Markdown project guidance shown on the public MCP
help page and in its API response. The key belongs in
`{config_dir}/yaml.d/pages.yml`:

```yaml
server:
  pages:
    help:
      mcp: |
        Operator-maintained Markdown shown in the MCP help page.
```

Markdown is rendered through Goldmark and the bluemonday allow-list policy
with raw HTML disabled. The value is limited to 64 KiB.
