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

## Source-control webhooks

The shared receiver is `POST https://casci.casapps.us/webhook`. The handler
recognizes GitHub, GitLab, Bitbucket, and Gitea headers. GitHub deliveries
require a valid `X-Hub-Signature-256`; GitLab uses `X-GitLab-Token`. The
configured shared secret is `server.security.webhook_secret`.

Webhook event support and repository-to-project resolution are incomplete.
Verify the response and application state before assuming a delivery queued a
build. Do not treat a webhook `200` as proof of completed pipeline execution.

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
REST, Blue Ocean, CLI, or plugin API. It also does not provide cloud-provider
capacity provisioning, external identity providers, or instance federation.

## Discovery

The server generates `/.well-known/security.txt` and serves `/.well-known/llms.txt`
plus `/llms.txt`. Unimplemented well-known paths return `404`; no WebFinger,
OAuth/OIDC discovery, mobile app associations, or Tor/I2P discovery is enabled
in the current server.
