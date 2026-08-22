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

## 2. LICENSE changed from WTFPL to MIT (needs user confirmation)

`LICENSE.md` was WTFPL v2, `Copyright (C) 2025 casjay <git-admin@casjaysdev.pro>`.
AI.md PART 2 states "All projects MUST use the MIT License" with
`Copyright (c) {year} {project_org}`, plus a mandatory third-party
attribution section.

Rewritten this session to MIT / `Copyright (c) 2025 webappsgo`, with an
Embedded Licenses table covering all 12 direct dependencies.

Needs confirmation because relicensing is legally significant even when the
sole copyright holder does it and even when both licenses are permissive.
Revert the file if the WTFPL choice was deliberate and AI.md's rule is meant
to be overridden here (that override belongs in `SPEC.md`, which is currently
empty).

## 3. Forbidden root files — relocated, needs review before commit

None of these are on AI.md's root-file allowlist. All are already staged:

- `build.sh` — deleted (`D`); its logic belongs in the `Makefile`
- `Makefile.old` — deleted (`D`)
- `API.md` -> `docs/api.md`
- `API_QUICKREF.md` -> `docs/api-quickref.md`
- `QUICKSTART.md` -> `docs/installation.md`
- `SERVICES.md` -> `docs/services.md`
- `Dockerfile.dev` -> `docker/Dockerfile.dev`
- `docker-compose.yml` -> `docker/docker-compose.yml`

Review before committing: the doc renames change published URLs, and
`build.sh`'s deletion assumes nothing external invokes it. Confirm the
`Makefile` now covers everything `build.sh` did.

## 4. Go module path — RESOLVED, listed here for review

`go.mod` declared `module github.com/casapps/casci`. AI.md (lines 6639, 6890,
10380) requires `module github.com/{project_org}/{internal_name}` —
i.e. `github.com/webappsgo/casci`, which also matches the actual git remote in
`IDEA.md`.

Renamed across `go.mod`, every import in the Go tree, and the repo URLs in the
`src/pkg/server/templates/` pages (about, contact, help, footer) — 31 files.
`internal_org: casapps` is unaffected and stays frozen; it governs the
on-disk paths (`/etc/casapps/casci`), not the module path.

Called out here only because it rewrites every import line in the diff.

## 5. Service files ship as standalone scripts (needs decision)

`scripts/` contains `casci.service`, `com.casapps.casci.plist`,
`install-systemd.sh`, `install-launchd.sh`, `install-windows.ps1`, and
`uninstall-service.sh`. AI.md PART 24/25 requires the binary itself to detect
the init system and generate/install these via `--service --install`; no
standalone service files or installer scripts should exist in the repo.

Not auto-removed: deleting them leaves no install path until
`--service --install` is actually implemented in Go (that work is tracked in
`AUDIT.AI.md`). Decide whether to keep them as a stopgap until the binary
implements PART 24/25, or remove them now.

## 6. Session signing key reuses the master encryption key

`src/pkg/server/server.go:99` calls
`users.NewAuthManager(cfg.EncryptionKey)`, so the same secret both derives
AES-GCM credential-encryption keys and signs session tokens. That is key
reuse across two unrelated cryptographic purposes: rotating the signing
secret (e.g. after a token leak) would silently make every stored credential
undecryptable, and vice versa.

The AI.md-prescribed fix is NOT a second env-sourced secret — PART 5
enumerates the environment variables exhaustively and defines no
`JWT_SECRET`. AI.md PART 11 → "Other Project-Level Secrets" (line 15357)
instead requires a 32-byte `app_secrets.cookie_signing_key` row in
`server.db`, auto-rotated every 90 days with the previous key valid for 7
days, distributed to cluster replicas via shared DB read, alongside
`installation_secret` (line 15346) and `csrf_token_secret` (line 15358).
`server.security.encryption_key` (line 15364) stays in `server.yml` and is
the at-rest AES key only.

Left open because it is not a local patch: it requires the `app_secrets`
table, the rotation/grace-period machinery, and a decision on whether the
current JWT-based session mechanism is replaced by the signed session
cookies AI.md actually describes. Tracked alongside the rest of the PART 11
key-management work in `AUDIT.AI.md`.

## 7. `web.cors` defaults to `*` — needs an explicit decision

`src/internal/config/config.go` sets `defaultCORS = "*"`, so a fresh install
ships `Access-Control-Allow-Origin: *` on every response. The README and the
docs previously described the default as empty/same-origin-only and framed
that as a deliberate security choice, so the documentation and the code
disagreed; the docs have been corrected to describe the wildcard honestly,
but the default itself is untouched.

Not changed here because tightening it to same-origin is a user-visible
behavior change that would break any existing browser client calling this
API cross-origin. Decide whether the shipped default becomes empty
(same-origin) or stays `*`.

