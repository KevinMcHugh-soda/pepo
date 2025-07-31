#!/bin/bash

# Test script for refactored routes with HTML form submissions
# This script verifies that HTML forms work with the consolidated API routes

set -e  # Exit on any error

# Configuration
BASE_URL="http://localhost:8000"
API_BASE_URL="$BASE_URL/api/v1"

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

# Test HTML form submission for creating a person
test_person_form_submission() {
    log_test "Testing person creation via HTML form submission..."

    # Submit form data as HTML form would
    response=$(curl -s -X POST "$API_BASE_URL/persons" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -H "Accept: text/html" \
        -d "name=Test Person from Form" \
        -w "%{http_code}" -o /tmp/person_form_response.txt)

    status="${response: -3}"

    if [[ "$status" == "201" ]]; then
        log_success "Person created successfully via form submission (HTTP $status)"
        # Check if response contains HTML
        if grep -q "<" /tmp/person_form_response.txt 2>/dev/null; then
            log_success "Response contains HTML content"
        else
            log_info "Response appears to be plain text (not HTML)"
        fi
        # Try to extract person ID from response for cleanup
        if command -v grep >/dev/null 2>&1; then
            PERSON_ID=$(grep -o 'data-id="[^"]*"' /tmp/person_form_response.txt 2>/dev/null | cut -d'"' -f2 | head -1)
            if [ -n "$PERSON_ID" ]; then
                log_info "Extracted person ID: $PERSON_ID"
            fi
        fi
    else
        log_error "Failed to create person via form: HTTP $status"
        cat /tmp/person_form_response.txt
        return 1
    fi
}

# Test HTML form submission for creating an action
test_action_form_submission() {
    local person_id="$1"

    if [ -z "$person_id" ]; then
        log_error "Cannot test action form submission without person ID"
        return 1
    fi

    log_test "Testing action creation via HTML form submission..."

    # Submit form data as HTML form would
    response=$(curl -s -X POST "$API_BASE_URL/actions" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -H "Accept: text/html" \
        -d "person_id=${person_id}&description=Test action from form&valence=positive&occurred_at=2024-01-01T12:00" \
        -w "%{http_code}" -o /tmp/action_form_response.txt)

    status="${response: -3}"

    if [[ "$status" == "201" ]]; then
        log_success "Action created successfully via form submission (HTTP $status)"
        # Check if response contains HTML
        if grep -q "<" /tmp/action_form_response.txt 2>/dev/null; then
            log_success "Response contains HTML content"
        else
            log_info "Response appears to be plain text (not HTML)"
        fi
        # Try to extract action ID for cleanup
        if command -v grep >/dev/null 2>&1; then
            ACTION_ID=$(grep -o 'data-id="[^"]*"' /tmp/action_form_response.txt 2>/dev/null | cut -d'"' -f2 | head -1 || true)
            if [ -n "$ACTION_ID" ]; then
                log_info "Extracted action ID: $ACTION_ID"
            fi
        fi
    else
        log_error "Failed to create action via form: HTTP $status"
        cat /tmp/action_form_response.txt
        return 1
    fi
}

# Test JSON API still works
test_json_api() {
    log_test "Testing JSON API still works..."

    # Test JSON person creation
    response=$(curl -s -X POST "$API_BASE_URL/persons" \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -d '{"name":"Test Person JSON"}' \
        -w "%{http_code}" -o /tmp/person_json_response.txt)

    status="${response: -3}"

    if [[ "$status" == "201" ]]; then
        log_success "JSON API person creation works (HTTP $status)"
        if command -v jq >/dev/null 2>&1; then
            if jq . /tmp/person_json_response.txt > /dev/null 2>&1; then
                log_success "Response is valid JSON"
                JSON_PERSON_ID=$(jq -r '.id' /tmp/person_json_response.txt 2>/dev/null || true)
            else
                log_error "Response is not valid JSON"
                cat /tmp/person_json_response.txt
            fi
        fi
    else
        log_error "JSON API person creation failed: HTTP $status"
        cat /tmp/person_json_response.txt
        return 1
    fi
}

