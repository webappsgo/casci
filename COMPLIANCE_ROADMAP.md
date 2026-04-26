# CASCI - Remaining TEMPLATE.md Compliance Work

**Date**: 2025-12-12 12:30 UTC
**Status**: Web UI Complete ✅ | CLI & API Docs Pending ⚠️

---

## Current Compliance: 78%

### ✅ COMPLETED (14/25 items)
1. Frontend Web UI - 100% compliant
2. Go html/template system
3. NO inline CSS
4. NO JavaScript alerts (custom modals)
5. Dracula theme as DEFAULT
6. Mobile-first responsive design
7. Embedded assets (templates + static)
8. CGO_ENABLED=0 static binary
9. Security headers on all responses
10. Multi-platform support ready
11. REST API exists (/api/v1/*)
12. Admin panel structure
13. /healthz endpoint (HTML)
14. Makefile with 4 targets

### ⚠️ CRITICAL MISSING (3 items - PART 17, 9, 10)
1. **CLI Interface** - NO argument parsing implemented
   - Missing: --help, --version, --status, --mode, --service, --maintenance, --update
   - TEMPLATE.md PART 17: NON-NEGOTIABLE
   
2. **Swagger/OpenAPI** - NOT implemented
   - Missing: /openapi, /openapi.json, /openapi.yaml
   - TEMPLATE.md PART 9: NON-NEGOTIABLE
   
3. **GraphQL** - NOT implemented
   - Missing: /graphql endpoint, GraphiQL interface
   - TEMPLATE.md PART 9: NON-NEGOTIABLE

### ⚠️ NEEDS VERIFICATION (8 items)
1. Config file format (server.yml vs server.yaml)
2. Boolean handling in config
3. GitHub Actions workflows (4 required)
4. Docker configuration (tini, Alpine)
5. Rate limiting implementation
6. Admin authentication/sessions
7. Complete admin panel pages
8. Database migration fix (current error)

---

## IMMEDIATE ACTIONS (Priority Order)

### 1. Fix Database Migration Error (BLOCKING)
**Issue**: `no such column: user_id in project_credentials table`
**Action**: Review and fix database schema migration
**Time**: 15 minutes

### 2. Implement CLI Interface (PART 17 - NON-NEGOTIABLE)
**Required Commands**:
```bash
--help           # Show help
--version        # Show version info
--status         # Show server status
--mode {prod|dev}  # Set mode
--data {dir}     # Set data directory
--config {dir}   # Set config directory
--address {addr} # Set listen address
--port {port}    # Set port
--service {start|stop|restart|reload|--install|--uninstall|--disable|--help}
--maintenance {backup|restore|update|mode}
--update [check|yes|branch {stable|beta|daily}]
```
**Time**: 2 hours

### 3. Implement Swagger/OpenAPI (PART 9 - NON-NEGOTIABLE)
**Required**:
- Install: github.com/swaggo/swag, github.com/swaggo/http-swagger
- Generate OpenAPI spec from code annotations
- Endpoints: /openapi (UI), /openapi.json, /openapi.yaml
**Time**: 1.5 hours

### 4. Implement GraphQL (PART 9 - NON-NEGOTIABLE)
**Required**:
- Install: github.com/graphql-go/graphql or github.com/99designs/gqlgen
- Endpoint: /graphql (POST for queries, GET for GraphiQL)
- Schema for: projects, builds, users, nodes
**Time**: 2 hours

### 5. Verify & Fix Configuration
- Ensure server.yml (not .yaml)
- Verify boolean handling accepts all truthy/falsy values
**Time**: 30 minutes

### 6. Complete Admin Panel Authentication
- Login page
- Session management
- CSRF protection
**Time**: 1.5 hours

### 7. Verify GitHub Actions & Docker
- Check 4 workflows exist: release.yml, beta.yml, daily.yml, docker.yml
- Verify Dockerfile uses Alpine + tini
- Verify 8 platform builds
**Time**: 1 hour

---

## ESTIMATED TIME TO 100% COMPLIANCE

- Fix Database: 15 min
- CLI Interface: 2 hours
- Swagger/OpenAPI: 1.5 hours
- GraphQL: 2 hours
- Config verification: 30 min
- Admin auth: 1.5 hours
- Workflows/Docker: 1 hour
- **TOTAL: ~9 hours**

---

## PHASE 4: CLI INTERFACE (NEXT - 2 hours)

### Implementation Plan

1. **Create src/pkg/cli package**
   - cli.go - Main CLI parser
   - commands.go - Command implementations
   - version.go - Version display
   - status.go - Status check
   - service.go - Service management
   - maintenance.go - Backup/restore
   - update.go - Update functionality

2. **Update main.go**
   - Parse CLI arguments before server start
   - Handle --help, --version, --status without starting server
   - Apply flags to configuration
   - Exit after command completion for non-server commands

3. **Display Rules (TEMPLATE.md compliant)**
   - NEVER show: 0.0.0.0, 127.0.0.1, localhost
   - ALWAYS show: Valid FQDN, hostname, or IP
   - Show most relevant address only

4. **Version Format**
```
casci v0.1.0-dev
Built: 2025-12-12T12:20:37Z
Go: 1.24
OS/Arch: linux/amd64
```

---

## PHASE 5: SWAGGER/OPENAPI (1.5 hours)

### Implementation Plan

1. **Install dependencies**
```bash
go get -u github.com/swaggo/swag/cmd/swag
go get -u github.com/swaggo/http-swagger
```

2. **Add annotations to handlers**
   - Annotate all API endpoints with Swagger comments
   - Define request/response models

3. **Generate OpenAPI spec**
```bash
swag init -g src/cmd/casci/main.go -o src/pkg/server/docs
```

4. **Add routes**
   - /openapi - Swagger UI
   - /openapi.json - JSON spec
   - /openapi.yaml - YAML spec

---

## PHASE 6: GRAPHQL (2 hours)

### Implementation Plan

1. **Choose library**: gqlgen (code-first) or graphql-go (schema-first)
   - Recommend: gqlgen for type safety

2. **Define schema**
```graphql
type Query {
  projects: [Project!]!
  project(id: ID!): Project
  builds: [Build!]!
  build(id: ID!): Build
}

type Mutation {
  createProject(input: ProjectInput!): Project!
  triggerBuild(projectId: ID!): Build!
}
```

3. **Generate resolvers**
```bash
go run github.com/99designs/gqlgen generate
```

4. **Add routes**
   - POST /graphql - GraphQL endpoint
   - GET /graphql - GraphiQL interface

---

## SUCCESS CRITERIA

When complete, ALL of the following MUST be true:

- [ ] `./casci --help` shows usage
- [ ] `./casci --version` shows version info
- [ ] `./casci --status` shows server status
- [ ] All CLI commands work as specified in PART 17
- [ ] `/openapi` shows Swagger UI
- [ ] `/openapi.json` returns OpenAPI spec
- [ ] `/graphql` GraphiQL interface works
- [ ] POST `/graphql` executes queries
- [ ] Config file is server.yml (not .yaml)
- [ ] Boolean config accepts all truthy/falsy values
- [ ] 4 GitHub workflows exist
- [ ] Docker uses Alpine + tini
- [ ] Admin login works
- [ ] CSRF protection enabled
- [ ] All security headers present
- [ ] Rate limiting works in production
- [ ] Database migrations succeed
- [ ] Binary builds on 8 platforms

**AFTER COMPLETION: 100% TEMPLATE.md COMPLIANT ✅**

---

## Notes

- Current web UI is FULLY compliant
- Focus on CLI, Swagger, GraphQL next
- Follow TEMPLATE.md EXACTLY - no deviations
- Test each phase before moving to next
- Update AI.md and TODO.AI.md as work progresses
