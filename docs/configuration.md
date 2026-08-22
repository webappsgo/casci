# Configuration

CASCI works with zero configuration. Everything below is optional; the defaults are chosen so a fresh binary starts and serves without any operator input.

## Config File

The config file is always named `server.yml` and lives in the config directory.

| Platform | Privileged | Unprivileged |
|----------|------------|--------------|
| Linux | `/etc/casapps/casci/server.yml` | `~/.config/casapps/casci/server.yml` |
| FreeBSD / OpenBSD / NetBSD | `/usr/local/etc/casapps/casci/server.yml` | `~/.config/casapps/casci/server.yml` |
| macOS | `/Library/Application Support/casapps/casci/server.yml` | `~/Library/Application Support/casapps/casci/server.yml` |
| Windows | `%ProgramData%\casapps\casci\server.yml` | `%APPDATA%\casapps\casci\server.yml` |

Whether the privileged or unprivileged path is used is decided once at process start from the effective UID, before any privilege drop, and stays fixed for the lifetime of the process.

`server.yml` is the persisted source of truth. It is generated on first run and written with mode `0600`, because it holds the credential-vault encryption key and the webhook secret.

There are two top-level blocks, `server:` and `web:`.

```yaml
server:
  # Chosen at random in the 64000-64999 range on first run and written back here
  port: 64237
  # Public FQDN; seeded from DOMAIN
  fqdn: ""
  # Listen address; IPv6 wildcard, which also accepts IPv4 on a dual-stack host
  address: "[::]"
  mode: production

  # The single path segment the admin panel is mounted under, below /server
  admin_path: administration
  # The version segment of every /api/{api_version}/ route
  api_version: v1

  healthz:
    root:
      # Mounts /healthz as an alias of /server/healthz; off by default
      enabled: false

  branding:
    title: casci
    tagline: ""

  ssl:
    enabled: false
    cert: /etc/casapps/casci/ssl/local/server.crt
    key: /etc/casapps/casci/ssl/local/server.key
    min_version: TLS1.2

  security:
    # Base64-encoded 256-bit key, generated on first run
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
    # Defaults to server.db inside the database directory
    url: ""
    max_conns: 25
    max_idle: 5
    replica_urls: []

web:
  # A single origin, a comma-separated list, or "*"
  cors: "*"

  security:
    # RFC 9116 security.txt fields
    report_url: https://github.com/webappsgo/casci/security/advisories/new
    contact: ""
    # RFC 3339; regenerated one year out whenever it is missing or expired
    expires: ""
```

Most operators never edit this file by hand. Everything in it is editable from the admin panel at `/server/administration`, or at whatever `server.admin_path` is set to.

## Directories

| Purpose | Linux/BSD (root) | Linux/BSD (user) |
|---------|------------------|------------------|
| Config | `/etc/casapps/casci/` | `~/.config/casapps/casci/` |
| Data | `/var/lib/casapps/casci/` | `~/.local/share/casapps/casci/` |
| Cache | `/var/cache/casapps/casci/` | `~/.cache/casapps/casci/` |
| Logs | `/var/log/casapps/casci/` | `~/.local/log/casapps/casci/` |
| Backups | `/mnt/Backups/casapps/casci/` | `~/.local/share/Backups/casapps/casci/` |
| PID file | `/var/run/casapps/casci.pid` | `~/.local/share/casapps/casci/casci.pid` |

Every directory is created on demand if missing. The backup directory prefers the system backup path when it is writable and otherwise falls back to a `backup/` subdirectory of the data directory.

Containers do not write a PID file. The container runtime supervises the process, and PIDs are namespace-local, so a PID file on a mounted volume would point at the wrong process.

## Listen Port

CASCI has no fixed default port.

| Situation | Port |
|-----------|------|
| First run, native install | Random port in `64000-64999`, written to `server.yml` |
| Subsequent runs, native install | The port persisted in `server.yml` |
| Inside a container | `80` |
| First run with `PORT` set | The value of `PORT`, which must be `1-65535`, persisted to `server.yml` |

