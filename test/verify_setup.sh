#!/bin/bash

# Pepo Performance Tracking - Setup Verification Script
# This script verifies that all components of the project are properly set up

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Emojis for visual feedback
CHECK="✅"
CROSS="❌"
WARN="⚠️"
INFO="ℹ️"
ROCKET="🚀"

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0

# Function to print colored output
print_header() {
    echo -e "${PURPLE}================================================${NC}"
    echo -e "${PURPLE}  $1${NC}"
    echo -e "${PURPLE}================================================${NC}"
    echo ""
}

print_section() {
    echo -e "${CYAN}▶ $1${NC}"
}

print_check() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}${CHECK} $1${NC}"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        echo -e "  ${RED}${CROSS} $1${NC}"
    fi
}

print_success() {
    echo -e "${GREEN}${CHECK} $1${NC}"
}

print_error() {
    echo -e "${RED}${CROSS} $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}${WARN} $1${NC}"
}

print_info() {
    echo -e "${BLUE}${INFO} $1${NC}"
}

# Verification functions
verify_go_environment() {
    print_section "Go Environment"

    # Check Go installation
    if command -v go >/dev/null 2>&1; then
        GO_VERSION=$(go version | cut -d' ' -f3)
        print_success "Go is installed: $GO_VERSION"
    else
        print_error "Go is not installed"
        return 1
    fi

    # Check Go module
    if [ -f "go.mod" ]; then
        MODULE_NAME=$(head -1 go.mod | cut -d' ' -f2)
        print_success "Go module exists: $MODULE_NAME"
    else
        print_error "go.mod file not found"
        return 1
    fi

    # Check dependencies
    if go mod verify >/dev/null 2>&1; then
        print_success "Go dependencies are valid"
    else
        print_error "Go dependencies verification failed"
        return 1
    fi

    echo ""
}

verify_development_tools() {
    print_section "Development Tools"

    # Check dbmate
    if [ -f "$HOME/go/bin/dbmate" ]; then
        DBMATE_VERSION=$(~/go/bin/dbmate --version 2>/dev/null | head -1 || echo "unknown")
        print_success "dbmate is installed: $DBMATE_VERSION"
    else
        print_error "dbmate is not installed"
        return 1
    fi

    # Check sqlc
    if [ -f "$HOME/go/bin/sqlc" ]; then
        SQLC_VERSION=$(~/go/bin/sqlc version 2>/dev/null || echo "unknown")
        print_success "sqlc is installed: $SQLC_VERSION"
    else
        print_error "sqlc is not installed"
        return 1
    fi

    # Check ogen
    if [ -f "$HOME/go/bin/ogen" ]; then
        print_success "ogen is installed"
    else
        print_error "ogen is not installed"
        return 1
    fi

    echo ""
}

