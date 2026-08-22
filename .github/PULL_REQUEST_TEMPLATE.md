# Pull Request

## Summary

What this pull request changes, in one or two sentences.

## Why

The problem this solves or the requirement it satisfies. Link the issue it closes
(`Closes #123`) or the `AI.md` / `IDEA.md` section it implements.

## Test Evidence

Paste the actual commands you ran and their results. "Should work" is not evidence.

- [ ] `make test` passes
- [ ] Unit coverage is at or above 60 percent
- [ ] `gofmt` and `go vet` are clean
- [ ] `./tests/run_tests.sh` run for runtime behavior changes
- [ ] `./tests/e2e.sh` run for web frontend changes

```text

```

## Documentation and Configuration

- [ ] `README.md` and `docs/` updated for every new or changed flag, environment
      variable, config key, route, or default
- [ ] Sample configuration and `docker/` compose files updated where affected
- [ ] No documentation change needed — explain why:

## Breaking Changes

- [ ] None — this is backward compatible
- [ ] Yes — describe what breaks, who is affected, and the migration path:

## Security and Privacy Impact

- [ ] No security or privacy impact
- [ ] Yes — describe the impact below

Confirm for every change:

- [ ] No tokens, API keys, passwords, or private keys are committed
- [ ] Credentials and personal data are redacted in logs, tests, and fixtures
- [ ] Untrusted input is validated; queries are parameterized; path traversal, SSRF,
      XSS, CSRF, and IDOR are guarded where relevant
- [ ] Any new GitHub Action is pinned to a full 40-character commit SHA with the
      version in a trailing comment

## Completeness Checklist

- [ ] Every line works as written — no stubs, no placeholder behavior, no dead code
- [ ] No `TODO`, `FIXME`, or `HACK` markers introduced
- [ ] No commented-out code left behind
- [ ] Comments sit above the code they describe, never trailing it
- [ ] Every text file ends with a single trailing newline
- [ ] One logical change — unrelated fixes are split into their own pull requests
