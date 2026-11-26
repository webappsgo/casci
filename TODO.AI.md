# CASCI AI.md Alignment

Implementation order follows the dependencies in AI.md: foundation and configuration; identity and data boundaries; server/API and browser flows; pipeline execution and workers; operations and integrations; client/agent; docs and release verification.

## Critical (P0)

- [x] Make credential and notification handlers derive user identity only from the typed authentication context; remove spoofable `X-User-ID` fallback handling.
- [x] Restrict API user authentication to Bearer credentials, ignore `X-API-Token`, prevent invalid Authorization headers from falling back to browser cookies, and return the required Bearer challenge on failure.
- [x] Limit CSRF validation to server-rendered page routes; backend and compatibility routes bypass the browser form token check, while malformed and non-Bearer Authorization values do not bypass page CSRF.
- [ ] Fix security and tenant-isolation defects in existing handlers, repositories, middleware, credential storage, build execution, logs, and artifacts against IDEA.md and AI.md PARTS 10, 11, 14, 16, and 34.
- [ ] Complete the first-run server-admin flow, setup-token lifecycle, authentication/session management, and admin-only route protection required by PART 17.
- [ ] Implement isolated, clean-per-build execution, secret injection and cleanup, resource quotas, fair scheduling, artifact signing, SBOM generation, and default vulnerability/secret/license scans required by IDEA.md and PARTS 10, 11, 14, and 34.
- [ ] Replace placeholder/incomplete API behavior with the required versioned REST surface, canonical response/error envelopes, authorization, request IDs, validation, rate limits, and synchronized OpenAPI/GraphQL/MCP routes from PART 14.
- [x] Add owner-scoped user API token list, create, inspect, revoke endpoints, hashed storage, and the `/users/tokens` server-rendered no-JavaScript page.
- [ ] Complete the server-rendered public and admin page flows, including no-JavaScript forms, CSRF, accessibility, shared shell, preferences/help hubs, page registry, and required admin pages from PARTS 16 and 17.

## High (P1)

- [x] Add the required application-mode package, canonical mode aliases, `MODE`/`DEBUG` precedence, and startup profiling/logging integration from PART 6.
- [x] Remove prohibited `src/pkg/` and `src/internal/` trees, move packages under `src/`, and extract path resolution into `src/paths`.
- [x] Embed the release version, commit ID, and UTC build timestamp in server, client, and agent binaries; use cross-platform service helper definitions so the required OS/architecture matrix compiles.
- [ ] Complete the remaining exact PART 3 layout, including required client/common/mode/ssl/scheduler/service/admin packages and the server handler/service/model/store/template structure.
- [ ] Complete platform-aware path resolution, mode detection, static binary metadata, the full server/CLI/agent command interface, and runtime lifecycle from PARTS 4, 6, 7, 8, and 24.
- [ ] Implement the complete YAML configuration system, drop-in files, database-backed settings, environment/flag precedence, validation, hot reload, secret-file handling, and admin UI from PARTS 5 and 12.
- [ ] Complete the embedded database schema, PostgreSQL cluster mode, migrations-by-idempotent-startup, connection management, and cluster node protocol from PART 10.
- [ ] Implement all required server security controls, security headers, CSRF/CORS, CSP, audit logging, secret redaction/encryption/rotation, compliance retention, blocklists, GeoIP, and vulnerability data management from PARTS 11, 20, and 34.
- [ ] Implement project-format detection and full required CI/CD format parsing/execution, Jenkins compatibility, webhook validation, pipeline scheduling, build logs, artifacts, cancellation, and status reporting from IDEA.md and PARTS 14, 19, and 33.
- [ ] Complete server-rendered pages and matching APIs for projects, builds, nodes, credentials, notifications, security, audit, compliance, preferences, and all admin configuration areas; add page-conformance and route-drift coverage.
- [ ] Implement email templates, SMTP detection and delivery, WebUI notification center, preferences, and required security/account messages from PART 18.
- [ ] Implement the scheduler and required periodic work, failure handling, and admin controls from PART 19.
- [ ] Implement system and build metrics, Prometheus output, admin views, and required health/readiness/liveness behavior from PARTS 13 and 21.
- [ ] Implement encrypted backup, restore validation, retention, scheduling, and UI/API management from PART 22.
- [ ] Implement secure update discovery, staged update download/verification/apply, rollback safety, and user-facing update state from PART 23.
- [ ] Implement native service installation and management for supported operating systems, privilege escalation prompts, and safe uninstall behavior from PARTS 24 and 25.
- [ ] Implement Let's Encrypt and local TLS certificate lifecycle, renewal, deployment, and UI/API management from PART 15.
- [ ] Implement Tor auto mode and opt-in I2P overlay behavior with graceful degradation from PART 32.
- [ ] Implement the required CLI client and the CASCI build-node agent, including shared API models, auth, translations, output modes, and cross-platform binaries from PART 33.
- [ ] Implement i18n, locale integrity checks, translated server/client strings, language preferences, and WCAG 2.1 AA accessibility from PART 31.
- [ ] Complete multi-user PART 34 because IDEA.md requires regular users, user-owned projects, scoped credentials, and per-user isolation; record the activation in SPEC.md.

## Medium (P2)

- [ ] Complete all required server pages and admin sections to AI.md PARTS 16 and 17, including About, Privacy, Contact, Help, MCP help, Terms, DPO, security policy/thanks, login/setup, dashboard, settings, logs, branding, pages, SSL, network, email, scheduler, backup, metrics, audit, and user security flows.
- [ ] Replace stale or incomplete public pages, admin pages, templates, static assets, and JavaScript with the shared page shell, theme tokens, responsive layout, semantic markup, and accessible controls specified in PART 16.
- [ ] Complete PWA manifest, service worker, offline shell, update lifecycle, install behavior, and privacy-safe cache rules from PART 16.
- [ ] Complete the required root project files, licenses and dependency attribution, `.gitignore`, `.dockerignore`, scripts, and exact allowed directory structure from PARTS 2 and 3.
- [ ] Complete Makefile targets, version metadata, all eight OS/architecture builds, release checksums/SBOM/provenance, Docker production/development/test images, Compose files, and entrypoint behavior from PARTS 7, 26, and 27.
- [ ] Complete GitHub, GitLab, Gitea, Forgejo, and Jenkins CI/CD definitions, security scanning, pinned actions, workflow policy checks, release flows, and provider-specific parity from PART 28.
- [ ] Complete container, Incus, and browser end-to-end test runners; achieve the specified coverage gate; add regression, security, conformance, i18n, API, and compatibility tests from PART 29.
- [x] Reconcile IDEA.md with the required three-section structure and document data sensitivity, external-service trust/failure assumptions, threat/abuse cases and defenses, security exceptions, and the project-specific API route inventory; SPEC.md activates PART 34 only.

## Low (P3)

- [ ] Bring README.md, MkDocs navigation, ReadTheDocs configuration, and every docs page current with implemented behavior; add all required docs and translation guidance from PART 30.
- [ ] Synchronize OpenAPI, GraphQL, MCP, CLI help, templates, translations, and docs with the final implemented routes and behavior.
- [ ] Verify the full FINAL CHECKPOINT, all eight build targets, Docker images, documentation build, CI workflow graphs, security gates, and post-push provider run.
