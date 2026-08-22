# Project Audit

Started: 2026-08-22

Full six-pass audit against AI.md (65538 lines, PART 0-37). AI.md is the source
of truth; every item below is a place where the code contradicts it or where a
spec-mandated feature is absent. Delete each entry when fully resolved; delete
this file when all entries are gone.

## Pass 1: Security

- [ ] server/templates.go:220: CSP still allows 'unsafe-inline' in style-src (script-src is now nonce-based, so this is the last CSP gap)
- [ ] web.cors defaults to "*" in config.go — a fresh install ships a wildcard Access-Control-Allow-Origin. Needs a user decision on the default, so it is TODO.AI.md item 4 rather than a silent change
- [ ] No rate limiting exists anywhere — AI.md 4595-4615 specifies per-IP sliding windows (read 120/min, write 10/min, health 120/min, global burst 240/min) plus stricter auth limits (login 5/15min with lockout, registration 5/hour, upload 10/hour). Nothing is implemented and there is no config for it

## Pass 2: Code Quality

- [ ] 25 TODO/FIXME markers remain in production code (was 41) — mostly the S3/GCS/Azure storage stubs
- [ ] inline `//` comments — PART 0 requires comments above the code
- [ ] 134 stdlib log.Printf/Println call sites — PART 11 requires structured slog JSON
- [ ] cli/service.go, cli/update.go, cli/maintenance.go are stubs printing "not yet implemented"

## Pass 4: Documentation

- [ ] docs/: PART 30 page set and mkdocs nav — in progress

## Pass 5: Spec Compliance

- [ ] Package tree uses `src/internal/config` + `src/pkg/*` — PART 3 (AI.md 979-1054) specifies packages directly under `src/` (`src/config/`, `src/server/`, `src/graphql/`, `src/swagger/`, `src/common/`, `src/paths/`, `src/ssl/`, `src/mode/`, `src/scheduler/`, `src/service/`, `src/admin/`); the strings "src/pkg" and "src/internal" appear nowhere in AI.md
- [ ] API responses are ad-hoc — PART 9/14 require {"ok":…,"data":…} / {"ok":false,"error":CODE,…}, indented, newline-terminated
- [ ] Routes end with `/` — PART 14 forbids it
- [ ] CLI exposes 11 flags — PART 8 requires --cache --log --backup --pid --baseurl --daemon --debug --color --lang --shell among others
- [ ] Go test coverage is far below the PART 29 >=60% gate — src/pkg/server is the only package with tests
- [ ] scripts/ ships standalone service files — PART 24/25 require the binary to generate them via --service --install
- [ ] No i18n package, no locales, no src/common/ — PART 31
- [ ] No scheduler (PART 19), GeoIP (PART 20), backup (PART 22), self-update (PART 23) packages
- [ ] No Let's Encrypt/autocert support — PART 15 requires it on all projects
- [ ] No Valkey/Redis cache driver — PART 12 marks it REQUIRED

## Pass 6: Code Flow Trace

- [ ] ~45 handlers in webhooks/, audit/, compliance/, credentials/ and security/ still discard the json.NewEncoder(w).Encode error; writeJSON is duplicated in three packages. Sequenced behind the PART 3 package flatten (TODO.AI.md item 12)
- [ ] users/projects/builds handlers parse path IDs with a hardcoded "/api/v1/" literal, so any other server.api_version makes every by-ID route return 400 (TODO.AI.md item 13)

## Completed