The port is generated once and then persisted, so a native install keeps the same port across restarts and upgrades. `PORT` only seeds the value on first run; on later starts it is ignored. Change the port afterwards with `--port`, from the admin panel, or by editing `server.port` in `server.yml`.

The listen address defaults to `[::]`, the IPv6 wildcard, which also accepts IPv4 connections on a dual-stack host. Override it with `LISTEN` on first run, `--address` at any time, or `server.address` in `server.yml`.

## Environment Variables

Environment variables are unprefixed. There is no `CASCI_` prefix.

The list below is closed. Anything not on it is not an environment variable — it is a `server.yml` key, and setting it in the environment has no effect.

### Runtime Variables

Read on every start.

| Variable | Description | Default |
|----------|-------------|---------|
| `NO_COLOR` | Disable ANSI color and emoji output when set and non-empty | unset |
| `TERM` | `TERM=dumb` disables all ANSI escapes and forces CLI mode | inherited |
| `DOMAIN` | Public FQDN, persisted to `server.fqdn`. It does **not** feed the CORS allowlist | unset |
| `MODE` | `production` (`prod`), `development` (`dev`, `devel`), or `debug` | `production` |
| `DATABASE_DRIVER` | `sqlite`, `postgres`, `postgresql`, `mysql`, or `mariadb` | `sqlite` |
| `DATABASE_URL` | Driver-specific connection string | `server.db` in the database directory |

There are no `SMTP_*` environment variables. Mail settings belong to an individual notification config — see [Integrations](integrations.md).

### Init-Only Variables

Read once during first run, persisted to `server.yml`, and ignored on every later start. To change one of these afterwards, edit `server.yml` or use the admin panel; re-exporting the variable will not do anything.

| Variable | Description | Persisted as |
|----------|-------------|--------------|
| `CONFIG_DIR` | Configuration directory | path resolution |
| `DATA_DIR` | Data directory | path resolution |
| `CACHE_DIR` | Cache directory | path resolution |
| `LOG_DIR` | Log directory | path resolution |
| `DATABASE_DIR` | SQLite database directory | path resolution |
| `BACKUP_DIR` | Backup directory | path resolution |
| `PORT` | Listen port | `server.port` |
| `LISTEN` | Listen address | `server.address` |
| `APPLICATION_NAME` | Application title | `server.branding.title` |
| `APPLICATION_TAGLINE` | Application description | `server.branding.tagline` |

The directory variables are re-read on every start, but they only steer path resolution — they are never written into `server.yml`. The rest seed a value once and are then ignored.

### Settings That Are Not Environment Variables

These are configured in `server.yml` or from the admin panel. There is no environment variable for any of them.

| Setting | `server.yml` key |
|---------|------------------|
| Allowed CORS origins | `web.cors` |
| Admin panel path segment | `server.admin_path` |
| API version path segment | `server.api_version` |
| Root `/healthz` alias | `server.healthz.root.enabled` |
| Enable direct HTTPS | `server.ssl.enabled` |
| TLS certificate path | `server.ssl.cert` |
| TLS key path | `server.ssl.key` |
| Minimum TLS version | `server.ssl.min_version` |
| Vulnerability report URL | `web.security.report_url` |
| Security contact address | `web.security.contact` |
| `security.txt` expiry stamp | `web.security.expires` |
| Credential vault master key | `server.security.encryption_key` |
| Encryption key version | `server.security.encryption_key_version` |
| Webhook shared secret | `server.security.webhook_secret` |
| Analytics tracking ID | `server.footer.tracking_id` |
| Custom footer HTML | `server.footer.custom_html` |
| Cookie consent banner | `server.footer.cookie_consent` |
| Consent banner text | `server.footer.cookie_message` |
| Consent policy link | `server.footer.cookie_policy_url` |
| Maximum open DB connections | `server.database.max_conns` |
| Maximum idle DB connections | `server.database.max_idle` |
| Read-replica connection strings | `server.database.replica_urls` |

