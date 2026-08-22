# Integrations

## External Identity

**None enabled.** CASCI authenticates against its own user store. There is no OIDC, OAuth, LDAP, Active Directory, or SAML integration in this build, and no third-party sign-in buttons. Accounts are created through `POST /api/v1/auth/register` or the web UI. See [Security](security.md).

## Discovery & Protocol Endpoints

`/.well-known/security.txt` is served, generated per request from `web.security.report_url`, `web.security.contact`, and `web.security.expires`. See [Security](security.md).

Nothing else in that namespace is: no `change-password`, no WebFinger, no OpenID provider metadata, no MTA-STS, no autodiscover. Those paths return `404` and are never redirected.

There are no native app association files either — no Android `assetlinks.json`, no Apple `apple-app-site-association`. CASCI has no mobile app to associate with.

The discovery surface that does exist is developer-facing rather than protocol-level:

| Endpoint | Purpose |
|----------|---------|
| `/api/v1/server/swagger` | OpenAPI specification, JSON (alias: `/api/swagger`) |
| `/api/v1/server/swagger.yaml` | OpenAPI specification, YAML (alias: `/api/swagger.yaml`) |
| `/server/docs/swagger` | Swagger UI |
| `/api/v1/server/graphql` | GraphQL endpoint (alias: `/api/graphql`) |
| `/server/docs/graphql` | Interactive GraphQL playground |
| `/api/json` | Jenkins-compatible server description |
| `/crumbIssuer/api/json` | Jenkins-compatible CSRF crumb issuer |

The `v1` segment comes from `server.api_version` and is not a literal. The GraphQL endpoint is routed but its resolvers are unimplemented stubs — see [API Reference](api.md).

## Federation

**None.** CASCI instances do not federate. Distributing work across machines is done with build nodes joined to a single instance, not by linking instances together. See the Build Nodes section below.

## Source Control Webhooks

One receiver handles every supported provider. The provider is identified from the request headers and the payload is verified before anything is queued.

```
POST https://casci.casapps.us/webhook
```

| Provider | Detection header | Verification | Events handled |
|----------|------------------|--------------|----------------|
| GitHub | `X-GitHub-Event` | HMAC-SHA256 in `X-Hub-Signature-256`, checked against the webhook secret | `push`, `pull_request`, `ping` |
| GitLab | `X-GitLab-Event` | `X-GitLab-Token` compared against the webhook secret | `Push Hook`, `Merge Request Hook` |
| Gitea | `X-Gitea-Event` | Delegated to the GitHub handlers, which require `X-Hub-Signature-256` | `push`, `pull_request` |
| Bitbucket | `X-Event-Key` | None | none — see below |

The shared secret is `server.security.webhook_secret` in `server.yml`, set from the admin panel or by editing the file. It is not an environment variable; exporting `WEBHOOK_SECRET` has no effect.

Set it before exposing the endpoint. A GitHub delivery arriving without a signature header is rejected rather than trusted.

Unrecognized event types are acknowledged with `200` and ignored.

!!! warning "Deliveries are verified but do not queue builds yet"
    The receiver resolves the target project by matching the payload's repository URL, and that lookup is not implemented — it returns an error for every input. A correctly signed GitHub, GitLab, or Gitea push therefore passes verification and then responds `404 Project not found` rather than starting a build.

    Bitbucket is further behind: `repo:push` and `pullrequest:*` are matched and answered `200` with no signature check and no action taken. Do not treat a `200` from a Bitbucket delivery as evidence that anything ran.

    Until this lands, trigger builds through `POST /api/v1/projects/{id}/builds`.

### Configuring the Provider

Point the provider at `https://<your-host>/webhook`, choose `application/json` as the content type, and set the secret to the same value as `server.security.webhook_secret`.

```bash
# Verify the endpoint is reachable before wiring up the provider
curl -q -LSsf -X POST https://casci.casapps.us/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: ping" \
  -H "X-Hub-Signature-256: sha256=<signature>" \
  -d '{}'
```

## Pipeline Format Compatibility

CASCI reads pipelines written for other CI systems and converts them to one internal model. Existing repositories run unchanged; no migration file is required.

| Format | Detected from |
|--------|---------------|
| Jenkins | `Jenkinsfile` |
| GitHub Actions | `.github/workflows/` |
| GitLab CI | `.gitlab-ci.yml` |
| CircleCI | `.circleci/config.yml` |
| Travis CI | `.travis.yml` |
| Azure Pipelines | `azure-pipelines.yml` |
| Bitbucket Pipelines | `bitbucket-pipelines.yml` |
| Drone CI | `.drone.yml` |
| Buildkite | `buildkite.yml` |

Detection stops at the first match, in the order listed. When a repository ships no pipeline file at all, the project language is auto-detected and a sensible default pipeline is used.

## Jenkins Client Compatibility

Existing Jenkins tooling can talk to CASCI without modification through `/api/json` and `/crumbIssuer/api/json`. This is a compatibility surface for clients, not a Jenkins plugin host — Jenkins plugins do not run on CASCI.

## Notifications

Configure delivery targets per user or per project at `/api/v1/notifications`. Query the live provider list with `/api/v1/notifications/types` and the event list with `/api/v1/notifications/events`.

### Chat

| Type | Service |
|------|---------|
| `slack` | Slack |
| `discord` | Discord |
| `teams` | Microsoft Teams |
| `telegram` | Telegram |
| `matrix` | Matrix |
| `irc` | IRC |
| `mattermost` | Mattermost |

### Email

| Type | Service |
|------|---------|
| `email` | SMTP |

