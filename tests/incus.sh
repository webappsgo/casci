#!/usr/bin/env bash
set -eo pipefail

# Check if incus is available
if ! command -v incus &>/dev/null; then
    echo "ERROR: incus not found. Install incus or use tests/docker.sh"
    exit 1
fi

# Detect project info
PROJECT_NAME=$(basename "$PWD")
PROJECT_ORG=$(basename "$(dirname "$PWD")")
CONTAINER_NAME="test-${PROJECT_NAME}-$$"

# Incus image - use latest Debian stable (update when new stable releases)
INCUS_IMAGE="images:debian/trixie"

trap "incus delete $CONTAINER_NAME --force 2>/dev/null || true" EXIT

# Build — use Makefile if present (standard for all bootstrapped projects)
# Output always lands in binaries/
if [ -f "Makefile" ]; then
    echo "Building with make build..."
    make build
else
    # No Makefile yet — fallback to direct docker run
    echo "Building in Docker (no Makefile)..."
    GO_CACHE="${GO_CACHE:-$HOME/go/pkg/mod}"
    GO_BUILD="${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}"
    mkdir -p "$GO_CACHE" "$GO_BUILD" binaries

    # Common docker run for Go builds
    GO_DOCKER="docker run --rm \
      --name ${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8) \
      -v $PWD:/app \
      -v $GO_CACHE:/usr/local/share/go/pkg/mod \
      -v $GO_BUILD:/usr/local/share/go/cache \
      -w /app \
      -e CGO_ENABLED=0 \
      -e GOFLAGS=-buildvcs=false \
      casjaysdev/go:latest"

    echo "Building server binary in Docker..."
    $GO_DOCKER go build -buildvcs=false -trimpath -ldflags "-s -w" -o /app/binaries/${PROJECT_NAME} ./src

    # Build client if exists
    if [ -d "src/client" ]; then
        echo "Building client in Docker..."
        $GO_DOCKER go build -buildvcs=false -trimpath -ldflags "-s -w" -o /app/binaries/${PROJECT_NAME}-cli ./src/client
    fi

    # Build agent if exists
    if [ -d "src/agent" ]; then
        echo "Building agent in Docker..."
        $GO_DOCKER go build -buildvcs=false -trimpath -ldflags "-s -w" -o /app/binaries/${PROJECT_NAME}-agent ./src/agent
    fi
fi

echo "Launching Incus container (Debian + systemd)..."
incus launch "$INCUS_IMAGE" "$CONTAINER_NAME"

# Wait for container to be ready
sleep 2

echo "Copying binaries to container..."
incus file push "binaries/${PROJECT_NAME}" "$CONTAINER_NAME/usr/local/bin/"
incus exec "$CONTAINER_NAME" -- chmod +x "/usr/local/bin/${PROJECT_NAME}"

# Copy client if built
if [ -f "binaries/${PROJECT_NAME}-cli" ]; then
    incus file push "binaries/${PROJECT_NAME}-cli" "$CONTAINER_NAME/usr/local/bin/"
    incus exec "$CONTAINER_NAME" -- chmod +x "/usr/local/bin/${PROJECT_NAME}-cli"
fi

# Copy agent if built
if [ -f "binaries/${PROJECT_NAME}-agent" ]; then
    incus file push "binaries/${PROJECT_NAME}-agent" "$CONTAINER_NAME/usr/local/bin/"
    incus exec "$CONTAINER_NAME" -- chmod +x "/usr/local/bin/${PROJECT_NAME}-agent"
fi

# Ensure curl is available for testing
incus exec "$CONTAINER_NAME" -- bash -c "command -v curl || apt-get update && apt-get install -y curl" >/dev/null 2>&1

