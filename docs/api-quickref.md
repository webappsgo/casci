# CASCI API Quick Reference

## 🚀 Quick Access URLs

When running on `http://localhost:8080`:

### Documentation & Playgrounds
- **Swagger UI**: http://localhost:8080/swagger/
- **GraphQL Playground**: http://localhost:8080/graphql/playground
- **OpenAPI JSON**: http://localhost:8080/openapi.json
- **OpenAPI YAML**: http://localhost:8080/openapi.yaml

### API Endpoints
- **REST API Base**: http://localhost:8080/api/v1/
- **GraphQL**: http://localhost:8080/graphql
- **Metrics**: http://localhost:8080/metrics
- **Health**: http://localhost:8080/health

### Web UI
- **Home**: http://localhost:8080/
- **Admin Dashboard**: http://localhost:8080/admin

## 📖 REST API Quick Reference

### Authentication
```bash
# Register
POST /api/v1/auth/register
{"username":"user","email":"user@example.com","password":"pass"}

# Login
POST /api/v1/auth/login
{"username":"user","password":"pass"}

# Response includes token
{"token":"eyJhbG...","user":{...}}

# Use token in requests
curl -H "Authorization: Bearer TOKEN" http://localhost:8080/api/v1/...
```

### Projects
```bash
# Create project
POST /api/v1/projects
{"name":"My Project","repositoryURL":"https://github.com/user/repo","branch":"main"}

# List projects
GET /api/v1/projects

# Get project
GET /api/v1/projects/{id}

# Update project
PUT /api/v1/projects/{id}

# Delete project
DELETE /api/v1/projects/{id}
```

### Builds
```bash
# Trigger build
POST /api/v1/projects/{id}/builds
{"commitHash":"abc123","branch":"main"}

# List builds
GET /api/v1/projects/{id}/builds

# Get build
GET /api/v1/builds/{id}

# Get build log
GET /api/v1/builds/{id}/log

# Cancel build
POST /api/v1/builds/{id}/cancel

# Restart build
POST /api/v1/builds/{id}/restart
```

### Nodes
```bash
# List nodes
GET /api/v1/nodes

# Register node
POST /api/v1/nodes/register
{"name":"node1","hostname":"host1","architecture":"amd64","operatingSystem":"linux","role":"builder","capacity":10,"token":"..."}

# Get node
GET /api/v1/nodes/{id}

# Update node
PUT /api/v1/nodes/{id}

# Drain node
POST /api/v1/nodes/{id}/drain

# Delete node
DELETE /api/v1/nodes/{id}
```

### Health & Metrics
```bash
# Health check
GET /health
GET /healthz
GET /readyz
GET /livez

# Prometheus metrics
GET /metrics

# JSON metrics
GET /metrics/json

# System metrics
GET /api/v1/metrics/system

# Build metrics
GET /api/v1/metrics/builds

# Node metrics
GET /api/v1/metrics/nodes
```

## 📊 GraphQL Quick Reference

### Queries

#### Get Current User
```graphql
query {
  me {
    id
    username
    email
    isAdmin
    createdAt
  }
}
```

#### List Projects with Builds
```graphql
query {
  projects(limit: 10, offset: 0) {
    edges {
      id
      name
      description
      repositoryURL
      branch
      builds(limit: 5) {
        edges {
          id
          number
          status
          commitHash
          startedAt
          finishedAt
        }
      }
    }
    totalCount
  }
}
```

#### Get Build Details
```graphql
query {
  build(id: "123") {
    id
    number
    status
    commitHash
    commitMessage
    startedAt
    finishedAt
    duration
    project {
      id
      name
    }
  }
}
```

#### List Nodes
```graphql
query {
  nodes(limit: 20) {
    edges {
      id
      name
      hostname
      architecture
      operatingSystem
      role
      status
      capacity
      currentLoad
      lastHeartbeat
    }
    totalCount
  }
}
```

#### Health Check
```graphql
query {
  health {
    status
    timestamp
    version
  }
}
```

#### System Metrics
```graphql
query {
  metrics {
    cpu
    memory
    disk
    goroutines
    buildQueue
    activeBuilds
  }
}
```

### Mutations

#### Register User
```graphql
mutation {
  register(input: {
    username: "newuser"
    email: "newuser@example.com"
    password: "securepass123"
  }) {
    token
    user {
      id
      username
      email
      isAdmin
    }
  }
}
```