SMTP settings live on the notification config itself — `smtp_host`, `smtp_port`, `smtp_user`, and `smtp_password` — not in `server.yml` and not in the environment. There are no `SMTP_*` environment variables; exporting them has no effect.

### Issue Trackers

| Type | Service |
|------|---------|
| `jira` | Jira |
| `github_issue` | GitHub Issues |
| `gitlab_issue` | GitLab Issues |
| `linear` | Linear |
| `asana` | Asana |

### Commit Status

| Type | Service |
|------|---------|
| `github_status` | GitHub commit status |
| `gitlab_status` | GitLab commit status |
| `bitbucket_status` | Bitbucket build status |
| `gitea_status` | Gitea commit status |

### Incident Response

| Type | Service |
|------|---------|
| `pagerduty` | PagerDuty |
| `opsgenie` | Opsgenie |
| `victorops` | VictorOps / Splunk On-Call |

### Generic

| Type | Behavior |
|------|----------|
| `webhook` | POST a JSON payload to any URL |
| `script` | Run a local script |

### Events

| Event | Fires when |
|-------|-----------|
| `build.started` | A build begins |
| `build.success` | A build completes successfully |
| `build.failure` | A build fails |
| `build.cancelled` | A build is cancelled |
| `deploy.started` | A deployment begins |
| `deploy.success` | A deployment succeeds |
| `deploy.failure` | A deployment fails |
| `security.alert` | A security scan produces a blocking finding |

Send a test delivery before relying on a target:

```bash
curl -q -LSsf -X POST https://casci.casapps.us/api/v1/notifications/test \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id":1}'
```

Provider credentials are stored in the encrypted credential vault and are never returned in an API response.

## Build Nodes

Remote build nodes join a single instance with a registration token. An administrator issues the token, the node registers itself, and the scheduler starts dispatching work to it.

```bash
# Admin: issue a registration token
curl -q -LSsf -X POST https://casci.casapps.us/api/v1/nodes/token \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Node: register with that token
curl -q -LSsf -X POST https://casci.casapps.us/api/v1/nodes/register \
  -H "Content-Type: application/json" \
  -d '{"token":"<registration-token>","name":"builder-01","labels":["linux","amd64"]}'
```

Nodes send periodic heartbeats to `/api/v1/nodes/{id}/heartbeat`. Drain a node with `/api/v1/nodes/{id}/drain` before removing it so running builds finish rather than being killed.

## Metrics Scraping

`/metrics` serves Prometheus-format metrics and `/metrics/json` serves the same data as JSON. Both require an admin credential, so a scrape job needs a bearer token:

```yaml
scrape_configs:
  - job_name: casci
    metrics_path: /metrics
    authorization:
      type: Bearer
      credentials: <admin-api-token>
    static_configs:
      - targets: ['casci.example.com']
```

Do not proxy `/metrics` to an unauthenticated endpoint to work around the auth requirement.

## Live Streams

`/ws` streams build log output and status changes over WebSocket, using the same authentication and per-user isolation as the REST API.

## Databases

| Driver | Accepted values | Role |
|--------|-----------------|------|
| SQLite | `sqlite` | Default; embedded, no external service |
| PostgreSQL | `postgres`, `postgresql` | External PostgreSQL |
| MySQL | `mysql`, `mariadb` | External MySQL or MariaDB |

Those are the only accepted values. Anything else fails at startup with `unsupported database type`.

The driver and connection string come from `DATABASE_DRIVER` and `DATABASE_URL`. Pooling and read replicas are `server.yml` keys: `server.database.max_conns`, `server.database.max_idle`, and `server.database.replica_urls`. See [Configuration](configuration.md).

## Artifact Storage

Build artifacts are written through a storage interface with four backends.

| Backend | Status |
|---------|--------|
| Local disk | Working. Stores, retrieves, deletes, gzip-compresses, and SHA-256 hashes artifacts under the data directory |
| S3 | **Not implemented.** Every operation returns `S3 storage not yet implemented` |
| GCS | **Not implemented.** Every operation returns `GCS storage not yet implemented` |
| Azure Blob | **Not implemented.** Every operation returns `Azure storage not yet implemented` |

The object-store backends are constructor and interface scaffolding only — no SDK is wired up behind them. Local disk is the only backend that stores anything.

## Security Scanners

Scanning does not happen inside the CASCI binary. Each scan type shells out to an external tool, which must be installed and on the `PATH` of whatever runs the scan.

| Scan type | External tool |
|-----------|---------------|
| `vulnerability` | `trivy`, `grype` |
| `secret` | `gitleaks` |
| `sast` | `semgrep` |
| `sbom` | `syft` |

If a tool is missing, that scan fails rather than silently passing. The container images used for builds do not install these tools for you.

## Reverse Proxies

Run CASCI behind nginx, Caddy, Traefik, or HAProxy and terminate TLS there, or enable TLS directly with `server.ssl.*`.

!!! warning "Forwarding headers are not read"
    CASCI does not currently parse `X-Forwarded-Proto`, `X-Forwarded-For`, `X-Forwarded-Port`, `X-Forwarded-Prefix`, or `X-Real-IP`. Sending them is harmless but changes nothing. Configure the public hostname explicitly with the `DOMAIN` environment variable or `server.fqdn`, and serve CASCI at the root of its own hostname rather than under a path prefix.

## Operator Notes

- Rotate `server.security.webhook_secret` in the provider and in CASCI together; deliveries signed with the old secret are rejected immediately
- A notification provider that starts failing does not fail the build; check the per-build delivery log at `/api/v1/builds/{id}/notifications`
- Node registration tokens are single-purpose credentials — treat them like passwords and do not commit them into node provisioning scripts
- When an integration is added to the code, it must be documented here; an undocumented integration is treated as a defect
