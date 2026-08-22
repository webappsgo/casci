## Project description

CASCI (CI/CD Application Server for Continuous Integration) is a complete CI/CD
platform delivered as a single static Go binary. It replaces Jenkins, GitHub
Actions, GitLab CI, CircleCI, Travis CI, and similar tools for teams and
individuals who want one deployable that runs existing CI/CD pipelines
unchanged, requires no configuration files, and scales from a laptop to a
cluster of 1000+ build nodes.

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
- Work with zero configuration on first run and never require CLI flags to
  operate; all configuration is database-driven and manageable from the web
  UI.
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
  username `administrator` (changeable). The first user created on a fresh
  install becomes the administrator.
- **Regular user** — creates and manages their own projects, runs builds,
  manages their own infrastructure connections (cloud accounts, private
  servers, storage, optional databases), and can only see their own
  projects, logs, and metrics.
- **Service account** — API-only access with limited, token-scoped
  permissions and a full audit trail; not a human login.
- Permission model is intentionally simple: admin or user, no general-purpose
  RBAC. Users own and are isolated within their own resources.

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
