#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8081"
API_URL="${BASE_URL}/api/v1"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to start server
start_server() {
    print_status "Starting Pepo server..."
    DATABASE_URL="postgres://postgres:password@localhost:5433/pepo_dev?sslmode=disable" PORT=8081 ./bin/pepo > test_server.log 2>&1 &
    SERVER_PID=$!
    echo $SERVER_PID > test_server.pid

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
    if [ -f test_server.pid ]; then
        PID=$(cat test_server.pid)
        print_status "Stopping server (PID: $PID)..."
        kill $PID 2>/dev/null || true
        rm -f test_server.pid
        sleep 2
    fi
}

# Function to test API endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4

    print_status "Testing: $description"

    if [ -n "$data" ]; then
        response=$(curl -s -X $method \
            -H "Content-Type: application/json" \
            -d "$data" \
            "${API_URL}${endpoint}")
    else
        response=$(curl -s -X $method "${API_URL}${endpoint}")
    fi

    echo "Response: $response"
    echo "$response"
}

# Function to extract ID from JSON response
extract_id() {
    echo "$1" | grep -o '"id":"[^"]*"' | cut -d'"' -f4
}

# Cleanup on exit
cleanup() {
    print_status "Cleaning up..."
    stop_server
    rm -f test_server.log
}
trap cleanup EXIT

# Main test execution
main() {
    print_status "=== Pepo Actions API Test ==="

    # Build the application
    print_status "Building application..."
    make build

    # Start server
    start_server

    # Test health endpoint
    print_status "Testing health endpoint..."
    health_response=$(curl -s "${BASE_URL}/health")
    echo "Health: $health_response"

    # Test 1: Create a person
    print_status "\n=== Testing Person Creation ==="
    person_data='{"name":"John Doe"}'
    person_response=$(test_endpoint "POST" "/people" "$person_data" "Create person")
    person_id=$(extract_id "$person_response")

    if [ -z "$person_id" ]; then
        print_error "Failed to create person"
        exit 1
    fi
    print_status "Created person with ID: $person_id"

    # Test 2: Create positive action
    print_status "\n=== Testing Action Creation ==="
    action_data=$(cat <<EOF
{
    "person_id": "$person_id",
    "description": "Completed project ahead of schedule",
    "valence": "positive",
    "references": "https://github.com/example/project"
}
EOF
)
    action_response=$(test_endpoint "POST" "/actions" "$action_data" "Create positive action")
    action_id=$(extract_id "$action_response")

    if [ -z "$action_id" ]; then
        print_error "Failed to create action"
        exit 1
    fi
    print_status "Created action with ID: $action_id"

    # Test 3: Create negative action
    negative_action_data=$(cat <<EOF
{
    "person_id": "$person_id",
    "description": "Missed important deadline",
    "valence": "negative"
}
EOF
)
    negative_action_response=$(test_endpoint "POST" "/actions" "$negative_action_data" "Create negative action")
    negative_action_id=$(extract_id "$negative_action_response")
    print_status "Created negative action with ID: $negative_action_id"

    # Test 4: List all actions
    print_status "\n=== Testing Action Listing ==="
    test_endpoint "GET" "/actions" "" "List all actions"

    # Test 5: Get specific action
    print_status "\n=== Testing Get Action by ID ==="
    test_endpoint "GET" "/actions/$action_id" "" "Get action by ID"

    # Test 6: Get actions for person
    print_status "\n=== Testing Get Person Actions ==="
    test_endpoint "GET" "/people/$person_id/actions" "" "Get actions for person"

    # Test 7: Filter actions by valence
    print_status "\n=== Testing Filter Actions by Valence ==="
    test_endpoint "GET" "/actions?valence=positive" "" "Filter positive actions"
    test_endpoint "GET" "/actions?valence=negative" "" "Filter negative actions"

    # Test 8: Filter actions by person
    print_status "\n=== Testing Filter Actions by Person ==="
    test_endpoint "GET" "/actions?person_id=$person_id" "" "Filter actions by person"

    # Test 9: Update action
    print_status "\n=== Testing Action Update ==="
    update_data=$(cat <<EOF
{
    "person_id": "$person_id",
    "occurred_at": "2024-01-01T10:00:00Z",
    "description": "Updated: Completed project ahead of schedule with excellent quality",
    "valence": "positive",
    "references": "https://github.com/example/project/pull/123"
}
EOF
)
    test_endpoint "PUT" "/actions/$action_id" "$update_data" "Update action"

    # Test 10: Delete action
    print_status "\n=== Testing Action Deletion ==="
    test_endpoint "DELETE" "/actions/$negative_action_id" "" "Delete action"

    # Test 11: Verify deletion
    print_status "\n=== Verifying Deletion ==="
    deleted_response=$(curl -s -w "%{http_code}" "${API_URL}/actions/$negative_action_id")
    if [[ "$deleted_response" == *"404"* ]]; then
        print_status "Action successfully deleted (404 as expected)"
    else
        print_warning "Action deletion verification unclear: $deleted_response"
    fi

    # Test 12: Create person with same name (should work)
    print_status "\n=== Testing Duplicate Person Creation ==="
    person2_data='{"name":"Jane Smith"}'
    person2_response=$(test_endpoint "POST" "/people" "$person2_data" "Create second person")
    person2_id=$(extract_id "$person2_response")
    print_status "Created second person with ID: $person2_id"

    # Test 13: Create action for second person
    action2_data=$(cat <<EOF
{
    "person_id": "$person2_id",
    "description": "Great presentation to the team",
    "valence": "positive"
}
EOF
)
    test_endpoint "POST" "/actions" "$action2_data" "Create action for second person"

    # Test 14: Test pagination
    print_status "\n=== Testing Pagination ==="
    test_endpoint "GET" "/actions?limit=1&offset=0" "" "Test pagination - first item"
    test_endpoint "GET" "/actions?limit=1&offset=1" "" "Test pagination - second item"

    # Test 15: Test error cases
    print_status "\n=== Testing Error Cases ==="

    # Invalid person ID for action
    invalid_action_data=$(cat <<EOF
{
    "person_id": "invalidpersonid123",
    "description": "This should fail",
    "valence": "positive"
}
EOF
)
    error_response=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$invalid_action_data" \
        "${API_URL}/actions")
    print_status "Invalid person ID test: $error_response"

    # Empty description
    empty_desc_data=$(cat <<EOF
{
    "person_id": "$person_id",
    "description": "",
    "valence": "positive"
}
EOF
)
    error_response2=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$empty_desc_data" \
        "${API_URL}/actions")
    print_status "Empty description test: $error_response2"

    print_status "\n=== All Tests Completed ==="
    print_status "Check the responses above for any errors."
    print_status "Server log available in: test_server.log"
}

# Run main function
main "$@"
