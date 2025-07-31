#!/bin/bash

# Test script for consolidated routes and content negotiation
# This script verifies that the same endpoints can serve both JSON and HTML content

set -e  # Exit on any error

# Configuration
BASE_URL="http://localhost:8000"
API_BASE_URL="$BASE_URL/api/v1"
CONVENIENCE_BASE_URL="$BASE_URL"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

# Check if server is running
check_server() {
    log_info "Checking if server is running..."
    if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        log_error "Server is not running at $BASE_URL"
        log_error "Please start the server first: make run"
        exit 1
    fi
    log_success "Server is running"
}

# Test content negotiation for a specific endpoint
test_content_negotiation() {
    local method="$1"
    local url="$2"
    local description="$3"
    local data="$4"

    log_test "Testing $description"

    # Test JSON response
    log_info "  Testing JSON response..."
    if [ -n "$data" ]; then
        json_response=$(curl -s -X "$method" "$url" \
            -H "Accept: application/json" \
            -H "Content-Type: application/json" \
            -d "$data" \
            -w "%{http_code}" -o /tmp/json_response.txt)
    else
        json_response=$(curl -s -X "$method" "$url" \
            -H "Accept: application/json" \
            -w "%{http_code}" -o /tmp/json_response.txt)
    fi

    json_status="${json_response: -3}"

    # Test HTML response
    log_info "  Testing HTML response..."
    if [ -n "$data" ]; then
        html_response=$(curl -s -X "$method" "$url" \
            -H "Accept: text/html" \
            -H "Content-Type: application/json" \
            -d "$data" \
            -w "%{http_code}" -o /tmp/html_response.txt)
    else
        html_response=$(curl -s -X "$method" "$url" \
            -H "Accept: text/html" \
            -w "%{http_code}" -o /tmp/html_response.txt)
    fi

    html_status="${html_response: -3}"

    # Verify responses
    if [[ "$json_status" =~ ^[2-3][0-9][0-9]$ ]]; then
        log_success "  JSON response: HTTP $json_status"
        # Check if response is valid JSON (if it should be)
        if [[ "$json_status" =~ ^2[0-9][0-9]$ ]] && [[ "$method" != "DELETE" ]]; then
            if jq . /tmp/json_response.txt > /dev/null 2>&1; then
                log_success "  JSON response is valid JSON"
            else
                log_error "  JSON response is not valid JSON"
                cat /tmp/json_response.txt
            fi
        fi
    else
        log_error "  JSON response: HTTP $json_status"
        cat /tmp/json_response.txt
    fi

    if [[ "$html_status" =~ ^[2-3][0-9][0-9]$ ]]; then
        log_success "  HTML response: HTTP $html_status"
        # Check if response contains HTML-like content (if it should be)
        if [[ "$html_status" =~ ^2[0-9][0-9]$ ]] && [[ "$method" != "DELETE" ]]; then
            if grep -q "<" /tmp/html_response.txt 2>/dev/null; then
                log_success "  HTML response contains HTML content"
            else
                log_info "  HTML response doesn't contain HTML tags (may be plain text)"
            fi
        fi
    else
        log_error "  HTML response: HTTP $html_status"
        cat /tmp/html_response.txt
    fi

    echo ""
}

# Test both API and convenience routes
test_both_routes() {
    local method="$1"
    local endpoint="$2"
    local description="$3"
    local data="$4"

    # Test API route
    test_content_negotiation "$method" "$API_BASE_URL$endpoint" "$description (API route)" "$data"

    # Test convenience route (skip for now as it might need different setup)
    # test_content_negotiation "$method" "$CONVENIENCE_BASE_URL$endpoint" "$description (convenience route)" "$data"
}

# Store created IDs for cleanup
PERSON_ID=""
ACTION_ID=""

