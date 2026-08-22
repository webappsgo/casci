# Development

Development instructions are for contributors. Operators running CASCI should use [Installation](installation.md) instead.

## Prerequisites

| Tool | Why |
|------|-----|
| Docker | Every build and test run happens in a container |
| Git | Source control, and the version metadata embedded at build time |
| Make | Entry point for every workflow |

Go is **not** required on the host. The Makefile runs the toolchain inside `casjaysdev/go:latest`, so builds are reproducible and no host toolchain version can influence the result.

## Getting the Source

```bash
git clone https://github.com/webappsgo/casci
cd casci
```

## Layout

```
src/            Go source
  main.go       Entry point
  internal/     Internal packages, not importable externally
  pkg/          Application packages
tests/          Shell-driven integration and end-to-end tests
scripts/        Service unit files and platform installers
docker/         Dockerfile and docker-compose.yml
docs/           This documentation set
binaries/       Build output, gitignored
releases/       Release archives, gitignored
```

Go directories are singular (`handler/`, `model/`, `middleware/`) so a directory name matches its package name. Tooling directories stay plural.

## Make Targets

| Target | Does |
|--------|------|
| `make dev` | Quick build to a temp directory for local iteration |
| `make build` | Cross-compile every supported platform into `binaries/` |
| `make local` | Build only for the current host platform |
| `make release` | Build and package release archives into `releases/` |
| `make docker` | Build the container image |
| `make test` | Run the test suite with coverage enforcement |
| `make clean` | Remove `binaries/` and `releases/` |

### Quick Build

```bash
make dev
```

Output goes to a random temp directory rather than the source tree, so an in-progress build never gets committed. The target prints the exact path when it finishes.

### Full Build

```bash
make build
```

Targets `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`, `freebsd/amd64`, and `freebsd/arm64`. Override with `PLATFORMS`:

```bash
make build PLATFORMS="linux/amd64 linux/arm64"
```

Binaries are static with `CGO_ENABLED=0`, stripped with `-s -w`, and carry version, commit, and build epoch through linker flags.

### Container Image

```bash
make docker
```

## Running Locally

```bash
# Build, then run the binary the dev target printed
/tmp/webappsgo/casci-XXXXXX/casci --mode development --debug
```

The first run picks a random port in `64000-64999`, writes it to `server.yml`, and reuses it thereafter. Pin it while developing:

```bash
casci --port 64580 --mode development --debug
```

Use throwaway directories so a development run never touches a real install:

```bash
casci --config /tmp/casci-dev/config \
      --data /tmp/casci-dev/data \
      --log /tmp/casci-dev/log \
      --port 64580
```

See [CLI Reference](cli.md) for the full flag list.

## Testing

```bash
make test
```

The suite runs `go test ./...` with coverage and **fails below 60% coverage**. Coverage output is written to a temp directory, never into the source tree.

`tests/` also holds shell-driven suites:

| Script | Covers |
|--------|--------|
| `tests/run_tests.sh` | Test runner entry point |
| `tests/api.sh` | REST API behavior against a running instance |
| `tests/e2e.sh` | End-to-end flows |
| `tests/docker.sh` | Container image behavior |
| `tests/incus.sh` | Incus container behavior |

Every behavior change needs a test that fails before the change and passes after.

## Code Style

- `gofmt` is authoritative: tabs for indentation, standard import grouping
- Package names are lowercase, single-word, and match their directory
- Comments sit **above** the code they describe, never at the end of a line
- No `TODO`, `FIXME`, or `HACK` markers in committed code
- No commented-out code
- Every file ends with exactly one trailing newline
- Errors are wrapped with context rather than returned bare
- Every `curl` example in code, scripts, and docs uses `curl -q -LSsf`

Run the lint gate before opening a pull request:

```bash
go-lint
```

## Adding a Pipeline Format

Pipeline parsing lives in `src/pkg/pipeline/parser.go`. A new format needs:

1. A `PipelineFormat` constant
2. A detection entry in `DetectFormat`, ordered so it cannot shadow an existing format
3. A parse function converting the format into the internal pipeline model
4. Tests covering detection and conversion
5. A row in the format table in [Integrations](integrations.md)

## Adding a Notification Provider

Providers live in `src/pkg/notifications/`. A new one needs a `NotificationType` constant, a sender implementation, registration so `/api/v1/notifications/types` reports it, tests, and a row in the provider tables in [Integrations](integrations.md).

## Documentation

Documentation is MkDocs with the Material theme, hosted on ReadTheDocs. `docs/` holds only ReadTheDocs files — source belongs in `src/`, scripts in `scripts/`.

```bash
pip install -r docs/requirements.txt
mkdocs serve
```

Preview at `http://127.0.0.1:8000`. Check both themes before submitting; the toggle covers dark, light, and auto, and dark is the default.

Any change that adds or alters an operator-visible, admin-visible, or integrator-visible behavior must land with its documentation update in the same change. Shipping a feature without documenting it is treated as an incomplete change.

## Contributing

1. Fork and branch from `main`
2. Make the change, with tests
3. Run `make test` and `go-lint` — both must pass
4. Update the affected documentation pages
5. Open a pull request describing what changed and why

Guidelines:

- One logical change per commit
- No partially implemented code; every committed line works as written
- No secrets, tokens, or keys in the repository or in test fixtures
- Do not introduce a dependency that only saves a few lines
- Match the surrounding style rather than importing a personal one

## Reporting Issues

Bugs and feature requests belong at [github.com/webappsgo/casci/issues](https://github.com/webappsgo/casci/issues). Security vulnerabilities do **not** — report those privately per [Security](security.md).
