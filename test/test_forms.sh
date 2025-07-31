#!/bin/bash

# Test script for Pepo form endpoints
set -e

BASE_URL="http://localhost:8080"
FORMS_URL="$BASE_URL/forms"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to check if server is running
check_server() {
    if curl -s -f "$BASE_URL/health" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to wait for server to start
wait_for_server() {
    print_status "Waiting for server to start..."
    for i in {1..30}; do
        if check_server; then
            print_success "Server is running!"
            return 0
        fi
        sleep 1
        echo -n "."
    done
    print_error "Server failed to start after 30 seconds"
    return 1
}

# Function to start the server
start_server() {
    print_status "Starting the server..."
    ./bin/pepo > server.log 2>&1 &
    SERVER_PID=$!
    echo $SERVER_PID > server.pid

    if wait_for_server; then
        return 0
    else
        print_error "Failed to start server"
        cat server.log
        return 1
    fi
}

# Function to stop the server
stop_server() {
    if [ -f server.pid ]; then
        SERVER_PID=$(cat server.pid)
        print_status "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        rm -f server.pid
        wait $SERVER_PID 2>/dev/null || true
    fi
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    stop_server
    rm -f server.log
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Test functions
test_list_persons_form() {
    print_status "Testing list people form endpoint..."
    response=$(curl -s -w "%{http_code}" "$FORMS_URL/persons/list")
    http_code="${response: -3}"
    body="${response%???}"

    if [ "$http_code" = "200" ]; then
        print_success "List people form endpoint works"
        echo "Response length: ${#body} characters"

        # Check if it's HTML
        if echo "$body" | grep -q "<div\|No people found"; then
            print_success "Response appears to be HTML"
        else
            print_warning "Response doesn't look like HTML"
            echo "First 200 chars: ${body:0:200}"
        fi
    else
        print_error "List people form endpoint failed (HTTP $http_code)"
        echo "Response: $body"
        return 1
    fi
}

test_create_person_form() {
    print_status "Testing create person form endpoint..."
    response=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "name=Test User Form" \
        "$FORMS_URL/persons/create")

    http_code="${response: -3}"
    body="${response%???}"

    if [ "$http_code" = "200" ]; then
        print_success "Person created successfully via form"
        echo "Response length: ${#body} characters"

        # Extract person ID if possible
        if echo "$body" | grep -q "person-"; then
            PERSON_ID=$(echo "$body" | grep -o 'person-[^"]*' | head -1 | cut -d'-' -f2)
            print_success "Created person with ID: $PERSON_ID"
        fi

        # Check if it's HTML
        if echo "$body" | grep -q "<div"; then
            print_success "Response appears to be HTML"
        else
            print_warning "Response doesn't look like HTML"
            echo "Response: $body"
        fi
    else
        print_error "Failed to create person via form (HTTP $http_code)"
        echo "Response: $body"
        return 1
    fi
}

test_create_person_form_empty_name() {
    print_status "Testing create person form with empty name..."
    response=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "name=" \
        "$FORMS_URL/persons/create")

    http_code="${response: -3}"
    body="${response%???}"

    if [ "$http_code" = "400" ]; then
        print_success "Form validation works (empty name rejected)"
        if echo "$body" | grep -q "Name is required"; then
            print_success "Error message is correct"
        else
            print_warning "Error message might be incorrect: $body"
        fi
    else
        print_error "Form validation failed (HTTP $http_code)"
        echo "Response: $body"
        return 1
    fi
}

test_delete_person_form() {
    if [ -z "$PERSON_ID" ]; then
        print_warning "Skipping delete person test - no person ID available"
        return 0
    fi

    print_status "Testing delete person form endpoint..."
    response=$(curl -s -w "%{http_code}" -X DELETE \
        "$FORMS_URL/persons/delete/$PERSON_ID")

    http_code="${response: -3}"

    if [ "$http_code" = "200" ]; then
        print_success "Person deleted successfully via form"
    else
        print_error "Failed to delete person via form (HTTP $http_code)"
        return 1
    fi
}

test_web_interface() {
    print_status "Testing main web interface..."
    response=$(curl -s -w "%{http_code}" "$BASE_URL/")
    http_code="${response: -3}"
    body="${response%???}"

    if [ "$http_code" = "200" ]; then
        print_success "Web interface loads"

        # Check for HTMX integration
        if echo "$body" | grep -q "htmx.org"; then
            print_success "HTMX is loaded"
        else
            print_warning "HTMX might not be loaded"
        fi

        # Check for form endpoints
        if echo "$body" | grep -q "/forms/persons"; then
            print_success "Form endpoints are referenced"
        else
            print_warning "Form endpoints might not be properly referenced"
        fi

        # Check for Tailwind CSS
        if echo "$body" | grep -q "tailwindcss.com"; then
            print_success "Tailwind CSS is loaded"
        else
            print_warning "Tailwind CSS might not be loaded"
        fi
    else
        print_error "Web interface failed to load (HTTP $http_code)"
        return 1
    fi
}

# Main execution
main() {
    echo "=========================================="
    echo "  Pepo Form Endpoints Tests"
    echo "=========================================="

    # Check if binary exists
    if [ ! -f "./bin/pepo" ]; then
        print_error "Binary not found. Please run 'make build' first."
        exit 1
    fi

    # Check if PostgreSQL is running
    if ! docker ps | grep -q pepo-postgres; then
        print_warning "PostgreSQL container not running. Starting it..."
        make docker-up >/dev/null 2>&1
        sleep 5
    fi

    # Start server
    if ! start_server; then
        print_error "Failed to start server"
        exit 1
    fi

    # Initialize test counters
    TOTAL_TESTS=0
    PASSED_TESTS=0

    # Initialize variables
    PERSON_ID=""

    # Run tests
    tests=(
        "test_web_interface"
        "test_list_persons_form"
        "test_create_person_form_empty_name"
        "test_create_person_form"
        "test_delete_person_form"
    )

    for test in "${tests[@]}"; do
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        echo ""
        echo "----------------------------------------"
        if $test; then
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            print_error "Test $test failed"
        fi
        echo "----------------------------------------"
    done

    # Summary
    echo ""
    echo "=========================================="
    echo "  Form Tests Summary"
    echo "=========================================="
    echo "Total tests: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $((TOTAL_TESTS - PASSED_TESTS))"

    if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
        print_success "All form tests passed! 🎉"
        echo ""
        print_status "Try the web interface at: http://localhost:8080"
        exit 0
    else
        print_error "Some form tests failed! ❌"
        exit 1
    fi
}

# Check if script is being executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
