#!/bin/bash
# Quick build script for CASCI development
set -e

VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

echo "🔨 Building CASCI $VERSION..."
echo "   Time: $BUILD_TIME"
echo "   Commit: $GIT_COMMIT"
echo ""

# Build with CGO for SQLite support
CGO_ENABLED=1 go build -tags dev -ldflags="$LDFLAGS" -o casci ./cmd/casci

echo "✓ Build complete: ./casci"
echo ""
echo "Run with: ./casci"
echo "Or use: make run"