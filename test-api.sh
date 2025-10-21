#!/bin/bash
# CASCI API Test Script - Complete workflow test

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Find the port CASCI is running on
PORT=8080
for p in 8080 64000 64500; do
    if curl -s http://localhost:$p/health > /dev/null 2>&1; then
        PORT=$p
        break
    fi
done

BASE_URL="http://localhost:$PORT"
API_URL="$BASE_URL/api/v1"

echo -e "${BLUE}╔════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   CASCI API Test Suite            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════╝${NC}"
echo ""

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

test_start() {
    echo -n "  ▶ $1... "
}

test_pass() {
    echo -e "${GREEN}✓${NC}"
    ((TESTS_PASSED++))
}

test_fail() {
    echo -e "${RED}✗${NC}"
    if [ -n "$1" ]; then
        echo "    Error: $1"
    fi
    ((TESTS_FAILED++))
}

# Check if server is running
echo -e "${YELLOW}[1/8] Server Health${NC}"
test_start "Checking server health"
if curl -s -f "$BASE_URL/health" > /dev/null; then
    test_pass
    echo "    Server running on port $PORT"
else
    test_fail "Server is not running"
    echo ""
    echo "Please start CASCI first: make run or ./casci"
    exit 1
fi
echo ""

# Authentication
echo -e "${YELLOW}[2/8] Authentication${NC}"
test_start "Registering user"
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","email":"test@example.com","password":"testpass123"}')

if echo "$REGISTER_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    test_pass
else
    # User exists, try login
    LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser","password":"testpass123"}')

    if echo "$LOGIN_RESPONSE" | grep -q '"token"'; then
        TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        test_pass
    else
        test_fail "Failed to authenticate"
        exit 1
    fi
fi

test_start "Getting user profile"
USER_RESPONSE=$(curl -s -X GET "$API_URL/users/me" -H "Authorization: Bearer $TOKEN")
if echo "$USER_RESPONSE" | grep -q '"username":"testuser"'; then
    test_pass
else
    test_fail
fi
echo ""

# Projects
echo -e "${YELLOW}[3/8] Project Management${NC}"
test_start "Creating project"
PROJECT_RESPONSE=$(curl -s -X POST "$API_URL/projects/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name":"hello-world","repository_url":"https://github.com/octocat/Hello-World","branch":"master"}')

if echo "$PROJECT_RESPONSE" | grep -q '"id"'; then
    PROJECT_ID=$(echo "$PROJECT_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    test_pass
    echo "    Project ID: $PROJECT_ID"
else
    test_fail "Response: $PROJECT_RESPONSE"
    exit 1
fi

test_start "Listing projects"
if curl -s -X GET "$API_URL/projects" -H "Authorization: Bearer $TOKEN" | grep -q "hello-world"; then
    test_pass
else
    test_fail
fi

test_start "Getting project details"
if curl -s -X GET "$API_URL/projects/$PROJECT_ID" -H "Authorization: Bearer $TOKEN" | grep -q "hello-world"; then
    test_pass
else
    test_fail
fi
echo ""

# Builds
echo -e "${YELLOW}[4/8] Build Triggering${NC}"
test_start "Triggering build"
BUILD_RESPONSE=$(curl -s -X POST "$API_URL/projects/$PROJECT_ID/builds" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"branch":"master","trigger":"manual"}')

if echo "$BUILD_RESPONSE" | grep -q '"build_number"'; then
    BUILD_ID=$(echo "$BUILD_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    BUILD_NUMBER=$(echo "$BUILD_RESPONSE" | grep -o '"build_number":[0-9]*' | cut -d':' -f2)
    test_pass
    echo "    Build #$BUILD_NUMBER (ID: $BUILD_ID)"
else
    test_fail "Response: $BUILD_RESPONSE"
fi

test_start "Getting build details"
if curl -s -X GET "$API_URL/builds/$BUILD_ID" -H "Authorization: Bearer $TOKEN" | grep -q "build_number"; then
    test_pass
else
    test_fail
fi

test_start "Listing builds"
if curl -s -X GET "$API_URL/projects/$PROJECT_ID/builds" -H "Authorization: Bearer $TOKEN" | grep -q "build_number"; then
    test_pass
else
    test_fail
fi
echo ""

# Build Status
echo -e "${YELLOW}[5/8] Build Execution${NC}"
echo "    Waiting for build to process..."
sleep 3

test_start "Checking build status"
BUILD_STATUS_RESPONSE=$(curl -s -X GET "$API_URL/builds/$BUILD_ID" -H "Authorization: Bearer $TOKEN")
if echo "$BUILD_STATUS_RESPONSE" | grep -q '"status"'; then
    BUILD_STATUS=$(echo "$BUILD_STATUS_RESPONSE" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
    test_pass
    echo "    Status: $BUILD_STATUS"
else
    test_fail
fi

test_start "Getting build statistics"
if curl -s -X GET "$API_URL/projects/$PROJECT_ID/builds/stats" -H "Authorization: Bearer $TOKEN" | grep -q "total_builds"; then
    test_pass
else
    test_fail
fi
echo ""

# Update
echo -e "${YELLOW}[6/8] Project Updates${NC}"
test_start "Updating project"
if curl -s -X PUT "$API_URL/projects/$PROJECT_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name":"hello-world-updated"}' | grep -q "hello-world-updated"; then
    test_pass
else
    test_fail
fi
echo ""

# Jenkins API
echo -e "${YELLOW}[7/8] Jenkins Compatibility${NC}"
test_start "Jenkins API endpoint"
if curl -s "$BASE_URL/api/json" | grep -q '"mode":"NORMAL"'; then
    test_pass
else
    test_fail
fi

test_start "Jenkins Crumb issuer"
if curl -s "$BASE_URL/crumbIssuer/api/json" | grep -q '"crumb"'; then
    test_pass
else
    test_fail
fi
echo ""

# Cleanup
echo -e "${YELLOW}[8/8] Cleanup${NC}"
test_start "Deleting project"
DELETE_RESPONSE=$(curl -s -w "%{http_code}" -X DELETE "$API_URL/projects/$PROJECT_ID" -H "Authorization: Bearer $TOKEN")
if echo "$DELETE_RESPONSE" | grep -q "200"; then
    test_pass
else
    test_fail
fi
echo ""

# Summary
echo -e "${BLUE}╔════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Test Results                     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════╝${NC}"
echo -e "  Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "  Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi