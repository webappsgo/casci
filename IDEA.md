## Project description

CASCI (CI/CD Application Server for Continuous Integration) is a complete CI/CD
platform delivered as a single static Go binary. It replaces Jenkins, GitHub
Actions, GitLab CI, CircleCI, Travis CI, and similar tools for teams and
individuals who want one deployable that runs existing CI/CD pipelines
unchanged, starts without hand-written configuration, and scales from a
laptop to a cluster of 1000+ build nodes.

The problem CASCI solves: current CI/CD tooling is fragmented (a different
product per platform), fragile (plugin dependency hell, breaking updates),
expensive (enterprise CI/CD billing), and a supply-chain risk (plugins and
third-party actions as an attack surface). CASCI's answer is one binary that
speaks every major CI/CD dialect natively, needs zero configuration to start,
has security scanning built in rather than bolted on, and is priced to run
for roughly $83/year.

## Project variables

project_name:  casci
project_org:   webappsgo
# FROZEN — never edit after first run
internal_name: casci
# FROZEN — never edit after first run
internal_org:  casapps
official_site: https://casci.casapps.us
repository:    https://github.com/webappsgo/casci

## Business logic

### Product scope & non-goals

CASCI must:
- Run as a single static binary with no required external services (embedded
  database by default; external database is opt-in, and when present the
  database — not a config file — is the source of truth).
- Work with zero operator-supplied configuration on first run and never
  require CLI flags for normal operation; operators can manage supported
  settings from the web UI, and cluster nodes share consistent configuration.
- Be self-healing: automatic error recovery, graceful degradation, and
  resistance to misconfiguration are product requirements, not aspirational.
- Scale from a single laptop instance up to clusters of 1000+ build nodes
  without a different deployment model.
- Target an operating cost around $83/year for a reference small deployment
  (this is a design constraint on resource usage and defaults, not a billing
  feature).
- Meet stated performance targets at scale: sub-100ms (p99) API response,
  build start under 2 seconds, sub-50ms log-streaming latency, and support
  for thousands of concurrent WebSocket connections and tens of thousands of
  builds/day on larger clusters.

CASCI must NOT:
- Require a plugin system to add functionality — plugin-equivalent behavior
  is implemented natively in the binary, not loaded from external code.
- Require any config file to be hand-edited for normal operation.
- Aggregate or bill for users' own cloud/infrastructure costs — CASCI
  provides the base platform; a user's own cloud accounts are used and paid
  for directly by that user, with no bill aggregation by CASCI.

### Compatibility requirements

- **Jenkins**: must be 100% compatible with the Jenkins REST API (including
  the Blue Ocean API and `jenkins-cli.jar`-style CLI operations) and must
  parse existing Jenkins XML job configuration and Jenkinsfiles (declarative
  and scripted) unchanged. Jenkins plugin APIs must respond successfully
  (functionality implemented natively) rather than by loading real plugins,
  covering the functionality of the most common Jenkins plugins in use.
- **CI/CD format parity** — the following formats must run unchanged, at the
  compatibility level noted:
  - Jenkinsfile / Blue Ocean / multibranch pipelines — 100%
  - GitHub Actions (`.github/workflows/*.yml`, all action types, matrix
    builds, secrets/variables) — 100%
  - GitLab CI (`.gitlab-ci.yml`, includes, extends/anchors, DAG pipelines) —
    100%
  - Travis CI (`.travis.yml`, build matrices, stages) — 100%
  - CircleCI (`.circleci/config.yml`, workflows, orbs) — 99%
  - Azure Pipelines (multi-stage, templates) — 99%
  - Bitbucket Pipelines (pipes) — 90%
  - Drone CI, Buildkite — 95% · Tekton, Woodpecker — 90% · CodeFresh — 85%
  - If a repository has none of the above, CASCI must be able to
    auto-generate a pipeline from project detection.
- **Cloud providers** — must support provisioning/destroying build capacity
  on AWS, Google Cloud, Azure, and Oracle Cloud, plus on-demand hourly/burst
  providers (Vultr, Hetzner Cloud, DigitalOcean, Linode). Provisioned
  instances must never sit idle — they are destroyed as soon as a build
  completes or times out.
- **Compliance standards** — must support operating under HIPAA, SOX,
  PCI-DSS, GDPR, FedRAMP, and ISO 27001, including each standard's audit-log
  retention expectations (HIPAA 6 years, SOX 7 years, PCI-DSS 1 year access
  logging) and GDPR rights (data minimization, right to deletion, data
  portability).

