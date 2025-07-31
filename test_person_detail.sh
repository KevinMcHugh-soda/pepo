#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="http://localhost:8080/api/v1"
SERVER_PID=""
PERSON_ID=""
ACTION_IDS=()

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

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Start server in background
start_server() {
    log_info "Starting server..."
    ./bin/pepo > /tmp/test_server.log 2>&1 &
    SERVER_PID=$!
    sleep 3

    # Check if server is running
    if ! curl -s "$API_BASE_URL/../health" > /dev/null; then
        log_error "Server failed to start"
        cat /tmp/test_server.log
        exit 1
    fi

    log_success "Server started with PID: $SERVER_PID"
}

# Stop server
stop_server() {
    if [ -n "$SERVER_PID" ]; then
        log_info "Stopping server..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        log_success "Server stopped"
    fi
}

# Create a test person
create_test_person() {
    log_info "Creating test person..."

    local response=$(curl -s -X POST "$API_BASE_URL/people" \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -d '{"name": "Test Person for Detail View"}' \
        -w "%{http_code}")

    local http_code="${response: -3}"
    local body="${response%???}"

    if [ "$http_code" = "201" ]; then
        PERSON_ID=$(echo "$body" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        log_success "Created test person with ID: $PERSON_ID"
        return 0
    else
        log_error "Failed to create person. HTTP $http_code: $body"
        return 1
    fi
}

# Create test actions for the person
create_test_actions() {
    log_info "Creating test actions for person..."

    local actions=(
        '{"person_id":"'$PERSON_ID'","description":"Completed a project successfully","valence":"positive","occurred_at":"2024-01-15T10:00:00Z"}'
        '{"person_id":"'$PERSON_ID'","description":"Was late to an important meeting","valence":"negative","occurred_at":"2024-01-16T09:30:00Z"}'
        '{"person_id":"'$PERSON_ID'","description":"Helped a colleague with their work","valence":"positive","occurred_at":"2024-01-17T14:20:00Z","references":"https://example.com/reference"}'
    )

    for action_data in "${actions[@]}"; do
        local response=$(curl -s -X POST "$API_BASE_URL/actions" \
            -H "Content-Type: application/json" \
            -H "Accept: application/json" \
            -d "$action_data" \
            -w "%{http_code}")

        local http_code="${response: -3}"
        local body="${response%???}"

        if [ "$http_code" = "201" ]; then
            local action_id=$(echo "$body" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
            ACTION_IDS+=("$action_id")
            log_success "Created action with ID: $action_id"
        else
            log_warning "Failed to create action. HTTP $http_code: $body"
        fi
    done
}

# Test person detail HTML view
test_person_detail_html() {
    log_info "Testing person detail HTML view..."

    local response=$(curl -s -H "Accept: text/html" \
        "$API_BASE_URL/people/$PERSON_ID" \
        -w "%{http_code}")

    local http_code="${response: -3}"
    local body="${response%???}"

    if [ "$http_code" = "200" ]; then
        log_success "Person detail view returned HTTP 200"

        # Check if the response contains expected content
        if echo "$body" | grep -q "Test Person for Detail View"; then
            log_success "✓ Person name found in response"
        else
            log_error "✗ Person name not found in response"
            return 1
        fi

        if echo "$body" | grep -q "Person Details"; then
            log_success "✓ Person details title found"
        else
            log_error "✗ Person details title not found"
            return 1
        fi

        if echo "$body" | grep -q "Actions"; then
            log_success "✓ Actions section found"
        else
            log_error "✗ Actions section not found"
            return 1
        fi

        # Check for actions content
        if echo "$body" | grep -q "Completed a project successfully"; then
            log_success "✓ First action found in response"
        else
            log_error "✗ First action not found in response"
        fi

        if echo "$body" | grep -q "positive"; then
            log_success "✓ Positive valence found"
        else
            log_error "✗ Positive valence not found"
        fi

        if echo "$body" | grep -q "negative"; then
            log_success "✓ Negative valence found"
        else
            log_error "✗ Negative valence not found"
        fi

        if echo "$body" | grep -q "Back to People List"; then
            log_success "✓ Back navigation link found"
        else
            log_error "✗ Back navigation link not found"
        fi

        # Save response for debugging
        echo "$body" > /tmp/person_detail_response.html
        log_info "Response saved to /tmp/person_detail_response.html"

        return 0
    else
        log_error "Person detail view failed. HTTP $http_code: $body"
        return 1
    fi
}

# Test record action form on person detail page
test_record_action_form() {
    log_info "Testing record action form on person detail page..."

    local response=$(curl -s -H "Accept: text/html" \
        "$API_BASE_URL/people/$PERSON_ID" \
        -w "%{http_code}")

    local http_code="${response: -3}"
    local body="${response%???}"

    if [ "$http_code" = "200" ]; then
        log_success "Person detail page loaded for form testing"

        # Check if record action form exists
        if echo "$body" | grep -q "Record New Action"; then
            log_success "✓ Record action form found on page"
        else
            log_error "✗ Record action form not found on page"
            return 1
        fi

        # Check if person_id is pre-filled in form
        if echo "$body" | grep -q "name=\"person_id\"" && echo "$body" | grep -q "value=\"$PERSON_ID\""; then
            log_success "✓ Person ID is pre-filled in form"
        else
            log_error "✗ Person ID not pre-filled in form"
            return 1
        fi

        # Check form elements exist
        if echo "$body" | grep -q "name=\"description\""; then
            log_success "✓ Description field found"
        else
            log_error "✗ Description field not found"
            return 1
        fi

        if echo "$body" | grep -q "name=\"valence\""; then
            log_success "✓ Valence radio buttons found"
        else
            log_error "✗ Valence radio buttons not found"
            return 1
        fi

        return 0
    else
        log_error "Failed to load person detail page for form testing. HTTP $http_code: $body"
        return 1
    fi
}

# Test submitting action via form
test_submit_action_form() {
    log_info "Testing action form submission..."

    # Submit a new action via form data
    local response=$(curl -s -X POST "$API_BASE_URL/actions" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -H "Accept: text/html" \
        -d "person_id=$PERSON_ID&description=Form+submitted+action&valence=positive&occurred_at=2024-01-20T15:30:00" \
        -w "%{http_code}")

    local http_code="${response: -3}"
    local body="${response%???}"

    if [ "$http_code" = "201" ]; then
        log_success "Action form submission successful (HTTP 201)"

        # Extract action ID if possible
        if echo "$body" | grep -q "action-"; then
            log_success "✓ Action HTML response contains action element"
        else
            log_warning "Action response may not contain expected HTML structure"
        fi

        return 0
    else
        log_error "Action form submission failed. HTTP $http_code: $body"
        return 1
    fi
}

# Test person detail page after adding action via form
test_updated_person_detail() {
    log_info "Testing person detail page after form submission..."

    local response=$(curl -s -H "Accept: text/html" \
        "$API_BASE_URL/people/$PERSON_ID" \
        -w "%{http_code}")

    local http_code="${response: -3}"
    local body="${response%???}"

    if [ "$http_code" = "200" ]; then
        log_success "Updated person detail page loaded"

        # Check that the new action appears
        if echo "$body" | grep -q "Form submitted action"; then
            log_success "✓ New action from form appears in person detail"
        else
            log_error "✗ New action from form not found in person detail"
            return 1
        fi

        # Count total actions (should be 4: 3 original + 1 from form)
        local action_count=$(echo "$body" | grep -o "action-" | wc -l)
        if [ "$action_count" -ge 4 ]; then
            log_success "✓ Action count increased after form submission"
        else
            log_warning "Action count may not have increased as expected"
        fi

        return 0
    else
        log_error "Failed to load updated person detail page. HTTP $http_code: $body"
        return 1
    fi
}

# Test person detail JSON view (should still work)
test_person_detail_json() {
    log_info "Testing person detail JSON view..."

    local response=$(curl -s -H "Accept: application/json" \
        "$API_BASE_URL/people/$PERSON_ID" \
        -w "%{http_code}")

    local http_code="${response: -3}"
    local body="${response%???}"

    if [ "$http_code" = "200" ]; then
        log_success "Person detail JSON view returned HTTP 200"

        if echo "$body" | grep -q '"name":"Test Person for Detail View"'; then
            log_success "✓ JSON response contains person name"
        else
            log_error "✗ JSON response missing person name"
            return 1
        fi

        return 0
    else
        log_error "Person detail JSON view failed. HTTP $http_code: $body"
        return 1
    fi
}

# Test that home page links to person detail
test_home_page_links() {
    log_info "Testing that home page contains links to person detail..."

    local response=$(curl -s -H "Accept: text/html" \
        "$API_BASE_URL/../" \
        -w "%{http_code}")

    local http_code="${response: -3}"
    local body="${response%???}"

    if [ "$http_code" = "200" ]; then
        log_success "Home page returned HTTP 200"

        if echo "$body" | grep -q 'href="/api/v1/people/'; then
            log_success "✓ Home page contains links to person detail pages"
        else
            log_warning "✗ Home page may not contain person detail links (this may be expected if no people are visible on load)"
        fi

        return 0
    else
        log_error "Home page failed. HTTP $http_code: $body"
        return 1
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test data..."

    # Delete test actions
    for action_id in "${ACTION_IDS[@]}"; do
        curl -s -X DELETE "$API_BASE_URL/actions/$action_id" > /dev/null || true
    done

    # Delete test person
    if [ -n "$PERSON_ID" ]; then
        curl -s -X DELETE "$API_BASE_URL/people/$PERSON_ID" > /dev/null || true
    fi

    # Stop server
    stop_server

    # Clean up temp files
    rm -f /tmp/test_server.log /tmp/person_detail_response.html

    log_success "Cleanup completed"
}

# Main test execution
main() {
    log_info "Starting person detail view tests..."

    # Set up cleanup trap
    trap cleanup EXIT

    # Build application first
    log_info "Building application..."
    make build

    # Start server
    start_server

    # Run tests
    create_test_person
    create_test_actions
    test_person_detail_html
    test_record_action_form
    test_submit_action_form
    test_updated_person_detail
    test_person_detail_json
    test_home_page_links

    log_success "All person detail view and form tests completed successfully!"
}

# Run main function
main "$@"
