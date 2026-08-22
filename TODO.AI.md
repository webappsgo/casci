# TODO.AI.md

Logged during AI.md/IDEA.md/CLAUDE.md spec migration (2026-08-22). Each item
below is either destructive or ambiguous enough that it was not resolved
automatically — needs a decision from the user before acting.

## 1. SECURITY — committed encryption key (urgent, needs decision)

`.casci/data/encryption.key` is tracked in git (first added in commit
`ee3603791417`, still present in HEAD). This is a real 44-byte secret
material file, not a template/example.

Fixed this session: `.casci/` untracked (`git rm --cached`) and added to
`.gitignore` — the project-directory-for-runtime-data violation is resolved
per AI.md (`NEVER use project directory for test/runtime data`).

Still needs a decision (irreversible, not auto-fixed):
- Rotate/regenerate this key (any data encrypted with it becomes
  unrecoverable unless migrated first) or treat it as already compromised
  and rotate regardless, since it has been in git history.
- Scrub it from git history (`git filter-repo` / BFG) since the repo has a
  public remote (`github.com/webappsgo/casci`).

Resolved this session (removed from tracking, no longer open items):
- Legacy report/status docs (`COMPLETION_REPORT.md`, `COMPLIANCE_ROADMAP.md`,
  `COMPLIANCE_STATUS.md`, `FINAL_STATUS.md`, `PROGRESS_UPDATE.md`,
  `SESSION_SUMMARY.md`, `SUMMARY.md`, `DEVELOPMENT.md`) — `git rm`'d;
  superseded content already lives in `IDEA.md`/`AI.md`.
- `TEMPLATE.md` — `git rm`'d; `AI.md` confirmed byte-identical to the
  canonical `~/Projects/github/claudemgr/go/SERVER.md` template.
- `project_org` vs `internal_org` — confirmed correct as written:
  `project_org: webappsgo` (current), `internal_org: casapps` (frozen).
- `AI.md.old` — confirmed as a violation (stray `.old` file at project
  root); deletion staged.