### Roles & permissions

- **Administrator** — manages server configuration, nodes, and settings
  only. Cannot create builds and cannot access other users' data. Default
  username `administrator` (changeable). The primary server administrator is
  created through first-run setup; regular-user registration never grants
  server-administrator access.
- **Regular user** — creates and manages their own projects, runs builds,
  manages their own infrastructure connections (cloud accounts, private
  servers, storage, optional databases), and can only see their own
  projects, logs, and metrics.
- **Service account** — API-only access with limited, token-scoped
  permissions and a full audit trail; not a human login.
- Regular-user access is owner-scoped; regular-user roles do not grant
  server-administrator access. Server-administrator roles follow AI.md PART 17.

### Multi-tenancy & trust boundaries

- **Project isolation** — a user must never be able to see another user's
  projects, credentials, workspaces, or build queues.
- **Build isolation** — each build runs in its own container/VM with network
  isolation and no storage shared with other users' builds; environments are
  clean per build.
- **Resource isolation** — per-user quotas and fair scheduling are required;
  no user's workload may starve another's.
- **Data isolation** — per-user encryption keys, no cross-user queries, and
  independent backup/retention per user.
- **Secrets are untrusted once outside their vault** — credentials and
  secrets must never appear in logs or build artifacts, must be injected
  just-in-time, and must be cleaned up automatically after use.
- **User-supplied infrastructure is trusted only for that user** — cloud
  credentials, private servers, and storage a user connects to CASCI are
  used exclusively for that user's builds and are never accessible to other
  tenants or to CASCI operators outside the scope the user granted.
- Every build's output must be verifiable: binaries are signed, dependencies
  get an SBOM, and every build is scanned for vulnerabilities, secrets, and
  license issues by default (with a configurable block-on-critical policy).

### Spending & budget controls

- Users set their own monthly and per-provider spending limits for
  cloud/hourly infrastructure they connect; CASCI must check budget before
  provisioning and must offer an emergency "destroy all hourly instances"
  control. Because provider billing APIs lag in real time, budget
  enforcement is explicitly best-effort, not a hard guarantee, and this
  limitation must be disclosed to the user.

### Data model & sensitivity

- Users own projects, pipeline definitions, builds, logs, artifacts, infrastructure connections, credentials, budgets, and notification settings. Administrators own server configuration and node/server settings; they do not own or browse user project data.
- Account identifiers, password hashes, session records, API tokens, provider credentials, private keys, pipeline contents, source metadata, build logs, artifacts, security scan results, audit records, and per-user encryption keys are sensitive. Provider spend limits and usage data are private to the user who connected that provider.
- Secrets are injected only for the owning user's build, are never emitted in logs or artifacts, and are removed after use. Build workspaces and queues are isolated between users.
- Build outputs are signed, accompanied by an SBOM, and scanned for vulnerabilities, secrets, and license issues by default.

### Trust boundaries & external services

- Repository content, pipeline files, build scripts, dependencies, webhook payloads, and build outputs are untrusted input, including when supplied by a user's own repository.
- Each user's cloud accounts, private servers, storage, and optional databases are trusted only for that user's builds. Their credentials and resulting resources are not shared with other users or used outside the scope the user granted.
- Git hosting and CI compatibility protocols, cloud APIs, hourly compute providers, user-supplied servers, and user-supplied storage are external dependencies. Their outages or rejected requests must not be reported as successful operations or weaken user/data isolation. Capacity provisioning must not proceed when provider authorization or the budget pre-check fails; a provider's delayed billing data means budget enforcement remains best-effort as stated above.
- CASCI does not aggregate or bill users for their infrastructure charges. Users pay their providers directly.

### Threat model & abuse cases

- **Assets:** user credentials and secrets; source and pipeline content; build workspaces, logs, and artifacts; provider accounts and spending; user identity and project ownership; audit evidence; server and cluster availability.
- **Attacker inputs:** anonymous HTTP/API traffic, authenticated users, repository and dependency contents, webhooks, pipeline scripts, artifacts, and responses from external providers.
- **Attacker goals:** read another user's projects or secrets, execute code outside an isolated build, steal credentials from logs/artifacts, tamper with build results, consume another user's quota, exhaust shared capacity, or evade required scans and audit records.
- **Required defenses:** strict per-user authorization and encryption boundaries; clean isolated build environments; just-in-time secret injection and cleanup; safe handling of untrusted files and scripts; resource quotas and fair scheduling; authenticated and validated webhooks; signed outputs, SBOMs, default vulnerability/secret/license scanning; immutable audit trails; and fail-closed authorization and provisioning checks. An external provider failure is surfaced as a failed or degraded operation, never as a success.

