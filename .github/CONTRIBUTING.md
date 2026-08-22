# Contributing to casci

Thanks for helping improve casci. This document describes how to build, test, and
submit changes.

## Before You Start

- Read `IDEA.md` for what casci is meant to do and `AI.md` for how it must be built.
  Both are the source of truth — a change that contradicts them will be rejected
  even if it works.
- Search existing issues before opening a new one.
- **Never open a public issue for a security vulnerability.** See
  [SECURITY.md](SECURITY.md) — report privately at
  <https://github.com/webappsgo/casci/security/advisories/new>.

## Local Setup

You need Docker and GNU Make. Everything builds inside `casjaysdev/go:latest`;
nothing is compiled on the host.

```bash
git clone https://github.com/webappsgo/casci.git
cd casci
```

## Build, Test, and Run

These are the real targets in the repository `Makefile`:

| Command | What it does |
|---------|--------------|
| `make build` | Cross-compile release binaries into `binaries/` (Docker) |
| `make local` | Build a single binary for the current platform into `binaries/` |
| `make release` | Build and stage release archives and checksums into `releases/` |
| `make docker` | Build the container image from `docker/Dockerfile` |
| `make test` | Run the Go unit and route test suite — the commit gate |
| `make dev` | Run the development stack from `docker/docker-compose.dev.yml` |
| `make clean` | Remove build output |

`make test` must pass before every commit. Unit test coverage must stay at or above
60 percent; CI fails the build below that.

## Binary Validation and E2E Tests

These are manual, developer-initiated suites. They are not part of the commit gate.

| Command | What it does |
|---------|--------------|
| `./tests/run_tests.sh` | Auto-detect Incus or Docker and run the full binary validation suite |
| `./tests/incus.sh` | Full integration plus systemd service tests in an Incus Debian container |
| `./tests/docker.sh` | Full integration tests in a Docker Alpine container |
| `./tests/run_tests.sh api` | HTTP API suite (`tests/api.sh`) against an already running server |
| `./tests/e2e.sh` | Browser E2E suite — all three tiers, Docker-wrapped headless Chromium |

Run `./tests/run_tests.sh` before opening a pull request that changes runtime
behavior, and `./tests/e2e.sh` for changes to the web frontend.

## Documentation

User-facing documentation lives in `docs/` and `README.md`. Any change that adds,
removes, or alters a flag, environment variable, config key, route, or default MUST
update the matching documentation in the same pull request.

## Branch and Pull Request Workflow

1. Fork the repository and branch from `main`.
2. Use a descriptive branch name (`fix/build-queue-deadlock`, `feat/gitea-webhooks`).
3. Keep one logical change per commit and per pull request.
4. Add or update tests for every behavior change — a test that fails before your
   change and passes after it.
5. Update documentation and configuration samples alongside the code.
6. Run `make test` and the relevant `./tests/*.sh` suite locally.
7. Open the pull request against `main` and fill in the pull request template
   completely.

Releases are tagged from `main`. `beta` carries prerelease builds; `daily` is an
automated rolling prerelease and is never branched from.

## Code Standards

- Go code must be `gofmt`-formatted and pass `go vet`; CI enforces both.
- No `TODO`, `FIXME`, `HACK`, stubs, or commented-out code in committed changes.
- Comments go on their own line above the code they describe, never trailing it.
- Never commit tokens, API keys, passwords, or private keys. Redact credentials in
  logs and test fixtures.
- Every third-party GitHub Action must be pinned to a full 40-character commit SHA
  with the version in a trailing comment. CI rejects tag-pinned actions.

## Reporting Bugs

Use the bug report template and include the version (`casci --version`), platform,
deployment mode, reproduction steps, and redacted logs. If the bug might have
security impact, stop and follow [SECURITY.md](SECURITY.md) instead.
