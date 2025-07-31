#!/bin/bash

# Test script for Pepo Performance Tracking API
set -e

BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

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
test_health() {
    print_status "Testing health endpoint..."
    response=$(curl -s -w "%{http_code}" "$BASE_URL/health")
    http_code="${response: -3}"
    body="${response%???}"

    if [ "$http_code" = "200" ]; then
        print_success "Health check passed"
        echo "Response: $body"
    else
        print_error "Health check failed (HTTP $http_code)"
        return 1
    fi
}

test_create_person() {
    print_status "Testing create person..."
    response=$(curl -s -w "%{http_code}" -X POST "$API_URL/persons" \
        -H "Content-Type: application/json" \
        -d '{"name": "John Doe"}')

    http_code="${response: -3}"
    body="${response%???}"

    if [ "$http_code" = "201" ]; then
        print_success "Person created successfully"
        PERSON_ID=$(echo "$body" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        echo "Created person with ID: $PERSON_ID"
        echo "Response: $body"
    else
        print_error "Failed to create person (HTTP $http_code)"
        echo "Response: $body"
        return 1
    fi
}

test_list_persons() {
    print_status "Testing list people..."
    response=$(curl -s -w "%{http_code}" "$API_URL/persons")

    http_code="${response: -3}"
    body="${response%???}"

    if [ "$http_code" = "200" ]; then
        print_success "List people successful"
        echo "Response: $body"
    else
        print_error "Failed to list people (HTTP $http_code)"
        echo "Response: $body"
        return 1
    fi
}

test_get_person() {
    if [ -z "$PERSON_ID" ]; then
        print_warning "Skipping get person test - no person ID available"
        return 0
    fi

    print_status "Testing get person by ID..."
    response=$(curl -s -w "%{http_code}" "$API_URL/persons/$PERSON_ID")

    http_code="${response: -3}"
    body="${response%???}"

    if [ "$http_code" = "200" ]; then
        print_success "Get person successful"
        echo "Response: $body"
    else
        print_error "Failed to get person (HTTP $http_code)"
        echo "Response: $body"
        return 1
    fi
}

test_update_person() {
    if [ -z "$PERSON_ID" ]; then
        print_warning "Skipping update person test - no person ID available"
        return 0
    fi

    print_status "Testing update person..."
    response=$(curl -s -w "%{http_code}" -X PUT "$API_URL/persons/$PERSON_ID" \
        -H "Content-Type: application/json" \
        -d '{"name": "Jane Doe"}')

    http_code="${response: -3}"
    body="${response%???}"

    if [ "$http_code" = "200" ]; then
        print_success "Person updated successfully"
        echo "Response: $body"
    else
        print_error "Failed to update person (HTTP $http_code)"
        echo "Response: $body"
        return 1
    fi
}

test_delete_person() {
    if [ -z "$PERSON_ID" ]; then
        print_warning "Skipping delete person test - no person ID available"
        return 0
    fi

    print_status "Testing delete person..."
    response=$(curl -s -w "%{http_code}" -X DELETE "$API_URL/persons/$PERSON_ID")

    http_code="${response: -3}"

    if [ "$http_code" = "204" ]; then
        print_success "Person deleted successfully"
    else
        print_error "Failed to delete person (HTTP $http_code)"
        return 1
    fi
}

test_xid_demo() {
    print_status "Testing XID generation demo..."
    response=$(curl -s -w "%{http_code}" "$API_URL/demo/xid")

    http_code="${response: -3}"
    body="${response%???}"

    if [ "$http_code" = "200" ]; then
        print_success "XID demo successful"
        echo "Response: $body"
    else
        print_error "XID demo failed (HTTP $http_code)"
        echo "Response: $body"
        return 1
    fi
}

# Main execution
main() {
    echo "=========================================="
    echo "  Pepo Performance Tracking API Tests"
    echo "=========================================="

    # Check if binary exists
    if [ ! -f "./bin/pepo" ]; then
        print_error "Binary not found. Please run 'make build' first."
        exit 1
    fi

    # Check if PostgreSQL is running
    if ! docker ps | grep -q pepo-postgres; then
        print_warning "PostgreSQL container not running. Starting it..."
        make docker-up
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

    # Run tests
    tests=(
        "test_health"
        "test_xid_demo"
        "test_create_person"
        "test_list_persons"
        "test_get_person"
        "test_update_person"
        "test_delete_person"
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
    echo "  Test Summary"
    echo "=========================================="
    echo "Total tests: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $((TOTAL_TESTS - PASSED_TESTS))"

    if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
        print_success "All tests passed! 🎉"
        exit 0
    else
        print_error "Some tests failed! ❌"
        exit 1
    fi
}

# Check if script is being sourced or executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
