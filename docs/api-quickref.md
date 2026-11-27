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

Sign-in endpoints are `/api/v1/server/auth/register` and
`/api/v1/server/auth/login`. Authenticated REST handlers accept user JWTs and
legacy user API tokens only through `Authorization: Bearer`. Other token
headers are ignored.

```bash
curl -q -LSsf https://casci.casapps.us/api/v1/users/me \
  -H 'Authorization: Bearer <bearer-token>'
```

## Main route groups

| Resource | Paths |
|----------|-------|
| Users | `/api/v1/users`, `/api/v1/users/me`, `/api/v1/users/{id}` |
| Projects | `/api/v1/projects`, `/api/v1/projects/{id}`, `/api/v1/projects/{id}/pipelines[/{pipeline_id}]` |
| Builds | `/api/v1/projects/{id}/builds`, `/api/v1/builds/{id}` and subroutes, including `/artifacts` |
| Webhooks | `/api/v1/webhooks/{provider}` and legacy `/webhook` |
| Nodes | `/api/v1/nodes`, `/api/v1/nodes/{id}` and subroutes |
| Credentials | `/api/v1/credentials/user/**`, `/api/v1/credentials/project/**` |
| Notifications | `/api/v1/notifications` and subroutes |
| Security | `/api/v1/security/**`, `/api/v1/builds/{id}/security/**` |
| Audit | `/api/v1/audit/**` |
| Compliance | `/api/v1/compliance/**` |
| Metrics | `/metrics`, `/metrics/json`, `/api/v1/metrics/**` |
| Preferences/help | `/api/v1/server/preferences/**`, `/api/v1/server/help/**` |

For the full method and access matrix, use [API Reference](api.md).
