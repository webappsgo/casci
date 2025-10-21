# CASCI Quick Start Guide

Get CASCI running in under 5 minutes!

## Prerequisites

- Docker and Docker Compose (recommended)
- OR Go 1.21+ and Docker CLI (for local development)

## Quick Start with Docker

### 1. Start CASCI with Docker Compose

```bash
# Clone the repository
git clone https://github.com/casapps/casci.git
cd casci

# Start CASCI (includes PostgreSQL)
make docker-run

# Or manually:
docker-compose up -d
```

CASCI will be available at: **http://localhost:8080**

### 2. Verify it's running

```bash
curl http://localhost:8080/health
# Should return: {"status": "healthy"}
```

### 3. Register your first user (becomes admin)

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "yourpassword"
  }'
```

Save the returned `token` and `api_token` for future requests.

### 4. Create your first project

```bash
# Set your token from previous step
TOKEN="your_jwt_token_here"

# Create a project
curl -X POST http://localhost:8080/api/v1/projects/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hello-world",
    "repository_url": "https://github.com/octocat/Hello-World",
    "branch": "master"
  }'
```

### 5. Trigger your first build

```bash
# Get the project ID from the previous response
PROJECT_ID=1

# Trigger a build
curl -X POST http://localhost:8080/api/v1/projects/$PROJECT_ID/builds \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "branch": "master",
    "trigger": "manual"
  }'
```

### 6. Check build status

```bash
# Get the build ID from the previous response
BUILD_ID=1

# Check status
curl http://localhost:8080/api/v1/builds/$BUILD_ID \
  -H "Authorization: Bearer $TOKEN"
```

### 7. View logs

```bash
docker-compose logs -f casci
```

---

## Quick Start without Docker

### 1. Build CASCI

```bash
# Install dependencies
make deps

# Build the binary
make build

# Or use the build script
./build.sh
```

### 2. Run CASCI

```bash
# Run directly
./casci

# Or with make
make run
```

CASCI will start on a random port (64000-64999). Check the output for the actual port.

### 3. Follow steps 2-6 above

Use the port shown in the startup message.

---

## Run API Tests

Automated test suite that covers the complete workflow:

```bash
# Make sure CASCI is running first
make docker-run

# Wait a few seconds for startup, then run tests
make test-api

# Or manually
./test-api.sh
```

The test script will:
- ✓ Verify server health
- ✓ Register/login a user
- ✓ Create a project
- ✓ Trigger a build
- ✓ Monitor build status
- ✓ Check build statistics
- ✓ Test Jenkins API compatibility
- ✓ Clean up resources

---

## Common Commands

### Docker Commands

```bash
# Start CASCI
make docker-run

# View logs
make docker-logs

# Stop CASCI
make docker-stop

# Clean everything (including volumes)
make docker-clean

# Rebuild and restart
make dev

# Open shell in container
make docker-shell
```

### Local Development

```bash
# Fast build and run (for development)
make run-fast

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint

# Build for all platforms
make build-all
```

### Database Management

```bash
# Reset database
make db-reset

# Open SQLite shell
make db-shell
```

---

## Directory Structure

After starting CASCI, the following directories are created:

```
/var/lib/casci/
├── casci.db              # SQLite database
├── workspaces/           # Build workspaces
│   └── project-{id}/
│       └── build-{id}/
├── cache/                # Build cache
└── artifacts/            # Build artifacts
    └── {user}/
        └── {project}/
            └── {build}/

/var/log/casci/
├── casci.log            # Main application log
└── builds/              # Build logs
    └── build-{id}.log
```

---

## Configuration

CASCI works with zero configuration, but you can customize via environment variables:

### Environment Variables

```bash
# Server configuration
CASCI_HOST=0.0.0.0        # Listen address (default: 0.0.0.0)
CASCI_PORT=8080           # Port (default: random 64000-64999)

