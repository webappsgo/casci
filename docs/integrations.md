# Integrations

Integration support is partial. A declared parser, provider type, or endpoint
does not imply full compatibility or a complete production workflow.

## Pipeline file parsers

The parser detects these formats in a repository, in the order shown:

| Format | Detection path |
|--------|----------------|
| Jenkins | `Jenkinsfile` |
| GitHub Actions | `.github/workflows/` |
| GitLab CI | `.gitlab-ci.yml` |
| CircleCI | `.circleci/config.yml` |
| Travis CI | `.travis.yml` |
| Azure Pipelines | `azure-pipelines.yml` |
| Bitbucket Pipelines | `bitbucket-pipelines.yml` |
| Drone CI | `.drone.yml` |
| Buildkite | `buildkite.yml` |

These are parser implementations, not claims of complete runtime parity. For
example, the Jenkinsfile parser currently handles only a minimal declarative
shape. Unsupported syntax or provider-specific features may be ignored or fail.
The parser does not yet implement every format listed in IDEA.md.

Recognized project-detection defaults and top-level pipeline environment values
are passed to the build container in deterministic order. CASCI-owned variables
such as `BUILD_ID`, `PROJECT_ID`, and `COMMIT_SHA` cannot be overridden by the
repository. Job-scoped environment and provider secret expressions are not yet
resolved with their full provider-specific semantics.

## External identity providers

Administrators configure OIDC, LDAP, and SAML providers under
`/server/{admin_path}/security/auth/{oidc,ldap,saml}`. The browser flows are
OIDC authorization-code with PKCE, LDAP bind/search, and signed SAML
authentication. Provider settings can also be managed under
`/api/{api_version}/server/{admin_path}/security/auth/{type}/providers`;
secret fields are write-only and stored encrypted in `secrets.yml`.

Browser entry routes are `/server/auth/oidc/{provider}`,
`/server/auth/ldap/{provider}`, and `/server/auth/saml/{provider}`.
OIDC and SAML support API initiation plus one-time request-ID polling and
exchange routes. LDAP API login is a JSON POST. SAML publishes SP metadata
and supports SP-initiated logout where the IdP has configured SLO.

Existing local accounts are linked by stable provider subject. New external
identities use the server's onboarding flow and must select a local username.
Administrators must configure explicit group/role mappings before relying on
external group membership for administrator access. Provider deletion removes
the provider's account-link records. Federation between CASCI instances is
not implemented.

## Source-control webhooks

The versioned receiver is `POST /api/v1/webhooks/{provider}`; the legacy
`POST /webhook` path remains available. The provider path must match the
provider headers. GitHub push and opened/synchronized
pull request deliveries require `X-Hub-Signature-256`; GitLab push and merge
request events use `X-GitLab-Token`; Bitbucket push and pull request events use
`X-Hub-Signature: sha256=<hex>`; Gitea push and pull request events use
`X-Gitea-Signature` with the raw HMAC-SHA256 hex digest. All use
`server.security.webhook_secret`. Request bodies are capped at 1 MiB.

Matching repository URLs trigger queued builds and receive `202 Accepted`.
Ping, ignored action, branch deletion, and unsupported event deliveries receive
a successful empty or small response without queuing a build. Repository
matching normalizes HTTPS, SSH, and `.git` URL spellings. A queued build does
not prove that its pipeline completed successfully. The queue applies global
and per-user pending limits, runs at most one build per user at a time, and
dispatches queued owners round-robin.

## Notifications

The notification package defines these channel identifiers. The active API
list is returned by `GET /api/v1/notifications/types`; availability and
delivery behavior can vary by implementation.

| Group | Channel identifiers |
|-------|---------------------|
| Chat | `slack`, `discord`, `teams`, `telegram`, `matrix`, `irc`, `mattermost` |
| Email | `email` |
| Issue tracking | `jira`, `github_issue`, `gitlab_issue`, `linear`, `asana` |
| Commit status | `github_status`, `gitlab_status`, `bitbucket_status`, `gitea_status` |
| Incident response | `pagerduty`, `opsgenie`, `victorops` |
| Other | `webhook`, `script` |

Create and manage configurations through `/api/v1/notifications`. Configure
secrets only in the instance, not in a repository. The API's test route is
`POST /api/v1/notifications/test`.

## API and protocol endpoints

| Route | Status |
|-------|--------|
| `/api/v1/server/swagger` | OpenAPI JSON route; generated document is incomplete |
| `/server/docs/swagger` | Swagger UI |
| `/api/v1/server/graphql` | GraphQL endpoint; resolver coverage is incomplete |
| `/server/docs/graphql` | GraphQL playground |
| `/api/v1/server/mcp` | Authenticated MCP Streamable HTTP endpoint; tool coverage is partial |
| `/api/json` | Small Jenkins-shaped response; not the Jenkins REST API |
| `/crumbIssuer/api/json` | Jenkins-shaped crumb handler; not full Jenkins CSRF compatibility |

CASCI does not currently load Jenkins plugins or implement the full Jenkins
REST, Blue Ocean, CLI, or plugin API. Users can manage encrypted cloud, private-server, and storage connection metadata, and set best-effort monthly budgets. CASCI does not yet synchronize provider billing data or provision cloud capacity. The hourly-instance destroy-all endpoint returns `503` until a provider cleanup backend is available. External OIDC, LDAP, and SAML identity providers are supported for browser and API sign-in; see the provider setup and limitations in [Security](security.md) and [Configuration](configuration.md). Instance federation is not provided.

## Discovery

The server generates `/.well-known/security.txt` and serves
`/.well-known/llms.txt` plus `/llms.txt`. It also exposes `GET
/api/autodiscover`, which reports the API version, public origins, and
published CLI/agent binary versions and checksums. Private and loopback hosts
are omitted from autodiscovery. Binary download authorization depends on the
binary type and the configured download policy; see [Configuration](configuration.md)
and [Agent management](agents.md).

`GET /manifest.json` publishes the PWA manifest and `GET /sw.js` serves the
versioned service worker. The worker caches same-origin static assets and
successful public pages explicitly marked eligible for offline use; it does
not cache API responses.

Unimplemented paths under `/.well-known/` return `404`. WebFinger, mobile app associations, and Tor/I2P discovery are not enabled. OIDC provider discovery is fetched from the configured identity provider and is not published by this server.