## 8. PART 11 security-report subsystem still unbuilt

The `Policy:` 404 is fixed — `/server/security` is now mounted
(`handleSecurityPolicy` in `src/pkg/server/wellknown.go`, rendering
`templates/pages/security.tmpl`), and `llms.txt` is served at both
`/.well-known/llms.txt` and `/llms.txt` per AI.md:16353-16416.

Still unbuilt: the wider PART 11 security-report subsystem — PGP keypair,
`installation_secret`-derived `{security_id}`, the
`Contact: /server/contact?security_id={id}` preference-order line that
AI.md:16299-16351 puts ahead of the mailto CC,
`/server/security/report/{tracking_id}`, and `/server/security/thanks`.
All of it depends on the `app_secrets` table from item 6.

## 9. CLI flags parsed but inert

`--shell` is entirely unimplemented, and `--pid`, `--daemon`, `--baseurl`,
and `--lang` are accepted and parsed but have no effect. `--baseurl` in
particular is inert because `wellknown.go` deliberately ignores all
`X-Forwarded-*` headers — honoring them needs a trusted-proxy allowlist that
does not exist yet (AI.md PART 8 "Resolution Order" requires exactly that
allowlist). The docs now mark each of these as unimplemented rather than
describing behavior that does not happen.

Left open because implementing `--baseurl` correctly means building the
trusted-proxy allowlist, which is a PART 8 feature and a design decision,
not a local patch.

## 10. Go unit test coverage far below the PART 29 gate

`src/pkg/server/templates_test.go` is the only Go test in the tree — it
covers template-set isolation, template-reference resolution, and the
no-inline-handler CSP rule. AI.md PART 29 requires a >=60% coverage gate;
every other package (`users`, `builds`, `projects`, `credentials`,
`webhooks`, `executor`, `artifacts`, `security`, `audit`, `compliance`,
`config`) still has none. Mechanical work, not a decision.

## 11. Admin settings form cannot save

`templates/admin/settings.tmpl` posts to
`/api/{api_version}/server/{admin_path}/config/settings`, but no handler is
mounted there — the form is inert. AI.md:9077 and AI.md:13039 specify the
write path plus pending-restart semantics: accepted changes are persisted,
and the UI must show a restart-required warning until the process is
restarted. Needs the settings-write handler, the persistence path back to
`server.yml`, and the pending-restart flag surfaced in `PageData`.

## 12. No HTML auth pages

AI.md:3213 defines `/server/auth/*` and
`/api/{api_version}/server/auth/*`. The JSON API half now lives at the
correct paths (`server.go`), but there are no HTML login/register/logout
pages — the header Login button points at the admin root and relies on the
401 error page. Needs `pages/login.tmpl`, `pages/register.tmpl`, and the
matching GET/POST handlers.

## 13. Admin sub-pages removed from the nav until they exist

`layouts/admin.tmpl` previously linked routes with no handler, so every
entry but Dashboard and `config/settings` 404'd. The nav was reduced to what
is actually mounted. Still to build, per PART 1's admin route table:
`config/users`, `config/cluster`, `config/security/*`, `config/compliance`,
`config/logs/audit`.

## 14. Contact form does not send mail

`handleContact` in `src/pkg/server/handlers_web.go` validates and accepts the
submission but never delivers it — it depends on the PART 12
`server.contact.*` subsystem (SMTP config, rate limiting, spam controls),
which is unbuilt. The form currently reports success without sending.

## 15. `writeJSON` is duplicated and 45 encode errors are still ignored

`writeJSON` exists three times over (`pkg/users/handlers.go`,
`pkg/builds/handlers.go`, `pkg/projects/handlers.go`) — identical bodies,
now each with error logging added. The remaining ~45 handlers in
`webhooks`, `audit`, `compliance`, `credentials`, and `security` call
`json.NewEncoder(w).Encode(...)` inline and discard the error. The fix is
one shared response helper, but where it lives depends on the PART 3
package flatten (`src/pkg/{...}` -> `src/{...}`) tracked in `AUDIT.AI.md`,
so it is sequenced after that move rather than creating a package the
flatten would immediately relocate.

## 16. API version hardcoded in path parsing

`pkg/users/handlers.go`, `pkg/projects/handlers.go` and
`pkg/builds/handlers.go` extract path IDs with
`strings.TrimPrefix(r.URL.Path, "/api/v1/users/")` and similar literals.
`server.api_version` is operator-configurable, so with any value other than
`v1` the prefix never matches, the untrimmed path fails `strconv.Atoi`, and
every by-ID request returns 400. Needs the configured version threaded into
the handlers (or a router that captures the ID as a parameter).
