# CI/CD Rules (PART 28)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Use Makefile targets in any CI/CD workflow — CI uses explicit `go build` commands with all flags visible
- Reference local user paths (`~/.local/share/go`) in CI — use `/tmp/` or CI-native caching
- Depend on local Docker containers for GitHub Actions builds — GitHub Actions uses native Go, not a mounted `casjaysdev/go:latest` container, for the build step itself
- Cross-cancel different release refs — auto-cancel only applies to older runs of the exact same branch/tag ref
- Include a `-musl` suffix in binary naming
- Mix local-development conventions (Makefile, host cache dirs) into CI/CD workflows — they are different execution contexts

## CRITICAL - ALWAYS DO
- Use explicit `go build` commands with all flags visible in every CI/CD job
- Use CI-native caching, never local host cache paths
- Set `VERSION`, `COMMIT_ID`, `BUILD_EPOCH` explicitly (`BUILD_DATE` is derived from `BUILD_EPOCH` for Docker OCI labels only, not an ldflag)
- Build all platforms in the release build matrix
- Auto-cancel older in-progress runs for the same ref on push workflows targeting `main`, `master`, `devel`, `dev`, or `beta`
- Place workflow files at `.github/workflows/*.yml` for this project (GitHub-hosted, no self-hosted CI option for GitHub Actions)

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Does CI use the Makefile? | Never — explicit `go build` with visible flags | PART 28, CI/CD vs Local Development |
| Where do workflow files live for this project? | `.github/workflows/*.yml` (GitHub remote) | PART 28, Workflow Files table |
| How is build info injected? | `VERSION`, `COMMIT_ID`, `BUILD_EPOCH` set explicitly as ldflags; `BUILD_DATE` derived from `BUILD_EPOCH` for Docker labels only | PART 28, CI/CD workflows MUST |
| Can in-progress runs be auto-canceled across different refs? | No — only older runs on the identical branch/tag ref | PART 28, CI/CD workflows MUST NOT |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| CI-native caching | The CI provider's own cache action/mechanism, not a bind-mounted host path |
| Build matrix | The full platform/arch combination built on every release workflow run |
| `BUILD_EPOCH` | Unix timestamp embedded at build time; source for `BUILD_DATE` (labels only) and `{build_datetime}` |

## QUICK REFERENCE
- CI/CD is a separate execution context from local dev — never Makefile, always explicit commands.
- This project uses GitHub Actions exclusively: `.github/workflows/*.yml`.
- Every release build sets VERSION/COMMIT_ID/BUILD_EPOCH and builds the full platform matrix.
- Auto-cancel is scoped strictly to the same ref.

---
For complete details, see AI.md PART 28
