# Agent management

CASCI includes a companion agent for collecting machine health and metrics.
Its `--version` output reports the embedded release version, commit ID, and
UTC build timestamp, matching the server and client binaries.
Server administrators manage server agents at
`https://casci.casapps.us/server/{admin_path}/agents`; regular users manage only their own agents at
`https://casci.casapps.us/users/agents`. The default administrator path is `administration`.

## Create an agent

Sign in and create an agent from the appropriate Agents page. Choose an
expiration of one hour, 24 hours, seven days, or never. The page displays the
generated install command once. Treat the token in that command as a secret:
CASCI stores only its SHA-256 hash. Use the rotate action if the command or
token is lost or exposed.

The machine token is a regular `adm_` or `usr_` token with `scope: agent` and
is limited to its own registration and reports. It cannot use account
management routes. User management requires the owner's bearer JWT; the agent
token must not be used for management calls.

## Agent API

The administrator base path is
`https://casci.casapps.us/api/{api_version}/server/{admin_path}/agents`; the user base path is
`https://casci.casapps.us/api/{api_version}/users/agents`. Management routes require an administrator
or owner account token respectively.

| Method | Suffix | Purpose |
|--------|--------|---------|
| GET | empty | List agents for the authenticated owner |
| POST | empty | Create an agent record and scoped token |
| GET | `/{id}` | Read one owned agent |
| PATCH | `/{id}` | Update its name, tags, or labels |
| DELETE | `/{id}` | Revoke its token and remove its metrics and record |
| POST | `/{id}/token` | Rotate its token |
| GET | `https://casci.casapps.us/api/{api_version}/server/{admin_path}/agents/{id}/metrics` | Read up to 100 recent metric samples |
| POST | `/register` | Register the machine using its agent token |
| POST | `/heartbeat` or `/report` | Submit agent health and metrics using its token |

Machine requests send the credential only in the authorization header:

```http
Authorization: Bearer adm_<agent-token>
Content-Type: application/json
```

Registration payload:

```json
{
  "hostname": "build-01",
  "os": "linux",
  "arch": "amd64",
  "version": "1.0.0",
  "tags": ["ci"],
  "labels": {"zone": "east"}
}
```

The health report includes the registered `agent_id` and a JSON `metrics`
object. The server records a metric sample and updates the last-seen time.
Agents with no report for 90 seconds appear offline. An agent token cannot
read metrics history, which is available only to its owner.

Downloads from `https://casci.casapps.us/agent/binaries/{name}` require any valid agent-scoped bearer
token. The server serves only published agent binaries with a matching
SHA-256 entry in the release checksum file. The `https://casci.casapps.us/api/autodiscover` response
advertises available platform versions and checksums.

## Current limits

Agent ownership currently supports administrator and individual user scopes.
Organization-scoped agent APIs, real-time connection notifications, and
cross-platform full-fidelity host metrics are not implemented. Agent
configuration, service behavior, update behavior, and platform coverage are
documented in the project specification; consult the implementation before
deploying agents in production.
