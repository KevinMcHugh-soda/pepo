#!/bin/bash

# Test script to verify 'people' terminology is used correctly in user-facing text
# This script checks that we successfully changed "persons" to "people" in UI text
# while keeping the technical API endpoints unchanged

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

# Test that HTML responses use "people" instead of "persons"
test_html_terminology() {
    log_test "Testing HTML responses use 'people' terminology..."

    # Get HTML response for persons list
    response=$(curl -s -H "Accept: text/html" "$API_BASE_URL/persons" -o /tmp/people_html.txt -w "%{http_code}")
    status="${response: -3}"

    if [[ "$status" == "200" ]]; then
        log_success "HTML persons list endpoint works (HTTP $status)"

        # Check that the response doesn't contain "persons" in user-visible text
        if grep -qi "persons" /tmp/people_html.txt 2>/dev/null; then
            log_error "Found 'persons' in HTML response - should use 'people'"
            echo "Problematic content:"
            grep -i "persons" /tmp/people_html.txt || true
            return 1
        else
            log_success "HTML response correctly avoids 'persons' terminology"
        fi

        # Check that it contains expected "people" friendly text
        if grep -qi "people\|person" /tmp/people_html.txt 2>/dev/null; then
            log_success "HTML response contains people-friendly terminology"
        else
            log_info "HTML response doesn't contain obvious people terminology (might be empty list)"
        fi
    else
        log_error "HTML persons list failed: HTTP $status"
        cat /tmp/people_html.txt
        return 1
    fi
}

# Test the main page for user-friendly terminology
test_main_page() {
    log_test "Testing main page uses friendly terminology..."

    response=$(curl -s "$BASE_URL/" -o /tmp/main_page.txt -w "%{http_code}")
    status="${response: -3}"

    if [[ "$status" == "200" ]]; then
        log_success "Main page loads successfully (HTTP $status)"

        # Check for user-friendly text
        if grep -q "People" /tmp/main_page.txt 2>/dev/null; then
            log_success "Main page uses 'People' in headings"
        else
            log_error "Main page should contain 'People' heading"
            return 1
        fi

        if grep -q "Add New Person" /tmp/main_page.txt 2>/dev/null; then
            log_success "Main page uses 'Add New Person' text"
        else
            log_error "Main page should contain 'Add New Person' text"
            return 1
        fi

        # Check that loading messages use friendly terminology
        if grep -q "Loading people" /tmp/main_page.txt 2>/dev/null; then
            log_success "Loading messages use 'people' terminology"
        elif grep -q "Loading persons" /tmp/main_page.txt 2>/dev/null; then
            log_error "Found 'Loading persons' - should be 'Loading people'"
            return 1
        else
            log_info "No loading messages found in static HTML"
        fi

    else
        log_error "Main page failed to load: HTTP $status"
        cat /tmp/main_page.txt
        return 1
    fi
}

# Test select options format uses friendly terminology
test_select_options() {
    log_test "Testing select options use friendly terminology..."

    response=$(curl -s -H "Accept: text/html" "$API_BASE_URL/persons?format=select" -o /tmp/select_options.txt -w "%{http_code}")
    status="${response: -3}"

    if [[ "$status" == "200" ]]; then
        log_success "Select options endpoint works (HTTP $status)"

        # Check for friendly option text
        if grep -q "Select a person" /tmp/select_options.txt 2>/dev/null; then
            log_success "Select options use friendly 'Select a person' text"
        else
            log_info "No 'Select a person' text found (might be empty)"
        fi

        # Check for unfriendly terminology
        if grep -qi "persons" /tmp/select_options.txt 2>/dev/null; then
            log_error "Found 'persons' in select options - should use 'people'"
            cat /tmp/select_options.txt
            return 1
        else
            log_success "Select options avoid 'persons' terminology"
        fi

        # Check for error messages
        if grep -q "Error loading people" /tmp/select_options.txt 2>/dev/null; then
            log_success "Error messages use 'people' terminology"
        elif grep -q "Error loading persons" /tmp/select_options.txt 2>/dev/null; then
            log_error "Found 'Error loading persons' - should be 'people'"
            return 1
        else
            log_info "No error messages found in select options"
        fi

    else
        log_error "Select options failed: HTTP $status"
        cat /tmp/select_options.txt
        return 1
    fi
}

# Test that API endpoints still work (technical names unchanged)
test_api_endpoints_unchanged() {
    log_test "Testing API endpoints still work with original names..."

    # Test that the /persons endpoint still works for JSON
    json_response=$(curl -s -H "Accept: application/json" "$API_BASE_URL/persons" -o /tmp/api_check.txt -w "%{http_code}")
    json_status="${json_response: -3}"

    if [[ "$json_status" == "200" ]]; then
        log_success "API endpoint /api/v1/persons still works (HTTP $json_status)"

        # Verify it returns JSON with the expected structure
        if command -v jq >/dev/null 2>&1; then
            if jq . /tmp/api_check.txt > /dev/null 2>&1; then
                log_success "API returns valid JSON"

                # Check that JSON still uses the technical field names
                if jq -e '.persons' /tmp/api_check.txt > /dev/null 2>&1; then
                    log_success "JSON response correctly uses 'persons' field name (technical)"
                else
                    log_info "JSON response structure might be different or empty"
                fi
            else
                log_error "API response is not valid JSON"
                cat /tmp/api_check.txt
                return 1
            fi
        fi
    else
        log_error "API endpoint failed: HTTP $json_status"
        cat /tmp/api_check.txt
        return 1
    fi
}

# Test error messages use friendly terminology
test_error_messages() {
    log_test "Testing error messages use friendly terminology..."

    # Try to trigger an error by sending invalid data
    error_response=$(curl -s -X POST "$API_BASE_URL/persons" \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -d '{"invalid": "data"}' \
        -o /tmp/error_response.txt -w "%{http_code}")

    error_status="${error_response: -3}"

    if [[ "$error_status" =~ ^4[0-9][0-9]$ ]]; then
        log_success "Error endpoint responds as expected (HTTP $error_status)"

        # Check if error message uses friendly terminology
        if grep -qi "persons" /tmp/error_response.txt 2>/dev/null; then
            log_error "Error message contains 'persons' - should use 'people'"
            cat /tmp/error_response.txt
            return 1
        else
            log_success "Error messages avoid 'persons' terminology"
        fi
    else
        log_info "No error triggered or unexpected response: HTTP $error_status"
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    rm -f /tmp/people_html.txt /tmp/main_page.txt /tmp/select_options.txt /tmp/api_check.txt /tmp/error_response.txt
    log_success "Cleanup completed"
}

# Main test execution
main() {
    log_info "Starting people terminology verification tests..."
    echo ""

    # Set cleanup trap
    trap cleanup EXIT

    # Run tests
    check_server
    echo ""

    test_main_page
    echo ""

    test_html_terminology
    echo ""

    test_select_options
    echo ""

    test_api_endpoints_unchanged
    echo ""

    test_error_messages
    echo ""

    log_success "All terminology tests completed successfully!"
    echo ""
    log_info "Summary of verification:"
    log_info "  ✓ Main page uses 'People' and 'Add New Person'"
    log_info "  ✓ HTML responses avoid 'persons' terminology"
    log_info "  ✓ Select options use friendly text"
    log_info "  ✓ API endpoints still work with technical names"
    log_info "  ✓ Error messages use friendly terminology"
    echo ""
    log_success "User interface now uses friendly 'people' terminology! 👥"
}

# Run the tests
main "$@"
