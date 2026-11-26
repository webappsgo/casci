# CLI reference

CASCI ships a server executable (`casci`), an API client (`casci-cli`), and a
build-node agent (`casci-agent`). All three binaries use the same release
version, commit ID, and UTC build timestamp. The client and agent are built by
`make build` for the eight supported OS and architecture combinations.

```text
casci-cli [flags] COMMAND
```

## Global flags

| Flag | Description |
|------|-------------|
| `--server URL` | CASCI server URL; required unless an official site was embedded at build time |
| `--api-version VERSION` | Versioned API path; defaults to `v1` |
| `--token TOKEN` | API token (highest token-source priority) |
| `--token-file FILE` | Read an API token from a private file |
| `--version` | Print the invoked binary name, version, commit ID, and build timestamp |
| `--help`, `-h` | Print usage |

The token source order is `--token`, `--token-file`, `CASCI_TOKEN`, the
`auth.token` value in `cli.yml`, then the standalone `token` file. The default
files are under the current user's configuration directory at
`webappsgo/casci/`. Credential files with group or world permissions are
rejected; use `chmod 0600` to secure them.

## Commands

| Command | Description | Authentication |
|---------|-------------|----------------|
| `status` or `health` | Fetch server health | Public health endpoint |
| `projects list` | List projects owned by the authenticated user | API token required |
| `projects create NAME REPOSITORY_URL [BRANCH]` | Create a project; branch defaults to `main` | API token required |
| `api METHOD /RESOURCE [JSON_BODY]` | Call another versioned API resource | API token required |
| `login USERNAME` | Authenticate and save a regular-user API token | Reads the password from standard input |

The generic `api` command accepts `GET`, `POST`, `PUT`, `PATCH`, and `DELETE`.
Its resource path is appended to `/api/{api-version}`; it must start with `/`
and cannot contain a query, fragment, or `..` sequence. Supply request
data as one JSON argument.

## Examples

```bash
casci-cli --server https://casci.example --version
casci-cli --server https://casci.example status
casci-cli --server https://casci.example --token-file "$HOME/.config/webappsgo/casci/token" projects list
casci-cli --server https://casci.example --token "$CASCI_TOKEN" projects create demo https://git.example/team/demo.git main
printf '%s\n' "$CASCI_PASSWORD" | casci-cli --server https://casci.example login alice
casci-cli --server https://casci.example --token "$CASCI_TOKEN" api GET /projects
```

The client requires `--server URL` unless the build embeds an official site
from `site.txt` or `OFFICIAL_SITE`. It sends tokens only in the `Authorization: Bearer` header and uses
the `casci-cli/{version}` User-Agent. HTTPS is required for non-local server
URLs; plain HTTP is accepted only for loopback development instances. Login
for a server administrator returns a browser session rather than a user API
token; the CLI login flow therefore requires regular-user login to be enabled.

## Server executable

The `casci` binary starts the server. Its flags include `--help`, `--version`,
`--status`, `--mode`, `--config`, `--data`, `--cache`, `--log`, `--backup`,
`--address`, `--port`, `--baseurl`, `--daemon`, `--debug`, `--color`, `--lang`,
`--service`, `--maintenance`, `--update`, and `--shell`. See
[Configuration](configuration.md) for runtime paths and precedence.

The client currently provides command-line API access. Its interactive
TUI/GUI setup wizard, full command tree, localized output, and automatic
version compatibility check are not implemented. The server binary provides
shell completion scripts through `casci --shell completions {bash,zsh,fish}`.
The client and agent are documented separately in [Agent management](agents.md);
their `--version` output also includes the commit ID and build timestamp.
