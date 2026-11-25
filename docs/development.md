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

The repository includes `tests/run_tests.sh`, `tests/docker.sh`,
`tests/api.sh`, `tests/incus.sh`, and `tests/e2e.sh`. Review each runner before
using it; some require an existing container runtime, browser, network access,
or a running application. The primary Go unit suite runs through `make test`.

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
