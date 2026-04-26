.PHONY: build release docker test help

# Version from release.txt
VERSION := $(shell cat release.txt 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags - ALWAYS CGO_ENABLED=0 for static binary
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Binary name
BINARY := casci

# Temp directory (BASE.md compliant: /tmp/{tmpdir}/{projectname}/)
TMPDIR := /tmp/casapps-build/casci

help: ## Show this help
@echo 'CASCI Makefile - BASE.md Compliant'
@echo ''
@echo 'Usage: make [target]'
@echo ''
@echo 'Required targets (NON-NEGOTIABLE):'
@awk 'BEGIN {FS = ":.*?## "} /^[a-z]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all platforms to ./binaries
@echo "Building CASCI $(VERSION) for all platforms..."
@mkdir -p binaries $(TMPDIR)
@echo "  Building linux-amd64..."
@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TMPDIR)/$(BINARY)-linux-amd64 ./src/cmd/casci
@mv $(TMPDIR)/$(BINARY)-linux-amd64 binaries/$(BINARY)-linux-amd64
@echo "  Building linux-arm64..."
@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TMPDIR)/$(BINARY)-linux-arm64 ./src/cmd/casci
@mv $(TMPDIR)/$(BINARY)-linux-arm64 binaries/$(BINARY)-linux-arm64
@echo "  Building darwin-amd64..."
@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TMPDIR)/$(BINARY)-darwin-amd64 ./src/cmd/casci
@mv $(TMPDIR)/$(BINARY)-darwin-amd64 binaries/$(BINARY)-darwin-amd64
@echo "  Building darwin-arm64..."
@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TMPDIR)/$(BINARY)-darwin-arm64 ./src/cmd/casci
@mv $(TMPDIR)/$(BINARY)-darwin-arm64 binaries/$(BINARY)-darwin-arm64
@echo "  Building windows-amd64..."
@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TMPDIR)/$(BINARY)-windows-amd64.exe ./src/cmd/casci
@mv $(TMPDIR)/$(BINARY)-windows-amd64.exe binaries/$(BINARY)-windows-amd64.exe
@echo "  Stripping -musl suffix from any binaries..."
@cd binaries && for f in *-musl*; do [ -f "$$f" ] && mv "$$f" "$${f//-musl/}" || true; done
@rm -rf $(TMPDIR)
@echo "✓ All builds complete in ./binaries/"
@ls -lh binaries/

release: ## GitHub release - production to ./releases
@echo "Creating production release $(VERSION)..."
@mkdir -p releases $(TMPDIR)
@echo "  Building linux-amd64..."
@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TMPDIR)/$(BINARY)-linux-amd64 ./src/cmd/casci
@mv $(TMPDIR)/$(BINARY)-linux-amd64 releases/$(BINARY)-linux-amd64
@echo "  Building linux-arm64..."
@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TMPDIR)/$(BINARY)-linux-arm64 ./src/cmd/casci
@mv $(TMPDIR)/$(BINARY)-linux-arm64 releases/$(BINARY)-linux-arm64
@echo "  Building darwin-amd64..."
@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TMPDIR)/$(BINARY)-darwin-amd64 ./src/cmd/casci
@mv $(TMPDIR)/$(BINARY)-darwin-amd64 releases/$(BINARY)-darwin-amd64
@echo "  Building darwin-arm64..."
@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TMPDIR)/$(BINARY)-darwin-arm64 ./src/cmd/casci
@mv $(TMPDIR)/$(BINARY)-darwin-arm64 releases/$(BINARY)-darwin-arm64
@echo "  Building windows-amd64..."
@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TMPDIR)/$(BINARY)-windows-amd64.exe ./src/cmd/casci
@mv $(TMPDIR)/$(BINARY)-windows-amd64.exe releases/$(BINARY)-windows-amd64.exe
@echo "  Stripping -musl suffix from any binaries..."
@cd releases && for f in *-musl*; do [ -f "$$f" ] && mv "$$f" "$${f//-musl/}" || true; done
@echo "  Compressing releases..."
@cd releases && for f in casci-*; do \
if [ -f "$$f" ]; then \
tar czf "$$f.tar.gz" "$$f" && rm "$$f"; \
fi; \
done
@rm -rf $(TMPDIR)
@echo "✓ Release $(VERSION) created in ./releases/"
@ls -lh releases/

docker: ## Docker release for ARM64/AMD64
@echo "Building Docker images for ARM64 and AMD64..."
@docker buildx version >/dev/null 2>&1 || (echo "ERROR: docker buildx not available" && exit 1)
@echo "  Building and pushing multi-arch image..."
@docker buildx build \
--platform linux/amd64,linux/arm64 \
--build-arg VERSION=$(VERSION) \
--build-arg BUILD_TIME=$(BUILD_TIME) \
--build-arg GIT_COMMIT=$(GIT_COMMIT) \
-t ghcr.io/casapps/$(BINARY):$(VERSION) \
-t ghcr.io/casapps/$(BINARY):latest \
--push \
-f Dockerfile .
@echo "✓ Docker images pushed to ghcr.io/casapps/$(BINARY)"

test: ## Run all tests
@echo "Running all tests..."
@go test -v -race -cover ./...
@echo "✓ All tests passed"

.DEFAULT_GOAL := help