#### Login
```graphql
mutation {
  login(input: {
    username: "user"
    password: "pass"
  }) {
    token
    user {
      id
      username
      email
      isAdmin
    }
  }
}
```

#### Create Project
```graphql
mutation {
  createProject(input: {
    name: "My New Project"
    description: "A cool project"
    repositoryURL: "https://github.com/user/repo"
    branch: "main"
  }) {
    id
    name
    description
    repositoryURL
    branch
    createdAt
  }
}
```

#### Trigger Build
```graphql
mutation {
  triggerBuild(
    projectID: "123"
    input: {
      commitHash: "abc123"
      branch: "main"
    }
  ) {
    id
    number
    status
    commitHash
    createdAt
  }
}
```

#### Cancel Build
```graphql
mutation {
  cancelBuild(id: "456") {
    id
    status
  }
}
```

#### Update Node
```graphql
mutation {
  updateNode(
    id: "789"
    input: {
      capacity: 20
      status: DRAINING
    }
  ) {
    id
    capacity
    status
  }
}
```

### Subscriptions

#### Watch Build Updates
```graphql
subscription {
  buildUpdated(projectID: "123") {
    id
    status
    startedAt
    finishedAt
  }
}
```

#### Watch Build Logs
```graphql
subscription {
  buildLogLine(buildID: "456") {
    buildID
    line
    timestamp
  }
}
```

#### Watch Node Status
```graphql
subscription {
  nodeStatusChanged {
    id
    name
    status
    lastHeartbeat
  }
}
```

## 🔐 Authentication

### REST API
Use JWT Bearer tokens in the `Authorization` header:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### GraphQL
Use HTTP headers in GraphQL Playground:
```json
{
  "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## 📋 Status Codes

### REST API
- `200` - Success
- `201` - Created
- `204` - No Content
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

### GraphQL
All queries return `200` with errors in the response body:
```json
{
  "data": null,
  "errors": [
    {
      "message": "Error message",
      "path": ["field"]
    }
  ]
}
```

## 🎯 Common Use Cases

### 1. Register and Create First Project
```bash
# 1. Register
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","email":"admin@example.com","password":"admin123"}' \
  | jq -r '.token')

# 2. Create project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My App","repositoryURL":"https://github.com/user/repo","branch":"main"}'
```

### 2. Trigger Build and Watch Status
```bash
# Trigger build
BUILD_ID=$(curl -s -X POST http://localhost:8080/api/v1/projects/1/builds \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  | jq -r '.id')

# Watch status
watch -n 1 "curl -s -H 'Authorization: Bearer $TOKEN' \
  http://localhost:8080/api/v1/builds/$BUILD_ID | jq '.status'"

# Get logs
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/builds/$BUILD_ID/log
```

### 3. Check System Health
```bash
# Health check
curl http://localhost:8080/health | jq

# Metrics
curl http://localhost:8080/metrics

# System metrics
curl http://localhost:8080/api/v1/metrics/system | jq
```

## 🛠️ Development Tools

### Test with curl
```bash
# Pretty print JSON responses
curl http://localhost:8080/api/v1/projects | jq

# Save token to file
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.token' > token.txt

# Use saved token
curl -H "Authorization: Bearer $(cat token.txt)" \
  http://localhost:8080/api/v1/projects
```

### Test with GraphQL Playground
1. Open http://localhost:8080/graphql/playground
2. Click "HTTP HEADERS" at bottom
3. Add authentication:
```json
{
  "Authorization": "Bearer YOUR_TOKEN_HERE"
}
```
4. Write queries in left panel
5. Click play button to execute
6. View results in right panel
7. Explore schema with "DOCS" button

### Test with Swagger UI
1. Open http://localhost:8080/swagger/
2. Click "Authorize" button
3. Enter: `Bearer YOUR_TOKEN_HERE`
4. Click "Authorize"
5. Try any endpoint
6. View request/response

## 📚 More Information

- **Full API Docs**: See API.md
- **Development Guide**: See DEVELOPMENT.md
- **Quick Start**: See QUICKSTART.md
- **Completion Report**: See COMPLETION_REPORT.md

---

**Tip**: Both Swagger UI and GraphQL Playground provide interactive documentation with auto-completion, making it easy to explore the APIs without reading docs!
