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
      require_email_verification: true
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
    allowlist:
      - cidr: 10.0.0.0/8
        description: "Internal network"
      - cidr: "2001:db8::/48"
        description: "Office IPv6 prefix"
    account_lockout:
      soft_lock_attempts: 5
      soft_lock_window: 15m
      soft_lock_duration: 15m
      hard_lock_attempts: 10
      hard_lock_window: 1h
      hard_lock_duration: 1h
      permanent_lock_attempts: 15
      permanent_lock_window: 24h
  contact:
    admin:
      email: ""
    security:
      email: "security@casci.example.test"
    abuse:
      email: ""
    dpo:
      name: ""
      email: ""
    general:
      email: ""
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
    expires: ""
  reports:
    rate_limit_per_minute: 60
    rate_limit_per_ip_burst: 10
  csp:
    reports_sample_rate: 1.0
privacy:
  hash_report_ips: true
```

## Page content and contact CAPTCHA

The server rendered Pages editor stores About, Privacy, Terms, and Footer
content under `server.branding.content` in `yaml.d/branding.yml`. Help content
and Contact settings live under `server.pages` in `yaml.d/pages.yml`. Footer
content must use `format: html`; its editor sanitizes the fragment before
preview and display. The older `server.footer.custom_html` value remains a
fallback when there is no `server.branding.content.footer` entry.

Contact CAPTCHA defaults to the built in simple challenge. `recaptcha` and
`hcaptcha` require both `captcha_site_key` and `captcha_secret_key`; invalid or
incomplete provider settings log a warning and use the simple challenge. The
site key is stored in `pages.yml`. The secret is encrypted in `secrets.yml`,
and the admin API and page never return it.

```yaml
# yaml.d/branding.yml
server:
  branding:
    content:
      footer:
        format: html
        body: '<p>Powered by <strong>Example</strong></p>'
```

```yaml
# yaml.d/pages.yml
server:
  pages:
    contact:
      captcha: hcaptcha
      captcha_site_key: public-site-key
      antispam:
        honeypot: true
        min_fill_seconds: 3
        max_age_seconds: 43200
        per_ip: 5
        per_email: 3
        global: 60
        window: 1h
        max_links: 10
        reject_markup: true
        blocked_words: []
        blocked_email_domains: []
        strikes: 3
        ban_hours: 24
        review_min_seconds: 2
        review_max_seconds: 900
        duplicate_window_hours: 24
```

Contact recipients are configured under `server.contact` in `server.yml`.
Security defaults to `security@{fqdn}` when omitted; explicitly setting it to
an empty string suppresses the public security mailbox. Abuse falls back to
the general address for public display; the admin address is never published.
DPO display falls back to the general address, while internal report routing
can fall back to the admin recipient. Webhook URLs are private and must not be
exposed in public responses.

Abuse, DMCA, and privacy submissions are reviewed and confirmed before
delivery. Configure outbound SMTP under `server.email` for delivery; each
accepted submission sends one plain-text notification, with the submitter's
address in `Reply-To` and no automatic reply. Abuse and DMCA route to the abuse
role, then general, then admin. Privacy requests route to DPO, then general,
then admin. The security form follows the encrypted security-report workflow
and requires a valid rotating `security_id` when opened from the security
policy page.

The contact antispam settings shown above default to a three-second minimum
fill time, 12-hour form-token lifetime, limits of 5 per IP, 3 per email, and
60 total per hour, a ten-link maximum, markup rejection, a three-strike
24-hour ban, a two-second to 15-minute review window, and a 24-hour duplicate
window. Settings are enforced server-side. Tor and I2P requests skip the
per-IP limit and strike bans; per-email and global limits still apply, and
those requests use the built-in simple CAPTCHA.

## External identity providers

Provider definitions live in `server.auth.oidc.providers`,
`server.auth.ldap.providers`, or `server.auth.saml.providers` in
`server.yml`. Each entry has a unique `name`, `display_name`, and
`enabled` flag. Provider secrets are moved to encrypted, name-keyed entries
in `secrets.yml` and merged at startup; do not put secret values in source
control.

An OIDC provider needs an issuer, client ID, and client secret. It uses
authorization code flow with PKCE and the configured `openid`, `profile`,
and `email` scopes by default. LDAP needs a server, bind DN, bind password,
base DN, user filter, and explicit TLS settings. SAML needs IdP metadata and
SP entity/ACS settings; the server can generate SP signing credentials and
stores the private key encrypted. Admin group and role mappings are provider
settings and should be configured explicitly before granting external users
administrator access. The admin panel has provider-specific list, edit, test,
and delete forms; its configuration editor accepts the provider JSON
structure.

See [Integrations](integrations.md) for sign-in routes and protocol behavior.
Secrets are never included in provider list/read responses.

SMTP passwords are stored encrypted in `secrets.yml`; they are omitted from
`server.yml`. `SMTP_*` environment variables override file values at startup,
including the sender name and address. The defaults are port 587, automatic TLS
mode, the application name as the sender name, and `no-reply@{fqdn}` as the
sender address when a public FQDN is configured. At startup, the server tests
configured SMTP or autodetects a local SMTP endpoint with an eight-second total
bound. Email-dependent features remain disabled unless the connection check
succeeds. Admins can update and test SMTP settings at
`/server/{admin_path}/email`.

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
submission. In private mode, public self-registration is unavailable;
invitations remain available from the admin panel, and existing users can
still sign in. `require_email_verification` defaults to true; it gates
registration only when working SMTP is enabled and sends a one-time link before
the new account can sign in.

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
