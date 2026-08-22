# Project Rules (PART 2, 3, 4)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Use GPL/AGPL/LGPL-licensed dependencies — they force derivative works to open-source under the same license
- Hardcode `{project_name}` or `{project_org}` — always infer from `git remote` or directory path
- Use `github.com/mattn/go-sqlite3` — CGO-based, forbidden
- Store plaintext passwords anywhere, in config files, or log them even hashed
- Allow passwords with leading/trailing whitespace — reject, never silently trim
- Run `go` directly — always use Makefile targets (`make dev`, `make test`, etc.) inside Docker
- Assume the current working directory is the project root — always resolve paths relative to project root
- Mix runtime-directory purposes — never put user config in `data_dir` or vice versa
- Commit runtime volume data (`volumes/`, binaries, releases) — all gitignored
- Hardcode a specific Go version — always use latest stable

## CRITICAL - ALWAYS DO
- Include `LICENSE.md` in project root (REQUIRED) plus a README.md license badge
- Attribute embedded/third-party licenses per the compact or full format when a dependency requires it
- Use `modernc.org/sqlite` for the SQLite driver — pure Go, no CGO
- Hash ALL passwords with Argon2id (OWASP 2023 parameters), random salt per password, store hash in the database only
- Keep `.gitignore` and `.dockerignore` present and current (both REQUIRED) — exclude runtime volume data from both source control and image
- Keep `CLAUDE.md`, `.claude/rules/`, and other AI loader files at their required locations
- Resolve all paths relative to PROJECT ROOT, determined via a project-root-detection mechanism, never the current working directory

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| SQLite driver? | `modernc.org/sqlite` only, never `mattn/go-sqlite3` | PART 3, Required Go Libraries |
| Password hashing? | Argon2id, OWASP 2023 params, DB storage only | PART 3, Password Hashing |
| Where does config live? | `server.yml`, never `.yaml` | PART 3/4, Config File |
| How to run Go commands? | Only via `make` targets inside Docker `casjaysdev/go:latest` | PART 3, Go Libraries |
| Where do variables like `{project_name}` come from? | Inferred from git remote / directory path, never hardcoded | PART 2, Variables |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| `{project_dir}` | Git repo root, or `$PWD` if not a git repo |
| `{project_name}` | `basename {project_dir}` |
| `{project_org}` | `basename $(dirname {project_dir})` |
| `config_dir` / `data_dir` / `log_dir` | OS-specific runtime directories (PART 4), never mixed |
| `server.yml` | The one and only config filename |

## QUICK REFERENCE
- License: `LICENSE.md` + README badge + third-party attribution for any dependency that requires it.
- Directory layout: `src/`, `docker/`, `docs/`, `scripts/`, `tests/`, `.claude/` per the exhaustive root-directory list.
- Config file is always `server.yml`.
- Passwords: Argon2id only, DB-stored, never logged or config-stored.
- SQLite: `modernc.org/sqlite` (pure Go) — never CGO driver.
- Go version: always latest stable, never pinned/hardcoded in docs.

---
For complete details, see AI.md PART 2, 3, 4