verify_database() {
    print_section "Database Setup"

    # Check if PostgreSQL container is running
    if docker ps | grep -q pepo-postgres; then
        print_success "PostgreSQL container is running"
    else
        print_warning "PostgreSQL container is not running - starting it..."
        if make docker-up >/dev/null 2>&1; then
            sleep 3
            print_success "PostgreSQL container started"
        else
            print_error "Failed to start PostgreSQL container"
            return 1
        fi
    fi

    # Check database connection
    if ~/go/bin/dbmate status >/dev/null 2>&1; then
        print_success "Database connection successful"
    else
        print_error "Database connection failed"
        return 1
    fi

    # Check migrations
    if [ -f "db/schema.sql" ]; then
        print_success "Database schema exists"
    else
        print_error "Database schema not found"
        return 1
    fi

    # Check Person table
    MIGRATION_COUNT=$(ls db/migrations/*.sql 2>/dev/null | wc -l)
    if [ $MIGRATION_COUNT -gt 0 ]; then
        print_success "Database migrations exist ($MIGRATION_COUNT found)"
    else
        print_error "No database migrations found"
        return 1
    fi

    echo ""
}

verify_code_generation() {
    print_section "Code Generation"

    # Check API specification
    if [ -f "api/openapi.yaml" ]; then
        print_success "OpenAPI specification exists"
    else
        print_error "OpenAPI specification not found"
        return 1
    fi

    # Check generated API code
    if [ -d "internal/api" ] && [ "$(ls -A internal/api 2>/dev/null)" ]; then
        API_FILES=$(ls internal/api/*.go 2>/dev/null | wc -l)
        print_success "Generated API code exists ($API_FILES files)"
    else
        print_error "Generated API code not found"
        return 1
    fi

    # Check SQL queries
    if [ -f "db/queries/persons.sql" ]; then
        print_success "SQL queries exist"
    else
        print_error "SQL queries not found"
        return 1
    fi

    # Check generated database code
    if [ -d "internal/db" ] && [ "$(ls -A internal/db 2>/dev/null)" ]; then
        DB_FILES=$(ls internal/db/*.go 2>/dev/null | wc -l)
        print_success "Generated database code exists ($DB_FILES files)"
    else
        print_error "Generated database code not found"
        return 1
    fi

    # Check sqlc configuration
    if [ -f "sqlc.yaml" ]; then
        print_success "sqlc configuration exists"
    else
        print_error "sqlc configuration not found"
        return 1
    fi

    echo ""
}

verify_build_system() {
    print_section "Build System"

    # Check Makefile
    if [ -f "Makefile" ]; then
        print_success "Makefile exists"
    else
        print_error "Makefile not found"
        return 1
    fi

    # Check if application builds
    if make build >/dev/null 2>&1; then
        print_success "Application builds successfully"
    else
        print_error "Application build failed"
        return 1
    fi

    # Check binary exists
    if [ -f "bin/pepo" ]; then
        BINARY_SIZE=$(ls -lh bin/pepo | awk '{print $5}')
        print_success "Application binary exists ($BINARY_SIZE)"
    else
        print_error "Application binary not found"
        return 1
    fi

    echo ""
}

verify_project_structure() {
    print_section "Project Structure"

    # Essential directories
    REQUIRED_DIRS=("api" "cmd/server" "db/migrations" "db/queries" "internal" "static" "templates")

    for dir in "${REQUIRED_DIRS[@]}"; do
        if [ -d "$dir" ]; then
            print_success "Directory exists: $dir"
        else
            print_error "Missing directory: $dir"
            return 1
        fi
    done

    # Essential files
    REQUIRED_FILES=("README.MD" "SETUP.md" "PROJECT_SUMMARY.md" "go.mod" "docker-compose.yml" ".env" ".gitignore")

    for file in "${REQUIRED_FILES[@]}"; do
        if [ -f "$file" ]; then
            print_success "File exists: $file"
        else
            print_error "Missing file: $file"
            return 1
        fi
    done

    echo ""
}

verify_api_functionality() {
    print_section "API Functionality"

    # Start server in background
    print_info "Starting server for API tests..."
    ./bin/pepo > /dev/null 2>&1 &
    SERVER_PID=$!

    # Wait for server to start
    sleep 3

    # Test health endpoint
    if curl -s -f "http://localhost:8080/health" >/dev/null; then
        print_success "Health endpoint responds"
    else
        print_error "Health endpoint failed"
        kill $SERVER_PID 2>/dev/null || true
        return 1
    fi

    # Test API endpoint
    if curl -s -f "http://localhost:8080/api/v1/persons" >/dev/null; then
        print_success "API endpoint responds"
    else
        print_error "API endpoint failed"
        kill $SERVER_PID 2>/dev/null || true
        return 1
    fi

    # Test XID generation
    if curl -s -f "http://localhost:8080/api/v1/demo/xid" >/dev/null; then
        print_success "XID generation works"
    else
        print_error "XID generation failed"
        kill $SERVER_PID 2>/dev/null || true
        return 1
    fi

    # Stop server
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true

    echo ""
}

verify_docker_setup() {
    print_section "Docker Setup"

    # Check Docker
    if command -v docker >/dev/null 2>&1; then
        print_success "Docker is available"
    else
        print_warning "Docker not found (optional)"
    fi

    # Check docker-compose file
    if [ -f "docker-compose.yml" ]; then
        print_success "docker-compose.yml exists"
    else
        print_error "docker-compose.yml not found"
        return 1
    fi

    echo ""
}

display_summary() {
    print_header "Verification Summary"

    echo -e "Total checks performed: ${BLUE}$TOTAL_CHECKS${NC}"
    echo -e "Checks passed: ${GREEN}$PASSED_CHECKS${NC}"
    echo -e "Checks failed: ${RED}$((TOTAL_CHECKS - PASSED_CHECKS))${NC}"
    echo ""

    if [ $PASSED_CHECKS -eq $TOTAL_CHECKS ]; then
        echo -e "${GREEN}${ROCKET} SETUP VERIFICATION COMPLETE! ${ROCKET}${NC}"
        echo -e "${GREEN}Your Pepo Performance Tracking application is ready!${NC}"
        echo ""
        echo -e "${BLUE}Quick Start Commands:${NC}"
        echo -e "  ${CYAN}make run${NC}          - Start the application"
        echo -e "  ${CYAN}make test-api${NC}     - Run API tests"
        echo -e "  ${CYAN}make dev${NC}          - Complete development setup"
        echo ""
        echo -e "${BLUE}Access Points:${NC}"
        echo -e "  ${CYAN}http://localhost:8080${NC}     - Web application"
        echo -e "  ${CYAN}http://localhost:8080/api/v1${NC} - API endpoints"
        echo -e "  ${CYAN}http://localhost:5050${NC}     - pgAdmin (admin@pepo.local / admin)"
        echo ""
        return 0
    else
        echo -e "${RED}${CROSS} SETUP VERIFICATION FAILED${NC}"
        echo -e "${YELLOW}Please review the failed checks above and fix them.${NC}"
        echo ""
        echo -e "${BLUE}For help, check:${NC}"
        echo -e "  ${CYAN}SETUP.md${NC}         - Detailed setup instructions"
        echo -e "  ${CYAN}PROJECT_SUMMARY.md${NC} - Complete project overview"
        echo -e "  ${CYAN}make help${NC}        - Available commands"
        echo ""
        return 1
    fi
}

# Main execution
main() {
    print_header "Pepo Performance Tracking - Setup Verification"

    # Change to project directory if not already there
    if [ ! -f "go.mod" ] && [ -d "pepo" ]; then
        cd pepo
    fi

    # Verify we're in the right directory
    if [ ! -f "go.mod" ] || ! grep -q "module pepo" go.mod 2>/dev/null; then
        print_error "Please run this script from the pepo project directory"
        exit 1
    fi

    # Run verification steps
    verify_go_environment
    verify_development_tools
    verify_database
    verify_code_generation
    verify_build_system
    verify_project_structure
    verify_docker_setup
    verify_api_functionality

    # Display summary
    display_summary
}

# Check if script is being executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
