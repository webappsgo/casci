.PHONY: help build run test clean docker-build docker-run dev api-test

# Version information
VERSION ?= dev
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"
DEV_FLAGS := -tags dev
PROD_FLAGS := -tags prod

# Binary name
BINARY := casci
DOCKER_IMAGE := casci:dev

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build CASCI binary for development
	@echo "Building CASCI $(VERSION)..."
	@mkdir -p bin
	CGO_ENABLED=1 go build $(DEV_FLAGS) -o $(BINARY) ./cmd/casci
	@echo "✓ Build complete: ./$(BINARY)"

build-prod: ## Build CASCI binary for production
	@echo "Building CASCI $(VERSION) for production..."
	@mkdir -p bin
	CGO_ENABLED=0 go build $(PROD_FLAGS) $(LDFLAGS) -o $(BINARY) ./cmd/casci
	@echo "✓ Build complete: ./$(BINARY)"

build-fast: ## Fast build without optimizations
	@echo "Building CASCI (fast mode)..."
	go build -o $(BINARY) ./cmd/casci
	@echo "✓ Build complete: ./$(BINARY)"

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(PROD_FLAGS) $(LDFLAGS) -o dist/casci-linux-amd64 ./cmd/casci
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(PROD_FLAGS) $(LDFLAGS) -o dist/casci-linux-arm64 ./cmd/casci
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(PROD_FLAGS) $(LDFLAGS) -o dist/casci-darwin-amd64 ./cmd/casci
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(PROD_FLAGS) $(LDFLAGS) -o dist/casci-darwin-arm64 ./cmd/casci
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(PROD_FLAGS) $(LDFLAGS) -o dist/casci-windows-amd64.exe ./cmd/casci
	@echo "All builds complete in dist/"

run: build ## Build and run CASCI
	@echo "Starting CASCI..."
	./$(BINARY)

run-fast: build-fast ## Fast build and run
	@echo "Starting CASCI..."
	./$(BINARY)

test: ## Run unit tests
	@echo "Running unit tests..."
	go test -v -race -cover ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -tags integration ./...

test-api: ## Run API tests (requires running server)
	@echo "Running API tests..."
	@bash test-api.sh

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

fmt: ## Format code
	go fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY)
	rm -rf dist/
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

clean-all: clean ## Clean everything including caches
	@echo "Deep cleaning..."
	rm -rf /var/lib/casci/*
	rm -rf /var/log/casci/*
	go clean -cache -testcache -modcache
	@echo "✓ Deep clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -f Dockerfile.dev -t $(DOCKER_IMAGE) .
	@echo "✓ Docker image built: $(DOCKER_IMAGE)"

docker-run: docker-build ## Run CASCI in Docker
	@echo "Starting Docker containers..."
	docker-compose up -d
	@echo "✓ CASCI running at http://localhost:8080"

docker-stop: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down
	@echo "✓ Containers stopped"

docker-logs: ## Show Docker logs
	docker-compose logs -f casci

docker-ps: ## Show running containers
	docker-compose ps

docker-clean: ## Clean Docker resources
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	docker rmi $(DOCKER_IMAGE) 2>/dev/null || true
	@echo "✓ Docker resources cleaned"

docker-shell: ## Open shell in CASCI container
	docker exec -it casci-dev sh

dev: ## Start development environment with Docker Compose
	docker-compose up --build

dev-postgres: ## Start development environment with PostgreSQL
	docker-compose --profile postgres up --build

dev-mysql: ## Start development environment with MySQL
	docker-compose --profile mysql up --build

deps: ## Download dependencies
	go mod download
	go mod tidy

install: build-prod ## Install CASCI system-wide
	@echo "Installing CASCI to /usr/local/bin..."
	sudo cp casci /usr/local/bin/casci
	sudo chmod +x /usr/local/bin/casci
	@echo "CASCI installed successfully"

uninstall: ## Uninstall CASCI from system
	@echo "Removing CASCI..."
	sudo rm -f /usr/local/bin/casci
	@echo "CASCI uninstalled"

# Database management
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	go run ./cmd/casci migrate

db-reset: ## Reset database
	@echo "Resetting database..."
	rm -f /var/lib/casci/casci.db
	@echo "✓ Database reset"

db-shell: ## Open SQLite shell
	sqlite3 /var/lib/casci/casci.db

# Development helpers
init: ## Initialize development environment
	@echo "Initializing CASCI development environment..."
	@mkdir -p /var/lib/casci/workspaces
	@mkdir -p /var/lib/casci/cache
	@mkdir -p /var/lib/casci/artifacts
	@mkdir -p /var/log/casci/builds
	@mkdir -p /etc/casci/security
	@echo "✓ Directories created"
	@go mod download
	@echo "✓ Dependencies downloaded"
	@echo "✓ Development environment ready"

check: fmt vet lint ## Run all checks

watch: ## Watch for changes and rebuild
	@which entr > /dev/null || (echo "entr not installed. Install with: apt-get install entr" && exit 1)
	@echo "Watching for changes..."
	find . -name '*.go' | entr -r make run-fast

.DEFAULT_GOAL := help