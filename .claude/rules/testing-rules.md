# Testing & Documentation Rules (PART 29, 30, 31)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Run `go build`, `go test`, or any built binary directly on the host — Go is not installed locally; everything runs in a container
- Run host-affecting commands (`reboot`, `systemctl`, `iptables`, `mount`, package installs) outside a container/VM/namespace — there is no testing exemption
- Use the project directory for runtime or test data, or mount project paths (`./volumes/`, `./docker/rootfs/`) as runtime volumes
- Use bare `mktemp -d`, `/tmp` root, or a generic temp path — temp dirs must be `/tmp/{project_org}/{internal_name}-XXXXXX/`
- Use `docker-compose.yml` or `docker-compose.dev.yml` as an AI assistant — those are human-only
- Commit config files (`server.yml`, `cli.yml`, torrc) to the repo — they are runtime-generated only
- Bypass authentication in tests, debug mode, beta testing, or CI — debug mode adds verbosity only, never an auth bypass
- Hardcode dev credentials or add auth-bypass flags/backdoors in any mode
- Use broad kill/prune commands: `pkill -f`, `killall`, `docker kill`, `docker system/container/image/volume/network prune`, `docker rm $(docker ps -aq)`
- Delete user data, configs, databases, SSL certs, or git repos without explicit confirmation
- Let `make test` run the browser E2E suite, or make E2E a required CI check
- Put non-ReadTheDocs files in `docs/`
- Hardcode any human-readable string — every user-facing string goes through a translation key
- Use `fmt.Sprintf`, `%s`/`%d` verbs, or positional `{0}` placeholders in locale files — literal `{variable}` replacement only
- Error or crash on an unsupported `--lang`/`Accept-Language` value — fall back silently to `en`
- Convey information by color alone

## CRITICAL - ALWAYS DO
- Run both required test phases: Phase 1 toolchain gate (`*_test.go` via `make test`, pre-commit) and Phase 2 binary validation (`./tests/*.sh`, manual)
- Create or update the matching `*_test.go` in the same work pass that changes package logic — never defer tests to the end
- Hold ≥60% Go unit coverage and 100% endpoint/route coverage; critical paths (auth, DB, token validation) and all reachable error paths are always tested
- Ship the minimum test scripts: `tests/run_tests.sh`, `tests/docker.sh`, `tests/incus.sh`, `tests/e2e.sh`
- Test every route with all applicable `Accept` headers (frontend: `text/html` + `text/plain`; API: `application/json` + `text/plain`)
- Prefer Incus + Debian for integration testing; fall back to Docker with `casjaysdev/go:latest` or `alpine:latest`
- Verify a PID or container name before any kill/remove, terminate gracefully first, and scope every removal to this project
- Run all three E2E tiers (SSR, no-JS browser, full browser) when E2E runs — a feature passing only Tier 3 is a progressive-enhancement violation
- Keep E2E hermetic: seeded fixtures, zero external network, port from 64000-64999, artifacts to the tempdir structure
- Convert every confirmed AI exploratory-testing finding into a committed deterministic chromedp test, and fix or log every finding in `TODO.AI.md`
- Ship `mkdocs.yml`, `.readthedocs.yaml`, and the required `docs/` pages (`index`, `installation`, `configuration`, `api`, `admin`, `security`, `integrations`, `development`, `requirements.txt`)
- Keep `docs/` describing the shipped product as it actually behaves, including public discovery/protocol endpoints
- Support all seven languages (`en`, `es`, `zh`, `fr`, `ar`, `de`, `ja`) in every binary — server, CLI, and agent
- Resolve language via the fixed chain: `?lang=` query param (sets cookie) → `lang` cookie → `Accept-Language` → `en`
- Keep every locale file key-identical to `en.json` and enforce it with `make i18n-validate` at build time
- Meet WCAG 2.1 AA: full keyboard operation, screen reader support, 4.5:1 text contrast, visible focus indicators, 44x44px touch targets, skip links first in tab order

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Where does test/runtime data go? | `/tmp/{project_org}/{internal_name}-XXXXXX/` — never the project directory | PART 29, Temporary Directory Structure |
| Which compose file may AI use? | Only `docker-compose.test.yml`, preferably via `tests/` scripts | PART 29, AI Docker Compose Rules |
| What is the unit coverage floor? | ≥60% via `go test -cover`; endpoints/admin routes are 100% | PART 29, Test Coverage Gates |
| Does debug mode ever skip auth? | Never — verbosity and diagnostics only | PART 29, Debug Mode - Never an Auth Bypass |
| Does `make test` run E2E? | No — E2E is `./tests/e2e.sh`, on demand, never a commit gate | PART 29, Browser E2E Testing |
| Are config files committed? | No — generated at runtime on first run | PART 29, Config Files: Runtime-Generated Only |
| What may live in `docs/`? | ReadTheDocs/MkDocs files only | PART 30, Overview |
| What is the default language? | English (`en`), also the fallback for missing keys and unsupported languages | PART 31, Core Requirements |
| How is language chosen in a URL? | `?lang=` query parameter that sets a 1-year cookie — never a URL path prefix | PART 31, Language Selection via Query Parameter |
| Which accessibility standard applies? | WCAG 2.1 AA, mandatory | PART 31, Accessibility (a11y) → Core Requirements |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| Phase 1 — Toolchain Gate | `*_test.go` run by `make test`; the pre-commit gate for source logic |
| Phase 2 — Binary Validation | `./tests/*.sh` shell scripts exercising the compiled binary; manual, developer-initiated |
| `./tests/*` | Executable shell scripts in the repository-root `tests/` directory — never Go `*_test.go` files |
| Tier 1 / 2 / 3 | E2E tiers: SSR (no browser), no-JS browser, full browser — all three mandatory |
| AI Exploratory Pass | Non-reproducible AI-driven browser testing; discovery only, never a gate |
| Fallback chain | `?lang=` → cookie → `Accept-Language` → `en` |
| `meta.direction` | Per-locale field driving the HTML `dir` attribute (`ltr`/`rtl`) |

## QUICK REFERENCE
- Containers for everything; host stays pristine; temp dirs always `/tmp/{org}/{project}-XXXXXX/`.
- Two mandatory phases: `make test` (≥60%, pre-commit) and `./tests/*.sh` (100% endpoints, manual).
- E2E: three tiers, hermetic, on demand via `./tests/e2e.sh`, never blocking.
- Kill/remove only the exact PID or `{project_name}` container after listing and verifying.
- `docs/` = MkDocs Material on ReadTheDocs, nine required pages plus `requirements.txt`.
- i18n: 7 languages, all binaries, no hardcoded strings, `make i18n-validate` at build time.
- a11y: WCAG 2.1 AA, keyboard-complete, 4.5:1 contrast, skip links, 44x44px targets.

---
For complete details, see AI.md PART 29, 30, 31