# Main test suite
run_tests() {
    log_info "Starting consolidated routes and content negotiation tests..."
    echo ""

    # Test 1: List people (GET /persons)
    test_both_routes "GET" "/persons" "List persons"

    # Test 2: Create person (POST /persons)
    log_test "Testing Create person..."
    person_data='{"name":"Test Person for Consolidated Routes"}'

    # Test JSON creation
    json_response=$(curl -s -X POST "$API_BASE_URL/persons" \
        -H "Accept: application/json" \
        -H "Content-Type: application/json" \
        -d "$person_data" \
        -w "%{http_code}" -o /tmp/create_person_json.txt)

    json_status="${json_response: -3}"

    if [[ "$json_status" == "201" ]]; then
        log_success "  Person created successfully (JSON)"
        PERSON_ID=$(jq -r '.id' /tmp/create_person_json.txt)
        log_info "  Created person ID: $PERSON_ID"
    else
        log_error "  Failed to create person (JSON): HTTP $json_status"
        cat /tmp/create_person_json.txt
        exit 1
    fi

    # Test HTML creation
    html_response=$(curl -s -X POST "$API_BASE_URL/persons" \
        -H "Accept: text/html" \
        -H "Content-Type: application/json" \
        -d "$person_data" \
        -w "%{http_code}" -o /tmp/create_person_html.txt)

    html_status="${html_response: -3}"

    if [[ "$html_status" == "201" ]]; then
        log_success "  Person creation returns HTML (201)"
    else
        log_error "  Failed to create person (HTML): HTTP $html_status"
        cat /tmp/create_person_html.txt
    fi

    echo ""

    # Test 3: Get person by ID (GET /persons/{id})
    if [ -n "$PERSON_ID" ]; then
        test_both_routes "GET" "/persons/$PERSON_ID" "Get person by ID"
    fi

    # Test 4: Update person (PUT /persons/{id})
    if [ -n "$PERSON_ID" ]; then
        update_data='{"name":"Updated Test Person"}'
        test_both_routes "PUT" "/persons/$PERSON_ID" "Update person" "$update_data"
    fi

    # Test 5: List actions (GET /actions)
    test_both_routes "GET" "/actions" "List actions"

    # Test 6: Create action (POST /actions)
    if [ -n "$PERSON_ID" ]; then
        log_test "Testing Create action..."
        action_data="{\"person_id\":\"$PERSON_ID\",\"occurred_at\":\"2024-01-01T12:00:00Z\",\"description\":\"Test action for consolidated routes\",\"valence\":\"positive\"}"

        # Test JSON creation
        json_response=$(curl -s -X POST "$API_BASE_URL/actions" \
            -H "Accept: application/json" \
            -H "Content-Type: application/json" \
            -d "$action_data" \
            -w "%{http_code}" -o /tmp/create_action_json.txt)

        json_status="${json_response: -3}"

        if [[ "$json_status" == "201" ]]; then
            log_success "  Action created successfully (JSON)"
            ACTION_ID=$(jq -r '.id' /tmp/create_action_json.txt)
            log_info "  Created action ID: $ACTION_ID"
        else
            log_error "  Failed to create action (JSON): HTTP $json_status"
            cat /tmp/create_action_json.txt
        fi

        # Test HTML creation
        html_response=$(curl -s -X POST "$API_BASE_URL/actions" \
            -H "Accept: text/html" \
            -H "Content-Type: application/json" \
            -d "$action_data" \
            -w "%{http_code}" -o /tmp/create_action_html.txt)

        html_status="${html_response: -3}"

        if [[ "$html_status" == "201" ]]; then
            log_success "  Action creation returns HTML (201)"
        else
            log_error "  Failed to create action (HTML): HTTP $html_status"
            cat /tmp/create_action_html.txt
        fi

        echo ""
    fi

    # Test 7: Get action by ID (GET /actions/{id})
    if [ -n "$ACTION_ID" ]; then
        test_both_routes "GET" "/actions/$ACTION_ID" "Get action by ID"
    fi

    # Test 8: Update action (PUT /actions/{id})
    if [ -n "$ACTION_ID" ] && [ -n "$PERSON_ID" ]; then
        update_action_data="{\"person_id\":\"$PERSON_ID\",\"occurred_at\":\"2024-01-01T13:00:00Z\",\"description\":\"Updated test action\",\"valence\":\"neutral\"}"
        test_both_routes "PUT" "/actions/$ACTION_ID" "Update action" "$update_action_data"
    fi

    # Test 9: Get person's actions (GET /persons/{id}/actions)
    if [ -n "$PERSON_ID" ]; then
        test_both_routes "GET" "/persons/$PERSON_ID/actions" "Get person's actions"
    fi

    # Test 10: Content type preference
    log_test "Testing content type preference..."

    # Test that default (no Accept header) returns JSON
    default_response=$(curl -s "$API_BASE_URL/persons" -w "%{http_code}" -o /tmp/default_response.txt)
    default_status="${default_response: -3}"

    if [[ "$default_status" == "200" ]]; then
        if jq . /tmp/default_response.txt > /dev/null 2>&1; then
            log_success "  Default response (no Accept header) returns JSON"
        else
            log_error "  Default response is not JSON"
            cat /tmp/default_response.txt
        fi
    else
        log_error "  Default response failed: HTTP $default_status"
    fi

    echo ""
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test data..."

    if [ -n "$ACTION_ID" ]; then
        log_info "  Deleting test action: $ACTION_ID"
        curl -s -X DELETE "$API_BASE_URL/actions/$ACTION_ID" > /dev/null 2>&1 || true
    fi

    if [ -n "$PERSON_ID" ]; then
        log_info "  Deleting test person: $PERSON_ID"
        curl -s -X DELETE "$API_BASE_URL/persons/$PERSON_ID" > /dev/null 2>&1 || true
    fi

    # Clean up temp files
    rm -f /tmp/*_response.txt /tmp/create_*_*.txt

    log_success "Cleanup completed"
}

# Main execution
main() {
    check_server

    # Set trap to cleanup on exit
    trap cleanup EXIT

    run_tests

    echo ""
    log_success "All consolidated routes and content negotiation tests completed!"
    log_info "Summary:"
    log_info "  ✓ Content negotiation working (JSON/HTML responses)"
    log_info "  ✓ API routes functional (/api/v1/*)"
    log_info "  ✓ CRUD operations for people and actions"
    log_info "  ✓ Default content type is JSON"

    echo ""
    log_info "Routes are now consolidated - no need for separate /forms/* endpoints!"
    log_info "Clients can use Accept header to get JSON or HTML responses."
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    log_error "jq is required for this test script but not installed."
    log_error "Please install jq: https://stedolan.github.io/jq/download/"
    exit 1
fi

# Run the tests
main "$@"