## CORS

`web.cors` is a single origin, a comma-separated list of origins, or `*`.

!!! danger "The generated default is `*`"
    On first run `web.cors` is written as `*`, so every response carries `Access-Control-Allow-Origin: *` and any website can issue cross-origin requests to your instance. Set it to your own origins before exposing the instance.

`DOMAIN` does not feed this value; the two are unrelated. Only an explicit, matched origin receives `Access-Control-Allow-Credentials: true` — the wildcard never does. See [Security](security.md) for the full behavior table.

## Encryption Key

The credential vault and other server-encrypted data are protected by a master key stored at `server.security.encryption_key` in `server.yml`. There is no separate `encryption.key` file and no `ENCRYPTION_KEY` environment variable.

On first run, if the key is empty, 32 bytes are read from `crypto/rand`, base64-encoded, written to `server.yml`, and `server.security.encryption_key_version` is set to `1`. The version increments on every rotation. This is why `server.yml` is written with mode `0600`.

The key is validated at startup: it must be valid base64 decoding to exactly 32 bytes. An invalid key is a fatal startup error rather than a silent fallback to something predictable, and failure to source real entropy is likewise fatal.

Losing the key makes every stored credential unrecoverable and forces every user to re-enroll 2FA. Back up `server.yml` alongside the database and treat it with the same care as a private key.

## Databases

| Driver | Connection string |
|--------|-------------------|
| `sqlite` | Path to the database file; defaults to `server.db` in the database directory |
| `postgres`, `postgresql` | `postgres://user:pass@host:5432/casci?sslmode=disable` |
| `mysql`, `mariadb` | `user:pass@tcp(host:3306)/casci?parseTime=true` |

Those are the only accepted values. Anything else is a fatal startup error.

`DATABASE_DRIVER` and `DATABASE_URL` are runtime environment variables and are re-read on every start. Connection pooling and read replicas are not: set `server.database.max_conns`, `server.database.max_idle`, and `server.database.replica_urls` in `server.yml`. Replica URLs are a list of connection strings using the same driver.

## TLS

Set `server.ssl.enabled` to `true` and point `server.ssl.cert` and `server.ssl.key` at your certificate and key. Both default to a `ssl/local/` path under the config directory. `server.ssl.min_version` pins the floor and defaults to `TLS1.2`.

When TLS is terminated by a reverse proxy, leave `server.ssl.enabled` off.

## Reverse Proxy

!!! warning "Forwarding headers are deliberately ignored"
    CASCI does not read `X-Forwarded-Proto`, `X-Forwarded-For`, `X-Forwarded-Port`, `X-Forwarded-Prefix`, or `X-Real-IP`. Honoring them safely requires a trusted-proxy allowlist, which this build does not have, so trusting them would let any client rewrite the URLs the server generates.

An absolute URL is built from the request's own `Host` header and from whether the connection itself is TLS. Two consequences follow:

- Serve CASCI at the root of its own hostname. There is no base-path or subdirectory support.
- Set `DOMAIN` (or `server.fqdn`) so the server still knows its public hostname when a request arrives without a `Host` header.

## Well-Known URIs

`/.well-known/security.txt` is generated per request in RFC 9116 form, listing the private vulnerability reporting URL from `web.security.report_url`, a CC mailbox from `web.security.contact` (defaulting to `security@<host>`), a policy link, and the `web.security.expires` stamp. The expiry is rewritten one year out whenever it is missing or already past, because an expired `security.txt` is invalid.

Every other `/.well-known/` path returns `404`. Nothing is redirected.

## Admin Panel

All settings are configurable from the WebUI at `/server/administration`, or at `server.admin_path` if you changed it. See [Admin Panel](admin.md).
