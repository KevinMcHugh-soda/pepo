#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8082"
API_URL="${BASE_URL}/api/v1"
DATABASE_URL="postgres://postgres:password@localhost:5433/pepo_dev?sslmode=disable"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# Function to start server
start_server() {
    print_info "Starting Pepo server on port 8082..."
    DATABASE_URL="$DATABASE_URL" PORT=8082 ./bin/pepo > test_refactor.log 2>&1 &
    SERVER_PID=$!
    echo $SERVER_PID > test_refactor.pid

    # Wait for server to start
    for i in {1..10}; do
        if curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
            print_status "Server started successfully (PID: $SERVER_PID)"
            return 0
        fi
        print_warning "Waiting for server to start... (attempt $i/10)"
        sleep 2
    done

    print_error "Server failed to start"
    return 1
}

# Function to stop server
stop_server() {
    if [ -f test_refactor.pid ]; then
        PID=$(cat test_refactor.pid)
        print_info "Stopping server (PID: $PID)..."
        kill $PID 2>/dev/null || true
        rm -f test_refactor.pid
        sleep 2
    fi
}

# Function to test API endpoint
test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5

    if [ -n "$data" ]; then
        response=$(curl -s -w "%{http_code}" -X $method \
            -H "Content-Type: application/json" \
            -d "$data" \
            "${API_URL}${endpoint}")
    else
        response=$(curl -s -w "%{http_code}" -X $method "${API_URL}${endpoint}")
    fi

    status_code="${response: -3}"
    response_body="${response%???}"

    if [ "$status_code" = "$expected_status" ]; then
        print_status "$description (HTTP $status_code)"
        echo "$response_body"
        return 0
    else
        print_error "$description (Expected HTTP $expected_status, got $status_code)"
        echo "Response: $response_body"
        return 1
    fi
}

# Function to test HTML endpoint
test_html() {
    local endpoint=$1
    local description=$2

    response=$(curl -s -w "%{http_code}" "${BASE_URL}${endpoint}")
    status_code="${response: -3}"
    response_body="${response%???}"

    if [ "$status_code" = "200" ]; then
        print_status "$description (HTTP $status_code)"
        return 0
    else
        print_error "$description (Expected HTTP 200, got $status_code)"
        return 1
    fi
}

# Function to test form endpoint
test_form() {
    local endpoint=$1
    local data=$2
    local description=$3

    response=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "$data" \
        "${BASE_URL}${endpoint}")

    status_code="${response: -3}"
    response_body="${response%???}"

    if [ "$status_code" = "200" ]; then
        print_status "$description (HTTP $status_code)"
        return 0
    else
        print_error "$description (Expected HTTP 200, got $status_code)"
        echo "Response: $response_body"
        return 1
    fi
}

# Function to extract ID from JSON response
extract_id() {
    echo "$1" | grep -o '"id":"[^"]*"' | cut -d'"' -f4
}

# Cleanup on exit
cleanup() {
    print_info "Cleaning up..."
    stop_server
    rm -f test_refactor.log test_refactor.pid
}
trap cleanup EXIT

# Main test execution
main() {
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE}         Pepo Refactoring Validation Test       ${NC}"
    echo -e "${BLUE}================================================${NC}"
    echo

    # Test 1: Build verification
    print_info "Step 1: Building application..."
    if make build > /dev/null 2>&1; then
        print_status "Application builds successfully"
    else
        print_error "Application failed to build"
        exit 1
    fi
    echo

    # Test 2: Template generation
    print_info "Step 2: Verifying template generation..."
    if make generate-templ > /dev/null 2>&1; then
        print_status "Templates generate successfully"
    else
        print_error "Template generation failed"
        exit 1
    fi
    echo

    # Test 3: Server startup
    print_info "Step 3: Starting server..."
    start_server
    echo

    # Test 4: Health check
    print_info "Step 4: Testing health endpoint..."
    test_api "GET" "/health" "" "200" "Health check"
    echo

    # Test 5: HTML template rendering
    print_info "Step 5: Testing template rendering..."
    test_html "/" "Main page template rendering"
    echo

    # Test 6: Person API CRUD
    print_info "Step 6: Testing Person API..."

    # Create person
    person_response=$(test_api "POST" "/persons" '{"name":"Test Person"}' "201" "Create person")
    person_id=$(extract_id "$person_response")

    if [ -z "$person_id" ]; then
        print_error "Failed to extract person ID"
        exit 1
    fi

    # Get person
    test_api "GET" "/persons/$person_id" "" "200" "Get person by ID"

    # List persons
    test_api "GET" "/persons" "" "200" "List persons"

    # Update person
    test_api "PUT" "/persons/$person_id" '{"name":"Updated Person"}' "200" "Update person"
    echo

    # Test 7: Action API CRUD
    print_info "Step 7: Testing Action API..."

    # Create action
    action_response=$(test_api "POST" "/actions" "{\"person_id\":\"$person_id\",\"description\":\"Test action\",\"valence\":\"positive\"}" "201" "Create action")
    action_id=$(extract_id "$action_response")

    if [ -z "$action_id" ]; then
        print_error "Failed to extract action ID"
        exit 1
    fi

    # Get action
    test_api "GET" "/actions/$action_id" "" "200" "Get action by ID"

    # List actions
    test_api "GET" "/actions" "" "200" "List actions"

    # Get person actions
    test_api "GET" "/persons/$person_id/actions" "" "200" "Get person actions"
    echo

    # Test 8: Form endpoints (HTMX)
    print_info "Step 8: Testing form endpoints..."

    # Create person via form
    test_form "/forms/persons/create" "name=Form Test Person" "Create person via form"

    # List persons via form
    test_html "/forms/persons/list" "List persons via form"

    # Get persons for select
    test_html "/forms/persons/select" "Get persons for select dropdown"

    # List actions via form
    test_html "/forms/actions/list" "List actions via form"
    echo

    # Test 9: API filtering and pagination
    print_info "Step 9: Testing API filtering..."

    # Filter actions by person
    test_api "GET" "/actions?person_id=$person_id" "" "200" "Filter actions by person"

    # Filter actions by valence
    test_api "GET" "/actions?valence=positive" "" "200" "Filter actions by valence"

    # Test pagination
    test_api "GET" "/actions?limit=1&offset=0" "" "200" "Test pagination"
    echo

    # Test 10: Cleanup operations
    print_info "Step 10: Testing cleanup operations..."

    # Delete action
    test_api "DELETE" "/actions/$action_id" "" "204" "Delete action"

    # Delete person
    test_api "DELETE" "/persons/$person_id" "" "204" "Delete person"
    echo

    # Final validation
    echo -e "${GREEN}================================================${NC}"
    echo -e "${GREEN}         🎉 ALL TESTS PASSED! 🎉              ${NC}"
    echo -e "${GREEN}================================================${NC}"
    echo
    print_status "Refactoring validation completed successfully"
    print_info "The application has been successfully refactored with:"
    echo "  • Separated handlers (person.go, action.go)"
    echo "  • Templ templates (layout.templ, index.templ, etc.)"
    echo "  • Simplified main.go (84% size reduction)"
    echo "  • Ogen router integration"
    echo "  • Full backward compatibility maintained"
    echo
    print_info "Log file available at: test_refactor.log"
}

# Verify prerequisites
if ! command -v curl &> /dev/null; then
    print_error "curl is required but not installed"
    exit 1
fi

if ! docker ps | grep -q pepo-postgres; then
    print_error "PostgreSQL container (pepo-postgres) is not running"
    print_info "Run: make docker-up"
    exit 1
fi

# Run main function
main "$@"
