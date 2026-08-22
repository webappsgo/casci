# Installation

## Docker (Recommended)

```bash
docker run -d \
  --name casci \
  -p 172.17.0.1:64580:80 \
  -v ./volumes/config:/config:z \
  -v ./volumes/data:/data:z \
  ghcr.io/webappsgo/casci:latest
```

Inside a container the listen port defaults to `80`. Override it with the `PORT` environment variable, and remember to update the published port mapping to match.

### Docker Compose

```bash
curl -q -LSsf -O https://raw.githubusercontent.com/webappsgo/casci/main/docker/docker-compose.yml
docker compose up -d
```

The shipped compose file binds `172.17.0.1:64580` on the Docker bridge only, leaving external exposure to a reverse proxy. It also runs a Valkey cache container and uses `casci --status` as the container health check.

### Docker with an External Database

```yaml
services:
  casci:
    image: ghcr.io/webappsgo/casci:latest
    environment:
      PORT: 80
      DATABASE_DRIVER: postgres
      DATABASE_URL: postgres://casci:casci@postgres:5432/casci?sslmode=disable
    ports:
      - "172.17.0.1:64580:80"
    volumes:
      - ./volumes/config:/config:z
      - ./volumes/data:/data:z
```

## Binary

Download from [releases](https://github.com/webappsgo/casci/releases):

```bash
# Linux AMD64
curl -q -LSsf -O https://github.com/webappsgo/casci/releases/latest/download/casci-linux-amd64
chmod +x casci-linux-amd64
./casci-linux-amd64
```

CASCI requires no flags. On first run it creates its directories, chooses a random port in the `64000-64999` range, and persists that port to `server.yml`. The startup banner prints the selected address and port; `casci --status` reports it afterwards.

## Verify the Installation

```bash
# Health check against a running instance
curl -q -LSsf https://casci.casapps.us/server/healthz

# Local status check (exit 0 = healthy, 1 = unhealthy)
casci --status
```

## Systemd Service

CASCI ships a unit file and installer under `scripts/`.

```bash
sudo ./scripts/install-systemd.sh
sudo systemctl enable --now casci
```

The installer creates the `casci` system user and group, installs the binary to `/usr/local/bin/casci`, creates `/var/lib/casapps/casci/` and `/var/log/casapps/casci/`, and installs the unit.

### Service Management

```bash
sudo systemctl start casci
sudo systemctl stop casci
sudo systemctl restart casci
sudo systemctl status casci
sudo journalctl -u casci -f
```

### Unit Configuration

Environment variables are unprefixed. Add them to a drop-in rather than editing the shipped unit:

```bash
sudo systemctl edit casci
```

```ini
[Service]
Environment="DATABASE_DRIVER=postgres"
Environment="DATABASE_URL=postgres://user:pass@localhost/casci"
Environment="PORT=64580"
```

```bash
sudo systemctl daemon-reload
sudo systemctl restart casci
```

`DATABASE_DRIVER` and `DATABASE_URL` are re-read on every start, so a drop-in change takes effect on restart. `PORT` is init-only: it seeds the port on first run and is ignored afterwards. To change the port on an existing install, edit `server.port` in `server.yml` or use the admin panel. Settings that are not environment variables at all — TLS, CORS, the encryption key, the webhook secret, footer options, and database pooling — live in `server.yml`; see [Configuration](configuration.md).

The unit runs as the non-privileged `casci` user with `NoNewPrivileges`, `PrivateTmp`, and `ProtectSystem` hardening, and grants write access only to the data and log directories.

## macOS (launchd)

```bash
sudo ./scripts/install-launchd.sh
```

The installer creates the hidden `_casci` system user and group with a system-range UID/GID, installs the binary to `/usr/local/bin/casci`, and loads `/Library/LaunchDaemons/com.casapps.casci.plist`.

### Service Management

```bash
sudo launchctl start com.casapps.casci
sudo launchctl stop com.casapps.casci
sudo launchctl list | grep -- casci
sudo launchctl unload /Library/LaunchDaemons/com.casapps.casci.plist
sudo launchctl load /Library/LaunchDaemons/com.casapps.casci.plist
```

### Plist Configuration

```xml
<key>EnvironmentVariables</key>
<dict>
    <key>DATABASE_DRIVER</key>
    <string>postgres</string>
    <key>DATABASE_URL</key>
    <string>postgres://user:pass@localhost/casci</string>
</dict>
```

Reload the daemon after editing:

```bash
sudo launchctl unload /Library/LaunchDaemons/com.casapps.casci.plist
sudo launchctl load /Library/LaunchDaemons/com.casapps.casci.plist
```

## Windows

Run PowerShell as Administrator:

```powershell
.\scripts\install-windows.ps1
```

The installer registers the `NT SERVICE\CASCI` virtual service account, installs the binary to `C:\Program Files\CASCI\casci.exe`, creates `C:\ProgramData\casapps\casci\`, and configures automatic startup with restart-on-failure recovery.

### Service Management

```powershell
Start-Service -Name CASCI
Stop-Service -Name CASCI
Restart-Service -Name CASCI
Get-Service -Name CASCI
Get-EventLog -LogName Application -Source CASCI -Newest 50
Set-Service -Name CASCI -StartupType Automatic
```

### Service Configuration

```powershell
$envPath = "HKLM:\SYSTEM\CurrentControlSet\Services\CASCI"

Set-ItemProperty -Path $envPath -Name "Environment" -Value @(
    "DATABASE_DRIVER=postgres",
    "DATABASE_URL=postgres://user:pass@localhost/casci"
)

Restart-Service -Name CASCI
```

## Uninstallation

### Linux / macOS

```bash
sudo ./scripts/uninstall-service.sh
```

The script stops and disables the service, removes the service definition, and prompts before removing the binary and data directories.

### Windows

```powershell
Stop-Service -Name CASCI -Force
sc.exe delete CASCI
Remove-Item -Recurse -Force "$env:ProgramFiles\CASCI"
```

Data under `C:\ProgramData\casapps\casci\` is left in place; remove it manually if you no longer need it.

## Directory Layout

| Purpose | Linux/BSD (root) | Linux/BSD (user) |
|---------|------------------|------------------|
| Config | `/etc/casapps/casci/` | `~/.config/casapps/casci/` |
| Data | `/var/lib/casapps/casci/` | `~/.local/share/casapps/casci/` |
| Cache | `/var/cache/casapps/casci/` | `~/.cache/casapps/casci/` |
| Logs | `/var/log/casapps/casci/` | `~/.local/log/casapps/casci/` |
| Backups | `/mnt/Backups/casapps/casci/` | `~/.local/share/Backups/casapps/casci/` |
| PID file | `/var/run/casapps/casci.pid` | `~/.local/share/casapps/casci/casci.pid` |

The config file is `server.yml` inside the config directory. See [Configuration](configuration.md) for platform-specific paths on macOS and Windows and for every available setting.

Containers do not create a PID file; the container runtime supervises the process.

## Troubleshooting

### Service will not start

```bash
sudo journalctl -u casci -n 50 --no-pager
sudo systemctl status casci
ls -l /usr/local/bin/casci
ls -ld /var/lib/casapps/casci
```

On macOS, validate the plist and read the daemon log:

```bash
plutil /Library/LaunchDaemons/com.casapps.casci.plist
tail -50 /var/log/casapps/casci/stderr.log
```

On Windows:

```powershell
Get-EventLog -LogName Application -Source CASCI -Newest 20
Get-Service -Name CASCI | Format-List *
Test-Path "$env:ProgramFiles\CASCI\casci.exe"
```

### Port already in use

CASCI persists its chosen port in `server.yml`. Set a different one and restart:

```bash
PORT=64581 ./casci-linux-amd64
```

Or edit `server.yml` directly and restart the service.

### Docker socket permission denied

Container-based build execution needs access to the Docker socket. Mount it and ensure the runtime user can read it:

```bash
docker run -d \
  --name casci \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/webappsgo/casci:latest
```

### Database connection errors

Verify `DATABASE_DRIVER` and `DATABASE_URL`. With the default `sqlite` driver, CASCI writes to `server.db` in the database directory and needs that directory to be writable.

## Next Steps

- [Configuration](configuration.md) - every setting and environment variable
- [Admin Panel](admin.md) - first-run setup and administration
- [API Reference](api.md) - REST and GraphQL surfaces
