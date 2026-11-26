# Installation

CASCI can be run from the provided container image or built as static Go
binaries. `make build` produces the server and companion `casci-cli` client.
The repository also contains service installer scripts; service management
flags in the server binary remain incomplete.

## Docker Compose

The production Compose file binds the application to the Docker bridge at
`172.17.0.1:64580` and starts a Valkey service. It mounts `./volumes/config`
and `./volumes/data` from the directory where Compose runs.

```bash
curl -q -LSsf -o docker-compose.yml \
  https://raw.githubusercontent.com/webappsgo/casci/main/docker/docker-compose.yml
docker compose -f docker-compose.yml up -d
```

The container listens on port `80`. The included health check runs
`casci --status`. The Compose file does not configure an external SQL database;
SQLite is the current default.

For development or tests, inspect `docker/docker-compose.dev.yml` and
`docker/docker-compose.test.yml`. Those files are project tooling, not
production configuration.

## Build and run a binary

All Go builds use the Docker toolchain configured in the Makefile.

```bash
make dev
```

`make dev` builds the current-platform binaries into an OS temporary directory
and prints the output paths. `make build` builds the server and CLI for Linux,
macOS, Windows, and FreeBSD on amd64 and arm64 into `binaries/`.

Run the resulting binary with no arguments to start the server. On first run,
CASCI creates its configuration and data directories and writes `server.yml`.
The default native port is selected from `64000–64999` and persisted. The
container port is `80`.

## Verify a running instance

```bash
curl -q -LSsf https://casci.casapps.us/server/healthz
```

The JSON API health endpoint is
`https://casci.casapps.us/api/v1/server/healthz`. `casci --status` checks the
configured local server address and port; it exits `0` when the health check
succeeds and non-zero when it fails.

## Linux systemd

The repository provides `scripts/install-systemd.sh` and
`scripts/casci.service`. Build or obtain the `casci` binary first, then place it
in the installer’s expected working directory and review the script before
running it as root:

```bash
sudo ./scripts/install-systemd.sh
sudo systemctl enable --now casci
```

The installer creates a service account, copies the binary to
`/usr/local/bin/casci`, installs the unit, and creates data/log paths. The
installer's hard-coded paths may differ from the application path resolver;
review the unit and configuration before using it on an existing host.

## macOS launchd and Windows services

The repository includes `scripts/install-launchd.sh`,
`scripts/install-windows.ps1`, service definitions, and an uninstall helper.
Review each script and its target paths before running it with administrative
privileges. These scripts have not been validated as a complete cross-platform
service lifecycle.

## Next steps

- Read [Configuration](configuration.md) and set a private CORS allowlist
  before exposing the service.
- Review [Security](security.md), especially build execution and user
  isolation limitations.
- Use [CLI Reference](cli.md) for server flags and client commands.