# Database configuration
CASCI_DB_TYPE=sqlite      # sqlite, postgres, mysql
CASCI_DB_HOST=localhost   # For postgres/mysql
CASCI_DB_PORT=5432        # Database port
CASCI_DB_NAME=casci       # Database name
CASCI_DB_USER=casci       # Database user
CASCI_DB_PASSWORD=casci   # Database password

# Paths
CASCI_DATA_DIR=/var/lib/casci
CASCI_LOG_DIR=/var/log/casci
```

### Using PostgreSQL

```yaml
# docker-compose.yml already includes PostgreSQL
# To use it, set:
environment:
  - CASCI_DB_TYPE=postgres
  - CASCI_DB_HOST=postgres
  - CASCI_DB_PORT=5432
  - CASCI_DB_NAME=casci
  - CASCI_DB_USER=casci
  - CASCI_DB_PASSWORD=casci
```

---

## Troubleshooting

### Docker socket permission denied

```bash
# Add your user to docker group
sudo usermod -aG docker $USER

# Log out and back in, or run:
newgrp docker
```

### Port already in use

```bash
# Change the port in docker-compose.yml
ports:
  - "8081:8080"  # Use 8081 instead
```

### Build fails with "Docker not available"

Make sure Docker is running and accessible:

```bash
# Check Docker is running
docker ps

# Check Docker socket is mounted (in container)
ls -l /var/run/docker.sock
```

### Cannot clone private repositories

Add SSH keys or HTTPS credentials to your projects:

```bash
# For private repos, use HTTPS with token
https://username:token@github.com/user/private-repo.git

# Or configure SSH keys (future feature)
```

### Database connection errors

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Reset database
make db-reset
```

---

## Next Steps

1. **Read the API documentation**: See [API.md](API.md) for complete API reference
2. **Configure projects**: Set up your repositories and branches
3. **Set up webhooks**: Configure GitHub/GitLab webhooks (coming soon)
4. **Add build nodes**: Scale by adding more build nodes
5. **Customize pipelines**: Create custom pipelines for your projects

---

## Getting Help

- **Documentation**: See [DEVELOPMENT.md](DEVELOPMENT.md) for detailed information
- **API Reference**: See [API.md](API.md) for API documentation
- **Issues**: Report bugs at https://github.com/casapps/casci/issues
- **Specification**: See [CLAUDE.md](CLAUDE.md) for complete system specification

---

## What's Working

✅ User authentication (JWT + API tokens)
✅ Project management
✅ Build triggering and queuing
✅ Docker container execution
✅ Git repository cloning
✅ Build status tracking
✅ Build logs
✅ Jenkins API compatibility
✅ Multi-database support (SQLite, PostgreSQL, MySQL)

## Coming Soon

🚧 Pipeline parsing (Jenkinsfile, GitHub Actions, etc.)
🚧 Webhook support (GitHub, GitLab, Bitbucket)
🚧 Advanced pipeline features (matrix builds, services, etc.)
🚧 Web UI
🚧 Security scanning (Trivy, Semgrep, etc.)
🚧 Artifact management
🚧 Notification system

---

## Example Workflow

```bash
# 1. Start CASCI
make docker-run

# 2. Register and save token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"dev","email":"dev@example.com","password":"dev123"}' \
  | jq -r '.token')

# 3. Create a project
PROJECT_ID=$(curl -s -X POST http://localhost:8080/api/v1/projects/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-app","repository_url":"https://github.com/user/my-app","branch":"main"}' \
  | jq -r '.id')

# 4. Trigger a build
BUILD_ID=$(curl -s -X POST http://localhost:8080/api/v1/projects/$PROJECT_ID/builds \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"branch":"main","trigger":"manual"}' \
  | jq -r '.id')

# 5. Monitor build
watch -n 2 "curl -s http://localhost:8080/api/v1/builds/$BUILD_ID \
  -H 'Authorization: Bearer $TOKEN' | jq '.status'"

# 6. Get logs
curl http://localhost:8080/api/v1/builds/$BUILD_ID/log \
  -H "Authorization: Bearer $TOKEN"
```

---

**That's it!** You now have a working CI/CD server. 🚀