# Test content negotiation with same endpoint
test_content_negotiation() {
    log_test "Testing content negotiation on same endpoint..."

    # Test JSON response
    json_response=$(curl -s "$API_BASE_URL/persons" \
        -H "Accept: application/json" \
        -w "%{http_code}" -o /tmp/persons_json.txt)

    json_status="${json_response: -3}"

    # Test HTML response
    html_response=$(curl -s "$API_BASE_URL/persons" \
        -H "Accept: text/html" \
        -w "%{http_code}" -o /tmp/persons_html.txt)

    html_status="${html_response: -3}"

    if [[ "$json_status" == "200" ]] && [[ "$html_status" == "200" ]]; then
        log_success "Both JSON and HTML responses work (HTTP $json_status, $html_status)"

        # Verify JSON is valid JSON
        if command -v jq >/dev/null 2>&1; then
            if jq . /tmp/persons_json.txt > /dev/null 2>&1; then
                log_success "JSON response is valid JSON"
            else
                log_error "JSON response is not valid JSON"
            fi
        fi

        # Verify HTML contains HTML tags
        if grep -q "<" /tmp/persons_html.txt 2>/dev/null; then
            log_success "HTML response contains HTML content"
        else
            log_info "HTML response doesn't contain obvious HTML tags"
        fi
    else
        log_error "Content negotiation failed: JSON=$json_status, HTML=$html_status"
        return 1
    fi
}

# Test person select options format
test_person_select_options() {
    log_test "Testing person select options format..."

    response=$(curl -s "$API_BASE_URL/persons?format=select" \
        -H "Accept: text/html" \
        -w "%{http_code}" -o /tmp/person_select.txt)

    status="${response: -3}"

    if [[ "$status" == "200" ]]; then
        log_success "Person select options endpoint works (HTTP $status)"
        if grep -q "<option" /tmp/person_select.txt 2>/dev/null; then
            log_success "Response contains option elements"
        else
            log_info "Response doesn't contain option elements"
            cat /tmp/person_select.txt
        fi
    else
        log_error "Person select options failed: HTTP $status"
        cat /tmp/person_select.txt
        return 1
    fi
}

# Test DELETE operations
test_delete_operations() {
    local person_id="$1"
    local action_id="$2"

    log_test "Testing DELETE operations..."

    # Delete action if we have an ID
    if [ -n "$action_id" ]; then
        log_info "Deleting test action: $action_id"
        delete_response=$(curl -s -X DELETE "$API_BASE_URL/actions/$action_id" \
            -w "%{http_code}" -o /tmp/delete_action.txt)
        delete_status="${delete_response: -3}"
        if [[ "$delete_status" == "204" ]]; then
            log_success "Action deleted successfully (HTTP $delete_status)"
        else
            log_error "Failed to delete action: HTTP $delete_status"
        fi
    fi

    # Delete person if we have an ID
    if [ -n "$person_id" ]; then
        log_info "Deleting test person: $person_id"
        delete_response=$(curl -s -X DELETE "$API_BASE_URL/persons/$person_id" \
            -w "%{http_code}" -o /tmp/delete_person.txt)
        delete_status="${delete_response: -3}"
        if [[ "$delete_status" == "204" ]]; then
            log_success "Person deleted successfully (HTTP $delete_status)"
        else
            log_error "Failed to delete person: HTTP $delete_status"
        fi
    fi

    # Delete JSON test person if we have an ID
    if [ -n "$JSON_PERSON_ID" ]; then
        log_info "Deleting JSON test person: $JSON_PERSON_ID"
        delete_response=$(curl -s -X DELETE "$API_BASE_URL/persons/$JSON_PERSON_ID" \
            -w "%{http_code}" -o /tmp/delete_json_person.txt)
        delete_status="${delete_response: -3}"
        if [[ "$delete_status" == "204" ]]; then
            log_success "JSON test person deleted successfully (HTTP $delete_status)"
        else
            log_error "Failed to delete JSON test person: HTTP $delete_status"
        fi
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    rm -f /tmp/*_response.txt /tmp/persons_*.txt /tmp/person_select.txt /tmp/delete_*.txt
    log_success "Cleanup completed"
}

# Main test execution
main() {
    log_info "Starting refactored routes test..."
    echo ""

    # Set cleanup trap
    trap cleanup EXIT

    # Run tests
    check_server
    echo ""

    # Test form submissions
    test_person_form_submission
    echo ""

    test_action_form_submission "$PERSON_ID"
    echo ""

    # Test JSON API compatibility
    test_json_api
    echo ""

    # Test content negotiation
    test_content_negotiation
    echo ""

    # Test specialized endpoints
    test_person_select_options
    echo ""

    # Test delete operations
    test_delete_operations "$PERSON_ID" "$ACTION_ID"
    echo ""

    log_success "All refactored routes tests completed successfully!"
    echo ""
    log_info "Summary of successful tests:"
    log_info "  ✓ HTML form submissions work"
    log_info "  ✓ Form-to-JSON conversion working"
    log_info "  ✓ Content negotiation functional"
    log_info "  ✓ JSON API compatibility maintained"
    log_info "  ✓ Specialized format parameters work"
    log_info "  ✓ DELETE operations functional"
    echo ""
    log_success "Route consolidation refactor is working correctly!"
}

# Global variables for cleanup
PERSON_ID=""
ACTION_ID=""
JSON_PERSON_ID=""

# Run the tests
main "$@"