### Security decisions & exceptions

- There are no intentionally weaker authentication, authorization, transport, build-isolation, secret-handling, or abuse-control modes. Compatibility with an existing CI/CD format does not grant its pipeline access to another user's resources or relax build isolation.
- Spending limits are explicitly best-effort because provider billing APIs can lag. This is a documented limitation, not a guarantee that a limit cannot be exceeded.
- SAML service-provider support uses `github.com/supatype/saml` v0.7.0, a maintained fork of `crewjam/saml` based on upstream commit `3465403`. It carries the namespace-binding and assertion-validation fixes for GO-2026-6909, plus updated `goxmldsig` v1.6.1 and replay-cache support. This replaces the vulnerable upstream `crewjam/saml` module; review the Go vulnerability database findings for the exact pinned dependency graph before release.

### API routes

These are CASCI product routes beyond the common server routes defined by AI.md. User-owned resources require the owning user's authentication; server-wide administration routes require a server administrator. Requests and responses use the canonical API envelopes defined by AI.md.

| Method | Path | Auth | Purpose and request/response summary |
|---|---|---|---|
| GET, POST | `/api/{api_version}/projects` | User | List the caller's projects; create a project from repository/source and project settings. |
| GET, PATCH, DELETE | `/api/{api_version}/projects/{project_id}` | Owning user | Read, update, or delete one owned project. |
| GET | `/api/{api_version}/projects/{project_id}/pipelines` | Owning user | List detected or configured pipelines and their source format. |
| GET | `/api/{api_version}/projects/{project_id}/pipelines/{pipeline_id}` | Owning user | Read a pipeline definition and its validation/detection result. |
| GET, POST | `/api/{api_version}/projects/{project_id}/builds` | Owning user | List builds or trigger a build with revision, ref, and permitted inputs; return the build record. |
| GET | `/api/{api_version}/builds/{build_id}` | Owning user | Read build state, timing, status, and scan summary. |
| POST | `/api/{api_version}/builds/{build_id}/cancel` | Owning user | Cancel an active build and return its resulting state. |
| GET | `/api/{api_version}/builds/{build_id}/logs` | Owning user | Read or stream the build's logs. |
| GET | `/api/{api_version}/builds/{build_id}/artifacts` | Owning user | List signed build artifacts and their metadata. |
| GET | `/api/{api_version}/builds/{build_id}/artifacts/{artifact_id}` | Owning user | Download an owned build artifact with safe content headers. |
| GET, POST | `/api/{api_version}/infrastructure/connections` | User | List or add the caller's cloud, private-server, or storage connection; secrets are write-only. |
| GET, PATCH, DELETE | `/api/{api_version}/infrastructure/connections/{connection_id}` | Owning user | Read safe connection metadata, update, or revoke one owned connection. |
| GET, PUT | `/api/{api_version}/budgets` | User | Read or set the caller's monthly and per-provider spending limits. |
| POST | `/api/{api_version}/infrastructure/hourly-instances/destroy-all` | User | Emergency destruction of the caller's hourly/burst instances; returns per-instance outcomes. |
| GET | `/api/{api_version}/security/reports` | Owning user | List security scan reports for the caller's builds. |
| GET | `/api/{api_version}/security/reports/{report_id}` | Owning user | Read one scan report belonging to the caller's build. |
| POST | `/api/{api_version}/webhooks/{provider}` | Provider signature | Accept a provider-signed event and enqueue the associated user's build without exposing credentials. |
| GET | `/api/{api_version}/nodes` | Server administrator | List registered worker nodes and their health. |
| GET, PUT, PATCH, DELETE | `/api/{api_version}/nodes/{node_id}` | Server administrator | Read, update, or delete a worker node. |
| POST | `/api/{api_version}/nodes/register` | Node registration token | Register a worker node. |
| POST | `/api/{api_version}/nodes/token` | Server administrator | Generate a node registration token. |
| POST | `/api/{api_version}/nodes/{node_id}/heartbeat` | Node credential | Submit node health and capacity. |
| POST | `/api/{api_version}/nodes/{node_id}/drain` | Server administrator | Drain a worker node from new work. |
| GET, POST | `/api/{api_version}/server/{admin_path}/backup` | Server administrator | List safe archive metadata or create and verify a full backup; an optional password is write-only. |
| GET, PUT | `/api/{api_version}/server/{admin_path}/backup/settings` | Server administrator | Read or update scheduled-backup enablement, hourly incremental enablement, and validated retention policy. |
| GET, DELETE | `/api/{api_version}/server/{admin_path}/backup/{id}` | Server administrator | Read metadata or delete an archive; deletion is refused when an incremental depends on it. |
| GET | `/api/{api_version}/server/{admin_path}/backup/{id}/download` | Server administrator | Download an archive as an attachment. |
| POST | `/api/{api_version}/server/{admin_path}/backup/test` | Server administrator | Verify an existing archive's checksum, manifest, and encryption using its filename and optional password. |
| POST | `/api/{api_version}/server/{admin_path}/backup/restore` | Server administrator | Verify and restore an archive, then restart through the supervisor; requires the backup password when encrypted and typed filename confirmation. |
| GET | `/api/{api_version}/security/reports` | Owning user | List the caller's security scan reports. |
| GET | `/api/{api_version}/security/reports/{report_id}` | Owning user | Read one owned security report. |
| GET, PUT | `/api/{api_version}/security/config` | Server administrator | Read or update scanning configuration. |
| GET | `/api/{api_version}/security/statistics` | Server administrator | Read aggregate scan statistics. |
| GET, POST | `/api/{api_version}/notifications` | User | List or create the caller's notification integrations. |
| GET, PUT, PATCH, DELETE | `/api/{api_version}/notifications/{notification_id}` | Owning user | Read, update, or remove one notification integration. |
| POST | `/api/{api_version}/notifications/{notification_id}/test` | Owning user | Send a test notification. |
| GET | `/api/{api_version}/notifications/types` | User | List supported notification channels. |
| GET | `/api/{api_version}/notifications/events` | User | List supported notification event types. |
| GET, POST | `/api/{api_version}/credentials/user` | User | List or create the caller's encrypted credentials. |
| GET, PUT, PATCH, DELETE | `/api/{api_version}/credentials/user/{credential_id}` | Owning user | Read safe metadata, update, or remove one user credential. Secret values are write-only. |
| GET, PUT, PATCH, DELETE | `/api/{api_version}/credentials/project/{project_id}/{credential_id}` | Owning user | Read safe metadata, update, or remove one project credential. Secret values are write-only. |
| GET | `/api/{api_version}/audit/events` | User | Query the caller's audit events. |
| GET | `/api/{api_version}/audit/status` | Server administrator | Read audit subsystem status. |
| GET, PUT, PATCH | `/api/{api_version}/audit/config` | Server administrator | Read or update audit retention configuration. |
| POST | `/api/{api_version}/audit/cleanup` | Server administrator | Run audit retention cleanup. |
| GET, POST | `/api/json` | Jenkins-compatible client authentication | Serve the Jenkins-compatible API discovery resource. |
| GET, PUT, PATCH | `/api/{api_version}/compliance/config` | Server administrator | Read or update compliance configuration. |
| GET | `/api/{api_version}/compliance/modes` | Server administrator | List available compliance modes. |
| POST, PUT | `/api/{api_version}/compliance/mode` | Server administrator | Set the active compliance mode. |
| GET | `/api/{api_version}/compliance/preset` | Server administrator | Read the active compliance preset. |
| POST | `/api/{api_version}/compliance/check` | Server administrator | Run a compliance check. |
| POST | `/users/settings` | User session + CSRF | Submit the account settings page as a standard HTML form when JavaScript is unavailable. |
| GET, POST | `/users/security` | User session; POST requires CSRF | Review account security and change the password with a standard HTML form. |
| GET, POST | `/users/security/sessions` | User session; POST requires CSRF | Review API sessions and revoke one session from a standard HTML form. |

Compatibility interfaces are additional to these native routes: Jenkins REST and Blue Ocean resources, Jenkins crumb and CLI operations; GitHub Actions, GitLab CI, Travis CI, CircleCI, Azure Pipelines, Bitbucket Pipelines, Drone, Buildkite, Tekton, Woodpecker, and CodeFresh pipeline formats; and the provider-specific webhook event routes required to trigger those pipelines. Each protocol must preserve the compatibility level stated above while enforcing the same user ownership, build isolation, and secret boundaries.