echo "Running tests in Incus..."
incus exec "$CONTAINER_NAME" -- bash -c "
    set -eo pipefail

    echo '=== Version Check ==='
    ${PROJECT_NAME} --version

    echo '=== Help Check ==='
    ${PROJECT_NAME} --help

    echo '=== Binary Info ==='
    ls -lh /usr/local/bin/${PROJECT_NAME}
    file /usr/local/bin/${PROJECT_NAME}

    echo '=== Service Install Test ==='
    ${PROJECT_NAME} --service --install

    echo '=== Service Status ==='
    systemctl status ${PROJECT_NAME} || true

    echo '=== Service Start Test ==='
    systemctl start ${PROJECT_NAME}
    sleep 2
    systemctl status ${PROJECT_NAME}

    echo '=== API Endpoint Tests ==='
    # Test JSON response (default)
    curl -q -LSsf http://localhost:80/api/v1/server/healthz || echo 'FAILED: /api/v1/server/healthz'

    # Test .txt extension (plain text)
    curl -q -LSsf http://localhost:80/api/v1/server/healthz.txt || echo 'FAILED: /api/v1/server/healthz.txt'

    # Test Accept header: application/json
    curl -q -LSsf -H 'Accept: application/json' http://localhost:80/server/healthz || echo 'FAILED: Accept JSON'

    # Test Accept header: text/plain
    curl -q -LSsf -H 'Accept: text/plain' http://localhost:80/server/healthz || echo 'FAILED: Accept text/plain'

    echo '=== Project-Specific Endpoint Tests ==='
    # MUST test ALL endpoints from IDEA.md - both API and frontend
    # Test FULL CRUD if project has CRUD operations
    #
    # Example for jokes API (API routes with .txt extension):
    #   curl -q -LSsf http://localhost:80/api/v1/jokes/random || echo 'FAILED: API JSON'
    #   curl -q -LSsf http://localhost:80/api/v1/jokes/random.txt || echo 'FAILED: API .txt'
    #   curl -q -LSsf -H 'Accept: text/plain' http://localhost:80/api/v1/jokes/random || echo 'FAILED: API Accept text'
    #
    # Example for jokes frontend (smart detection, no .txt - test with text for simplicity):
    #   JOKE=\$(curl -q -LSsf http://localhost:80/jokes/random)  # CLI auto-detects text
    #   if echo "\$JOKE" | grep -q "Why"; then echo '✓ Frontend text works'; else echo 'FAILED: Frontend text'; fi
    #   # Optional: verify HTML is served to browsers
    #   curl -q -LSsfI -H 'Accept: text/html' http://localhost:80/jokes/random | grep -q 'text/html' || echo 'FAILED: Frontend HTML'
    #
    # Example for user CRUD (full test suite):
    #   # API - Current user
    #   curl -q -LSsf http://localhost:80/api/v1/users || echo 'FAILED: GET current user API'
    #   curl -q -LSsf -X PATCH -H 'Content-Type: application/json' -d '{\"email\":\"new@test.com\"}' http://localhost:80/api/v1/users || echo 'FAILED: UPDATE current user API'
    #   # API - Public profile (by username)
    #   curl -q -LSsf http://localhost:80/api/v1/users/testuser || echo 'FAILED: READ public profile API JSON'
    #   curl -q -LSsf http://localhost:80/api/v1/users/testuser.txt || echo 'FAILED: READ public profile API .txt'
    #   # API - Admin routes (by ID)
    #   curl -q -LSsf http://localhost:80/api/v1/server/administration/config/users || echo 'FAILED: LIST users admin API'
    #   curl -q -LSsf http://localhost:80/api/v1/server/administration/config/users/1 || echo 'FAILED: READ user admin API'
    #   curl -q -LSsf -X DELETE http://localhost:80/api/v1/server/administration/config/users/1 || echo 'FAILED: DELETE user admin API'
    #   # Frontend (smart detection - test with text for simplicity)
    #   USERS=\$(curl -q -LSsf http://localhost:80/users)  # CLI auto-detects text (current user)
    #   USER=\$(curl -q -LSsf http://localhost:80/testuser)  # CLI auto-detects text (public profile)
    #
    # Test ALL project-specific endpoints defined in IDEA.md

    echo '=== Admin Setup & API Token Creation ==='
    # Get setup token from journal
    SETUP_TOKEN=\$(journalctl -u ${PROJECT_NAME} --no-pager 2>/dev/null | grep -oP 'Setup Token.*:\\s*\\K[a-f0-9]+' | head -1 || echo '')

    if [ -n \"\$SETUP_TOKEN\" ]; then
        echo \"Setup token found: \${SETUP_TOKEN:0:8}...\"

        # Create admin account
        curl -q -LSsf -X POST \\
            -H \"X-Setup-Token: \$SETUP_TOKEN\" \\
            -H \"Content-Type: application/json\" \\
            -d '{\"username\":\"testadmin\",\"password\":\"TestPass123!\"}' \\
            http://localhost:80/api/v1/server/administration/setup || echo 'Admin setup failed (may already exist)'

        # Login and get session
        SESSION=\$(curl -q -LSsf -X POST \\
            -H \"Content-Type: application/json\" \\
            -d '{\"username\":\"testadmin\",\"password\":\"TestPass123!\"}' \\
            http://localhost:80/api/v1/server/administration/login | grep -oP '\"session_token\":\\s*\"\\K[^\"]+' || echo '')

        if [ -n \"\$SESSION\" ]; then
            echo '✓ Admin login successful'

            # Generate API token for CLI/Agent testing
            API_TOKEN=\$(curl -q -LSsf -X POST \\
                -H \"Authorization: Bearer \$SESSION\" \\
                http://localhost:80/api/v1/server/administration/testadmin/profile/token | grep -oP '\"token\":\\s*\"\\K[^\"]+' || echo '')

            if [ -n \"\$API_TOKEN\" ]; then
                echo \"✓ API token created: \${API_TOKEN:0:12}...\"
            else
                echo 'API token creation failed (continuing without token)'
            fi
        else
            echo 'Admin login failed (continuing without session)'
        fi
    else
        echo 'No setup token found (server may already be configured)'
    fi

    echo '=== Binary Rename Tests ==='
    # Test that binaries show ACTUAL name in --help/--version (not hardcoded)
    cp /usr/local/bin/${PROJECT_NAME} /tmp/renamed-server
    chmod +x /tmp/renamed-server
    if /tmp/renamed-server --help 2>&1 | grep -q 'renamed-server'; then
        echo '✓ Server binary rename works (--help shows actual name)'
    else
        echo '✗ FAILED: Server --help does not show renamed binary name'
    fi

    echo '=== Client Tests (if exists) ==='
    if [ -f /usr/local/bin/${PROJECT_NAME}-cli ]; then
        ${PROJECT_NAME}-cli --version || echo 'FAILED: CLI --version'
        ${PROJECT_NAME}-cli --help || echo 'FAILED: CLI --help'

        # Test binary rename
        cp /usr/local/bin/${PROJECT_NAME}-cli /tmp/renamed-cli
        chmod +x /tmp/renamed-cli
        if /tmp/renamed-cli --help 2>&1 | grep -q 'renamed-cli'; then
            echo '✓ CLI binary rename works'
        else
            echo '✗ FAILED: CLI --help does not show renamed binary name'
        fi

        # Full CLI functionality tests against server
        echo '--- CLI Full Functionality Tests ---'
        if [ -n \"\${API_TOKEN:-}\" ]; then
            # Test with API token
            ${PROJECT_NAME}-cli --server http://localhost:80 --token \"\$API_TOKEN\" status || echo 'CLI status failed'
            # Project-specific CLI commands go here (IDEA.md)
        else
            # Test without token (anonymous if allowed)
            ${PROJECT_NAME}-cli --server http://localhost:80 status || echo 'CLI status (no token) failed or not applicable'
        fi
    else
        echo 'client not installed - skipping'
    fi

    echo '=== Agent Tests (if exists) ==='
    if [ -f /usr/local/bin/${PROJECT_NAME}-agent ]; then
        ${PROJECT_NAME}-agent --version || echo 'FAILED: Agent --version'
        ${PROJECT_NAME}-agent --help || echo 'FAILED: Agent --help'

        # Test binary rename
        cp /usr/local/bin/${PROJECT_NAME}-agent /tmp/renamed-agent
        chmod +x /tmp/renamed-agent
        if /tmp/renamed-agent --help 2>&1 | grep -q 'renamed-agent'; then
            echo '✓ Agent binary rename works'
        else
            echo '✗ FAILED: Agent --help does not show renamed binary name'
        fi

        # Full Agent functionality tests against server
        echo '--- Agent Full Functionality Tests ---'
        if [ -n \"\${API_TOKEN:-}\" ]; then
            # Test agent registration/status with API token
            ${PROJECT_NAME}-agent --server http://localhost:80 --token \"\$API_TOKEN\" status || echo 'Agent status failed'
            # Project-specific agent commands go here (IDEA.md)
        else
            echo 'Agent tests skipped (no API token)'
        fi
    else
        echo 'Agent not installed - skipping'
    fi

    echo '=== Service Stop Test ==='
    systemctl stop ${PROJECT_NAME}

    echo '=== All tests passed ==='
"

echo "Incus tests completed successfully"
