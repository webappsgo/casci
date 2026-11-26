# Development

CASCI is a Go project. Use the repository's Docker-based Make targets rather
than installing a separate host Go toolchain.

## Requirements

- Docker with the ability to run the project toolchain image
- Git
- A working directory with enough space for the Go module and build caches

The Makefile uses `casjaysdev/go:latest` and keeps build outputs in the
project's ignored `binaries/` directory or an OS temporary directory.

## Build and test

```bash
make dev
make build
make test
```

| Target | Behavior |
|--------|----------|
| `make dev` | Build for the current platform into a temporary directory |
| `make build` | Build the configured operating-system and architecture set |
| `make local` | Build for the current platform into `binaries/` |
| `make docker` | Build a multi-platform container image locally |
| `make test` | Run `go test ./...` with coverage enforcement |
| `make clean` | Remove generated binary/release output |

The test gate requires 60% total statement coverage. The current repository
suite has not met that threshold yet, so `make test` can fail after individual
packages pass. Do not lower the threshold to make the gate pass.

Build and test artifacts belong in ignored output or project-scoped OS
temporary directories, not among tracked source files.

## Dependency licenses

Run `scripts/verify-licenses.sh` inside `casjaysdev/go:latest`. It checks for
GPL-family dependencies and unreviewed license identifiers, then prints the
scanner report. Reports and saved license files are created under a private
project-scoped temporary directory and removed when the script exits. The
current `modernc.org/mathutil` license is verified against its BSD-style terms
because the scanner does not identify that license automatically; review the
source license again if that dependency changes.

## Source layout

```text
src/        Go application and packages
tests/      Docker/API/Incus/browser test runners
docker/     Production, development, and test container files
scripts/    Service installation helpers
docs/       MkDocs documentation
```

Go unit tests are kept beside the package source as `*_test.go` files.

## Test runners

`tests/run_tests.sh` runs the basic Docker suite in `tests/docker.sh` and exits
with status 69 when Docker is unavailable. It accepts no arguments. Run
`tests/incus.sh` and `tests/e2e.sh` explicitly for their separate integration
and browser suites; `tests/api.sh` targets an already-running server. The
primary Go unit suite runs through `make test`.

The browser suite builds the production image, starts the application and a
pinned headless Chromium container on an isolated Docker network, and checks
server-rendered pages, a JavaScript-disabled preferences submission, and the
administrator login, audit-page, settings, and logout flow. It requires Docker
and OpenSSL. A temporary
test CA enables HTTPS so the browser exercises the same `Secure` session-cookie
rules as a deployed server; the test keys and certificates are removed when
the runner exits.

## Documentation

The docs use MkDocs Material and ReadTheDocs configuration. Install the
Python dependencies from `docs/requirements.txt` in an isolated environment,
then preview with `mkdocs serve`. Update `mkdocs.yml` navigation whenever a
new documentation page is added.

## Contribution expectations

- Keep changes aligned with `AI.md`, `IDEA.md`, and `SPEC.md`.
- Add tests for behavior changes and update affected docs in the same change.
- Preserve existing user changes and do not add credentials or generated
  attribution to commits.
- Run the project-required tests, linters, build, and CI checks before merging.
