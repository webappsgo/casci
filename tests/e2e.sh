#!/usr/bin/env bash
set -eo pipefail

# Browser E2E suite entry point - Docker-wrapped, manual/developer-initiated
# Runs all three mandatory tiers: SSR, no-JS browser, full browser
# Never invoked by run_tests.sh and never part of the commit gate

PROJECT_NAME="$(basename "$PWD")"
PROJECT_ORG="$(basename "$(dirname "$PWD")")"

E2E_DIR="tests/e2e"
NETWORK_NAME=""
BROWSER_NAME=""
TEST_NAME=""
WORK_DIR=""

__cleanup() {
  if [ -n "$BROWSER_NAME" ]; then
    docker rm -f "$BROWSER_NAME" >/dev/null 2>&1 || true
  fi
  if [ -n "$TEST_NAME" ]; then
    docker rm -f "$TEST_NAME" >/dev/null 2>&1 || true
  fi
  if [ -n "$NETWORK_NAME" ]; then
    docker network rm "$NETWORK_NAME" >/dev/null 2>&1 || true
  fi
}

trap __cleanup EXIT

__require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "ERROR: $1 not found - install $1 to run the E2E suite"
    exit 1
  fi
}

__random_suffix() {
  tr -dc 'a-z0-9' </dev/urandom | head -c8
}

__require docker

if [ ! -d "$E2E_DIR" ]; then
  echo "ERROR: $E2E_DIR not found - browser E2E tests are not present in this checkout"
  exit 1
fi

# Failure artifacts (screenshots, page HTML, server logs) never touch the project tree
mkdir -p "${TMPDIR:-/tmp}/${PROJECT_ORG}"
WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/${PROJECT_ORG}/${PROJECT_NAME}-XXXXXX")"
mkdir -p "$WORK_DIR/artifacts"

GO_CACHE="${GO_CACHE:-$HOME/go/pkg/mod}"
GO_BUILD="${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}"
mkdir -p "$GO_CACHE" "$GO_BUILD"

SUFFIX="$(__random_suffix)"
NETWORK_NAME="${PROJECT_NAME}-e2e-${SUFFIX}"
BROWSER_NAME="${PROJECT_NAME}-e2e-browser-${SUFFIX}"
TEST_NAME="${PROJECT_NAME}-e2e-test-${SUFFIX}"

echo "Creating isolated E2E network..."
docker network create "$NETWORK_NAME" >/dev/null

echo "Starting headless Chromium sidecar..."
docker run -d --rm \
  --name "$BROWSER_NAME" \
  --network "$NETWORK_NAME" \
  --shm-size=1g \
  chromedp/headless-shell:latest \
  --no-sandbox --disable-gpu --remote-debugging-address=0.0.0.0 --remote-debugging-port=9222 >/dev/null

echo "Waiting for the browser to accept CDP connections..."
for _ in $(seq 1 30); do
  if docker run --rm --network "$NETWORK_NAME" alpine:latest \
    sh -c "wget -q -O- http://${BROWSER_NAME}:9222/json/version >/dev/null 2>&1"; then
    break
  fi
  sleep 1
done

echo "Running E2E suite (tiers 1-3)..."
docker run --rm \
  --name "$TEST_NAME" \
  --network "$NETWORK_NAME" \
  -v "$PWD:/app" \
  -v "$GO_CACHE:/usr/local/share/go/pkg/mod" \
  -v "$GO_BUILD:/usr/local/share/go/cache" \
  -v "$WORK_DIR:/e2e" \
  -w /app \
  -e CGO_ENABLED=0 \
  -e GOFLAGS=-buildvcs=false \
  -e E2E_ARTIFACT_DIR=/e2e/artifacts \
  -e E2E_BROWSER_URL="ws://${BROWSER_NAME}:9222" \
  casjaysdev/go:latest \
  go test -tags e2e -count=1 -v "./${E2E_DIR}/..."

echo "E2E tests completed successfully"
echo "Artifacts (if any): $WORK_DIR/artifacts"