- users/service.go: password hashing bcrypt(DefaultCost 10) -> Argon2id per PART 11; verified no bcrypt reference remains anywhere in src/
- users/repository.go: API tokens now stored and looked up as SHA-256 hashes instead of cleartext
- users/middleware.go: API token no longer accepted from the `?token=` query parameter (tokens land in access logs, proxy logs, and Referer headers); verified zero remaining call sites
- security/repository.go: unchecked `limit.(int)` type assertion removed — it panicked the request goroutine on any non-int
- server/server.go: /metrics, /metrics/json, and the four /api/{v}/metrics/* routes are now wrapped in userMiddleware.RequireAdmin — PART 11 marks metrics INTERNAL and they were fully public
- server/handlers_web.go: the `// TODO: Check authentication` placeholders on the admin pages are gone; admin routes are authenticated
- internal/config/config.go: TLS min version is configurable and defaults to TLS1.2, closing the PART 15 gap where TLS 1.0/1.1 were reachable
- server/server.go: health routes restructured to PART 13 — canonical /server/healthz and /api/{v}/server/healthz share one handler, and root /healthz is now an opt-in alias behind HealthzRootEnabled (default false) rather than a second independent route
- server/templates/ (admin/dashboard.tmpl, pages/error.tmpl, partials/nav.tmpl, partials/footer.tmpl): health links pointed at /healthz and /health. /health was never registered on the router at all, and /healthz is now off by default, so every one of these was a 404 in a default install. All four repointed at the canonical /server/healthz
- nodes/repository.go: node join tokens were persisted and looked up in cleartext, while the service doc comment already claimed only the SHA-256 hash was stored. CreateToken and GetToken now both hash at the repository boundary, so the raw token exists only in the single response that issues it
- server/csrf.go + server/cors.go: CSRFMiddleware and CORSMiddleware were both defined but never wrapped around the mux — dead code providing zero protection. Both are now in the chain at server.go:222 behind SecurityHeadersMiddleware
- server/templates.go: security headers added (HSTS et al) and script-src moved to a per-request nonce instead of 'unsafe-inline'
- audit/builds/security repositories: LIMIT/OFFSET are now bound parameters instead of string-interpolated
- server/wellknown.go: /.well-known/security.txt now served per PART 11, with every embedded URL built per request rather than frozen at startup
- webhooks/handler.go: findProjectByURL was a stub returning "project lookup not implemented", so every correctly signed push webhook answered 404 and no push ever triggered a build. Implemented as canonical host/path matching that handles scp-style git@host:owner/repo, embedded credentials, ports, query/fragment, .git suffix, and case, since the stored and delivered URL spellings routinely differ; added exported ErrProjectNotFound so callers can tell not-found from lookup failure
- docs/: seven pages corrected against actual behavior — fabricated server.yml schema, a fabricated SMTP_* env block, wrong config key names throughout (server.listen, server.cors_origins, server.tls.*, server.name), a non-existent "Service account" role, an invented proxy header resolution table, and three database drivers the code does not accept
- docker/: production Dockerfile added; Dockerfile.dev rebuilt (was golang:1.21-alpine with CGO_ENABLED=1 building a nonexistent ./cmd/casci); rootfs/ entrypoint added; docker-compose.yml brought to spec and dev.yml/test.yml added
- tests/: was empty; now carries the PART 29 harness (run_tests.sh, docker.sh, incus.sh, e2e.sh) plus api.sh, which is the old root-level test-api.sh relocated off the root allowlist and converted to `curl -q -LSsf`
- .github/workflows/: all six PART 28 workflows added (ci, release, beta, daily, docker, docker-aio)
- Jenkinsfile: added at root per PART 3
- .claude/rules/: all 14 PART 0 rule files now present (was 10) — testing-rules.md and optional-rules.md were the missing pair
- build.sh: removed from root — it built on the host, which PART 1 forbids; the Makefile is the only build entry point
- go.mod + 31 files: module path github.com/casapps/casci -> github.com/webappsgo/casci per PART 7 ({project_org}/{internal_name}); also rewrote the repo URLs in src/pkg/server/templates/ (about, contact, help, footer). internal_org casapps stays frozen and still governs on-disk paths
- src/main.go: package main now exists, so ./src is a buildable main package (PART 7/26) — the project previously could not produce a binary at all
- artifacts/storage.go: path traversal — all five LocalStorage methods built their target with filepath.Join(basePath, path) on an untrusted, build-config-derived path. Join cleans "../" into a real parent traversal rather than rejecting it, so Retrieve/Delete/Exists/GetURL could read or unlink any file the process could reach and Store could write outside the artifact root. Added a resolve() helper that re-checks the cleaned absolute path against the root and returns ErrPathEscapesBase
- notifications/handlers.go: IDOR — GetProjectConfigs returned every notification config for any project ID to any authenticated user, and NotificationConfig.Config carries provider credentials (SMTP passwords, Slack/webhook tokens) serialized as `json:"config"`. Results are now filtered to the requesting user's own configs, which also avoids disclosing whether a project exists
- notifications/handlers.go: GetBuildLogs had no authentication check at all and returned notification logs for any build ID. Now requires an authenticated user and returns only entries whose owning config belongs to that user, with per-request ownership caching so a build with many logs does not re-query the same config

- credentials/encryption.go: removed the dead unsalted single-round SHA-256 HashPassword (zero callers; users.Service.HashPassword is the real Argon2id one)
- credentials/encryption.go: deriveUserKey salt string(rune(userID)) -> strconv.Itoa; the old form UTF-8-encoded the ID as a code point, collapsing surrogate-range and >0x10FFFF IDs to U+FFFD and giving distinct users the same derived key
- internal/config/config.go: removed the hardcoded "casci-default-encryption-key-change-this-in-production" fallback; key loading is now fallible and a CSPRNG or persist failure aborts startup
- internal/config/config.go: os.WriteFile error on the generated encryption key is now checked and returned
- users/auth.go: NewAuthManager now takes a caller-supplied secret (min 32 bytes) instead of generating a per-process random one, and no longer discards the rand error
- webhooks/handler.go: handleGitHub now rejects a missing X-Hub-Signature-256 (was fail-open); the hardcoded "your-webhook-secret" is gone, secret comes from WEBHOOK_SECRET, and an unconfigured secret rejects rather than accepts
- webhooks/handler.go: handleGitLab's `// TODO: Verify token` / `_ = token` no-op replaced with a constant-time X-GitLab-Token check returning 401
- audit/handlers.go: config.ParseBool used instead of strconv.ParseBool, with the correct (value, default) arity
- internal/config/config.go: getEnvBool no longer hand-rolls "true"/"1"/"yes"; it delegates to MustParseBool per PART 5
- go.mod, database/postgres.go: github.com/lib/pq (forbidden by PART 3) replaced with github.com/jackc/pgx/v5/stdlib; sql.Open driver name "postgres" -> "pgx"
- LICENSE.md: was WTFPL v2 with a personal copyright line; rewritten to MIT / Copyright (c) 2025 webappsgo per PART 2, with an Embedded Licenses table for all 12 direct dependencies (no GPL/AGPL/LGPL present) — flagged in TODO.AI.md for confirmation since relicensing is legally significant
- README.md: rewritten to the PART 1 template — badges, About, Official Site, Features, Production (Docker/Compose/Binary), Configuration with a full env-var table, API with curl -q -LSsf examples, Development, Disclaimer, License
- database/sqlite.go: duplicate user_credentials / project_credentials CREATE TABLE definitions removed
- Makefile: was space-indented ("missing separator", could not run) and had non-spec targets building a nonexistent ./src/cmd/casci on the host — rewritten to PART 26 exactly (build/local/release/docker/test/dev/clean, all via casjaysdev/go:latest Docker)
- src/: 75 of 81 Go files had zero indentation — ran gofmt -w, now 0 unformatted
- Dockerfile.dev, docker-compose.yml: moved to docker/ per PART 3
- API.md -> docs/api.md, QUICKSTART.md -> docs/installation.md per PART 3
- Makefile.old: stray .old duplicate at root, git rm'd
- release.txt: was 0.1.0-dev (semver+text hybrid), set to 0.1.0
- .gitignore: line 2 now the literal `ignoredirmessage` token; volumes/ added; removed a blanket .claude/ ignore that contradicted PART 3 (.claude/rules/ and settings.json are version-controlled)
- server/templates/: none of the mandatory partials (head, header, nav, footer, scripts) or components/toast.tmpl were wrapped in a `{{define}}` block, so ParseFS registered them under their filenames and every `{{template "head" .}}` in both layouts resolved to nothing. Go reports this only at execution time, so it never failed the build — every HTML page in the application returned 500. All six now define their spec-named block (AI.md 27633-27645)
- server/templates.go + server.go: the embedded static tree is rooted at "static", but the file server was handed the raw embed.FS after StripPrefix had already removed /static/ from the request path, so every stylesheet, script, image and the favicon 404'd. Re-rooted through fs.Sub; a new test walks every /static/ reference in the templates and stats it in the embedded FS
- server/handlers_web.go: the admin settings page was handed an untyped map with lowercase keys ("server", "port") while the template read capitalised field paths (.Config.Server.Port), plus .Config.Server.RateLimit.*, .SSL.LetsEncrypt.* and .Web.UI.Theme, none of which are backed by any config. The page returned 500 on every request. Replaced with the typed SettingsView/SettingsServerView/SettingsSSLView/SettingsDatabaseView/SettingsWebView, so a template/model mismatch is now a compile error, and the template was cut back to the settings that actually exist — the controls for unbuilt features (rate limiting, Let's Encrypt, UI theme) were removed rather than left as inert widgets
- server/server.go: server.ssl.min_version was parsed and defaulted (TLS1.2) but never applied — the listener hardcoded tls.VersionTLS12, so setting TLS1.3 silently did nothing. Added config.TLSVersion and ServerConfig.TLSMinVersion; the resolver falls back to 1.2 so a typo can never lower the floor
- server/test_embed.go: deleted — a leftover debug file with an exported TestEmbed function nobody called, a second //go:embed of the entire templates tree (duplicating it in the binary), an ignored ReadDir error and a bare fmt.Println
- server/wellknown.go: /server/security added — the human-readable Policy: target that security.txt has been advertising, rendering the new pages/security.tmpl with the contact channels in RFC 9116 preference order and every URL built per request (AI.md 16299-16351)
- server/wellknown.go: llms.txt added at both /.well-known/llms.txt and /llms.txt, advertising only public and authenticated API surface — admin, internal and metrics routes are omitted per the AI.md 16353-16416 inclusion table
- server/templates.go: the renderer parsed every page into one namespace, so the single "content" block name meant the last page parsed answered every route; handlers named templates by path while ParseFS registers by base name; four pages referenced a nonexistent "layout"; and RenderWithLayout emitted the content body and then the whole layout again. Each page now gets its own template set, RenderWithLayout is gone, and tests cover set isolation and reference resolution
- server/templates/ + static/js/: every interactive control used an inline onclick, which the nonce CSP the middleware already sets silently blocks — the entire UI was dead. All inline handlers replaced with data-action attributes and delegated listeners; modal.js dropped new Function() (CSP-blocked and an injection vector) for a callback Map and stopped building nodes from interpolated innerHTML (XSS); a test now fails the build if an inline handler reappears
- server/server.go: auth routes moved to /api/{v}/server/auth/* per AI.md:3213 (they were at /api/{v}/auth/*)
- server/server.go: /api/healthz unversioned alias added alongside /api/{v}/server/healthz, sharing the one handler (AI.md 19732-19744)
- server/templates/: all /admin* links repointed at /server/{{.AdminPath}}, /docs at /server/help, /swagger/ at /server/docs/swagger and /graphql/playground at /server/docs/graphql; admin.tmpl's nav was cut to the two routes actually mounted instead of linking five 404s
- server/handlers_web.go: the hardcoded "0.1.0"/"0.1.0-dev" version and "production" mode literals are gone; populatePageData now fills version, mode, admin path, API version, site name, tagline and the CSP nonce for every page, including healthz which previously rendered with an empty nonce
- static/js/app.js: logout() no longer POSTs to /api/v1/auth/logout, which never existed — sessions are stateless bearer tokens with no revocation list, so signing out clears the local credentials; settings and admin URLs now come from meta tags rather than hardcoded /api/v1 and /admin literals
- users/builds/projects handlers.go: writeJSON no longer discards the encode error
- internal/config: env vars are the unprefixed PART 5 names (PORT, LISTEN, MODE, DOMAIN, …) and paths carry {internal_org}, so the earlier CASCI_-namespace and missing-/etc/casapps findings no longer apply; server.yml and its loader exist
- server/server.go: verified free of duplicate ServeMux patterns — the earlier "server cannot start" finding no longer reproduces
- server/csrf.go, server/cors.go: verified wired into the middleware chain, so the dead-public-API finding is closed
