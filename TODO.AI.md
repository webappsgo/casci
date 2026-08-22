# TODO.AI.md

Logged during AI.md/IDEA.md/CLAUDE.md spec migration (2026-08-22). Each item
below is either destructive or ambiguous enough that it was not resolved
automatically — needs a decision from the user before acting.

Resolved and removed from this list (user confirmed "accept as-is" for the
audit-driven commits that landed them): LICENSE.md WTFPL→MIT relicense,
forbidden-root-file relocations (`API.md`, `API_QUICKREF.md`,
`QUICKSTART.md`, `SERVICES.md`, `Dockerfile.dev`, `docker-compose.yml`,
removal of `build.sh`/`Makefile.old`), and the `go.mod` module path rename
to `github.com/webappsgo/casci`.

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

Deferred by explicit user instruction: leave the git history as-is for now.

## 2. Service files ship as standalone scripts (needs decision)

`scripts/` contains `casci.service`, `com.casapps.casci.plist`,
`install-systemd.sh`, `install-launchd.sh`, `install-windows.ps1`, and
`uninstall-service.sh`. AI.md PART 24/25 requires the binary itself to detect
the init system and generate/install these via `--service --install`; no
standalone service files or installer scripts should exist in the repo.

Not auto-removed: deleting them leaves no install path until
`--service --install` is actually implemented in Go (that work is tracked in
`AUDIT.AI.md`). Decide whether to keep them as a stopgap until the binary
implements PART 24/25, or remove them now.

## 3. Session signing key reuses the master encryption key

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

## 4. `web.cors` defaults to `*` — needs an explicit decision

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

## 5. PART 11 security-report subsystem still unbuilt

The `Policy:` 404 is fixed — `/server/security` is now mounted
(`handleSecurityPolicy` in `src/pkg/server/wellknown.go`, rendering
`templates/pages/security.tmpl`), and `llms.txt` is served at both
`/.well-known/llms.txt` and `/llms.txt` per AI.md:16353-16416.

Still unbuilt: the wider PART 11 security-report subsystem — PGP keypair,
`installation_secret`-derived `{security_id}`, the
`Contact: /server/contact?security_id={id}` preference-order line that
AI.md:16299-16351 puts ahead of the mailto CC,
`/server/security/report/{tracking_id}`, and `/server/security/thanks`.
All of it depends on the `app_secrets` table from item 3.

## 6. CLI flags parsed but inert

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

## 7. Go unit test coverage is far below the PART 29 60% floor

Only `src/pkg/server/templates_test.go` exists in the entire ~20-package
tree, so `go test -cover ./...` reports roughly 0% overall against the
PART 29 required 60% floor (100% for endpoints/admin routes). The `test`
CI job's coverage-threshold step will stay red until a real test-writing
pass covers the untested packages (`users`, `projects`, `builds`,
`executor`, `nodes`, `security`, `audit`, `credentials`, `notifications`,
`webhooks`, `artifacts`, `database`, `config`, etc.). This is a
substantial, multi-session effort — tracked here rather than attempted
piecemeal so it gets a dedicated pass with the user's input on priority
order (auth/DB/token-validation critical paths first, per PART 29).

## 8. Admin settings form cannot save

`templates/admin/settings.tmpl` posts to
`/api/{api_version}/server/{admin_path}/config/settings`, but no handler is
mounted there — the form is inert. AI.md:9077 and AI.md:13039 specify the
write path plus pending-restart semantics: accepted changes are persisted,
and the UI must show a restart-required warning until the process is
restarted. Needs the settings-write handler, the persistence path back to
`server.yml`, and the pending-restart flag surfaced in `PageData`.

## 9. No HTML auth pages

AI.md:3213 defines `/server/auth/*` and
`/api/{api_version}/server/auth/*`. The JSON API half now lives at the
correct paths (`server.go`), but there are no HTML login/register/logout
pages — the header Login button points at the admin root and relies on the
401 error page. Needs `pages/login.tmpl`, `pages/register.tmpl`, and the
matching GET/POST handlers.

## 10. Admin sub-pages removed from the nav until they exist

`layouts/admin.tmpl` previously linked routes with no handler, so every
entry but Dashboard and `config/settings` 404'd. The nav was reduced to what
is actually mounted. Still to build, per PART 1's admin route table:
`config/users`, `config/cluster`, `config/security/*`, `config/compliance`,
`config/logs/audit`.

## 11. Contact form does not send mail

`handleContact` in `src/pkg/server/handlers_web.go` validates and accepts the
submission but never delivers it — it depends on the PART 12
`server.contact.*` subsystem (SMTP config, rate limiting, spam controls),
which is unbuilt. The form currently reports success without sending.

## 12. `writeJSON` is duplicated and 45 encode errors are still ignored

`writeJSON` exists three times over (`pkg/users/handlers.go`,
`pkg/builds/handlers.go`, `pkg/projects/handlers.go`) — identical bodies,
now each with error logging added. The remaining ~45 handlers in
`webhooks`, `audit`, `compliance`, `credentials`, and `security` call
`json.NewEncoder(w).Encode(...)` inline and discard the error. The fix is
one shared response helper, but where it lives depends on the PART 3
package flatten (`src/pkg/{...}` -> `src/{...}`) tracked in `AUDIT.AI.md`,
so it is sequenced after that move rather than creating a package the
flatten would immediately relocate.

## 13. API version hardcoded in path parsing

`pkg/users/handlers.go`, `pkg/projects/handlers.go` and
`pkg/builds/handlers.go` extract path IDs with
`strings.TrimPrefix(r.URL.Path, "/api/v1/users/")` and similar literals.
`server.api_version` is operator-configurable, so with any value other than
`v1` the prefix never matches, the untrimmed path fails `strconv.Atoi`, and
every by-ID request returns 400. Needs the configured version threaded into
the handlers (or a router that captures the ID as a parameter).

## 14. `github.com/docker/docker` — 2 unfixed upstream CVEs (blocked, not actionable)

`govulncheck` flags `GO-2026-4887` (AuthZ plugin bypass via oversized
request bodies) and `GO-2026-4883` (off-by-one in plugin privilege
validation) in `github.com/docker/docker`, reached via
`src/pkg/executor/container.go`'s use of the Docker client. Both report
`Fixed in: N/A` — no patched version exists yet upstream; the module is
already pinned to the latest available tag (`v28.5.2+incompatible`). The
`vuln-scan` CI job will stay red on these two findings until upstream ships
a fix; re-run `govulncheck ./...` periodically and bump
`github.com/docker/docker` the moment a fixed version is tagged. Not
exploitable via this project's own attack surface — the executor only acts
as a Docker API client for build containers, never runs an AuthZ plugin or
receives plugin privilege requests itself.
