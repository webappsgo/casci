# Makefile Rules (PART 26)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Guess or assume the official-site value (`site.txt`) — it must be explicitly created by the user
- Add a `v` prefix to a non-numeric version tag (`vdev`, `vbeta`, `v20251218`) — `v` prefix is ONLY for numeric semantic versions (`vX.Y.Z`)
- Double the `v` prefix (`vv0.3.0`)
- Hardcode `PROJECT_NAME` / `PROJECT_ORG` in the Makefile — always inferred from git remote or directory path
- Infer the official site from project name, domain guesswork, or any source other than `site.txt`
- Copy or symlink built binaries out of `binaries/` — they stay there until explicitly moved for release (the sole exception is the CI/CD release step)
- Use the Makefile for CI/CD builds — it is local-dev only; CI/CD uses direct `go build`

## CRITICAL - ALWAYS DO
- Read the version from `release.txt`
- Follow the binary naming pattern and directory structure exactly for `make build`/`make release`/`make docker`
- Embed build info (version, commit, build date) via `-ldflags` at compile time
- Cache Go modules between builds where the build environment supports it
- Produce both server and CLI binaries (when the project has a CLI) plus metadata files (`version.txt`, etc.) on release
- Distinguish stable / beta / daily release types per their defined tagging rules

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Where does the version come from? | `release.txt` | PART 26, Version File |
| Does the CI pipeline use `make`? | No — CI/CD uses direct `go build`, not Makefile targets | PART 26, Build Matrix |
| Can binaries be symlinked elsewhere for convenience? | No — never copied/symlinked out of `binaries/` except the CI/CD release step | PART 26, NEVER Copy or Symlink Binaries |
| When does a version get a `v` prefix? | Only for numeric `X.Y.Z` semantic versions | PART 26, Version Tag `v` Prefix Rules |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| `release.txt` | Canonical version-number source file |
| `site.txt` | Optional, user-authored official-site URL file — never inferred |
| `binaries/` | Build output directory; binaries live here until release, never duplicated elsewhere |
| Stable / Beta / Daily | The three release types, each with its own tagging convention |

## QUICK REFERENCE
- `make build`, `make release`, `make docker`, `make test`, `make dev`, `make local` are the defined targets.
- Version and binary naming are fully deterministic from `release.txt` + inferred project name/org.
- Local dev only — CI/CD never calls `make`.

---
For complete details, see AI.md PART 26
