# API quick reference

This is a route index for the currently registered server. The default host
used in examples is the official site; read-only examples are shown to avoid
accidentally changing a live instance. The OpenAPI output is incomplete. See
the [full API reference](api.md) for access notes and limitations.

## Service endpoints

| Purpose | URL |
|---------|-----|
| HTML health | `https://casci.casapps.us/server/healthz` |
| JSON health | `https://casci.casapps.us/api/v1/server/healthz` |
| Readiness | `https://casci.casapps.us/readyz` |
| Liveness | `https://casci.casapps.us/livez` |
| Help JSON | `https://casci.casapps.us/api/v1/server/help` |
| MCP help JSON | `https://casci.casapps.us/api/v1/server/help/mcp` |
| OpenAPI JSON | `https://casci.casapps.us/api/v1/server/swagger` |
| Swagger UI | `https://casci.casapps.us/server/docs/swagger` |
| GraphQL | `https://casci.casapps.us/api/v1/server/graphql` |
| MCP | `https://casci.casapps.us/api/v1/server/mcp` |

## Authentication

Authentication endpoints are `https://casci.casapps.us/api/v1/server/auth/login` and, when public
registration is enabled, `https://casci.casapps.us/api/v1/server/auth/register`. Use
`https://casci.casapps.us/api/v1/server/auth/check` with a Bearer
token to validate a server or user credential. Authenticated REST handlers
require the credential type their route supports, sent as
`Authorization: Bearer`. Other token headers are ignored.

```bash
curl -q -LSsf https://casci.casapps.us/api/v1/users \
  -H 'Authorization: Bearer <bearer-token>'
```

## Main route groups

| Resource | Paths |
|----------|-------|
| Users | `https://casci.casapps.us/api/v1/users`, `https://casci.casapps.us/api/v1/users/tokens`, `https://casci.casapps.us/api/v1/users/tokens/{token_id}` |
| Projects | `https://casci.casapps.us/api/v1/projects`, `https://casci.casapps.us/api/v1/projects/{id}`, `https://casci.casapps.us/api/v1/projects/{id}/pipelines[/{pipeline_id}]` |
| Builds | `https://casci.casapps.us/api/v1/projects/{id}/builds`, `https://casci.casapps.us/api/v1/builds/{id}` and subroutes, including `/artifacts` |
| Webhooks | `https://casci.casapps.us/api/v1/webhooks/{provider}` and legacy `https://casci.casapps.us/webhook` |
| Nodes | `https://casci.casapps.us/api/v1/nodes`, `https://casci.casapps.us/api/v1/nodes/{id}` and subroutes |
| Credentials | `https://casci.casapps.us/api/v1/credentials/user/**`, `https://casci.casapps.us/api/v1/credentials/project/**` |
| Notifications | `https://casci.casapps.us/api/v1/notifications` and subroutes |
| Security | `https://casci.casapps.us/api/v1/security/**`, `https://casci.casapps.us/api/v1/builds/{id}/security/**` |
| Audit | `https://casci.casapps.us/api/v1/audit/**` |
| Compliance | `https://casci.casapps.us/api/v1/compliance/**` |
| Metrics | `https://casci.casapps.us/metrics`, `https://casci.casapps.us/metrics/json`, `https://casci.casapps.us/api/v1/metrics/**` |
| Preferences/help | `https://casci.casapps.us/api/v1/server/preferences/**`, `https://casci.casapps.us/api/v1/server/help/**` |

For the full method and access matrix, use [API Reference](api.md).
