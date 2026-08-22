#!/usr/bin/env bash
set -eo pipefail

TESTS_DIR="$(dirname "$0")"

__usage() {
  echo "Usage: $(basename "$0") [auto|incus|docker|api]"
  echo
  echo "  auto    Detect the available container runtime (default)"
  echo "  incus   Force the Incus integration + systemd suite"
  echo "  docker  Force the Docker Alpine integration suite"
  echo "  api     Run the HTTP API suite (tests/api.sh) against a running server"
  echo
  echo "Browser E2E lives in tests/e2e.sh and is never run from here."
}

__run_api() {
  if [ ! -x "$TESTS_DIR/api.sh" ]; then
    echo "ERROR: $TESTS_DIR/api.sh not found or not executable"
    exit 1
  fi
  exec "$TESTS_DIR/api.sh"
}

__run_incus() {
  if ! command -v incus &>/dev/null; then
    echo "ERROR: incus not found"
    exit 1
  fi
  exec "$TESTS_DIR/incus.sh"
}

__run_docker() {
  if ! command -v docker &>/dev/null; then
    echo "ERROR: docker not found"
    exit 1
  fi
  exec "$TESTS_DIR/docker.sh"
}

case "${1:-auto}" in
-h | --help | help)
  __usage
  exit 0
  ;;
api)
  __run_api
  ;;
incus)
  __run_incus
  ;;
docker)
  __run_docker
  ;;
auto) ;;
*)
  __usage
  exit 1
  ;;
esac

# Detect available container runtime and run appropriate test
if command -v incus &>/dev/null; then
  echo "Incus detected - running full systemd tests..."
  exec "$TESTS_DIR/incus.sh"
elif command -v docker &>/dev/null; then
  echo "Docker detected - running container tests..."
  exec "$TESTS_DIR/docker.sh"
else
  echo "ERROR: Neither incus nor docker found"
  echo "Please install one of the following:"
  echo "  - Incus (preferred): https://linuxcontainers.org/incus/"
  echo "  - Docker (fallback): https://docker.com/"
  exit 1
fi
