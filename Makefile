.PHONY: help build build-release version run test clean setup migrate migrate-up migrate-down migrate-status generate generate-api generate-db dev docker-up docker-down

# Default target
help:
	@echo "Available targets:"
	@echo "  help          - Show this help message"
	@echo "  setup         - Set up development environment"
	@echo "  build         - Build the application"
	@echo "  build-release - Build the application with version info"
	@echo "  version       - Show version information"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  test-api      - Run API integration tests"
	@echo "  test-forms    - Run form integration tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  generate      - Generate all code (API + DB)"
	@echo "  generate-api  - Generate API code from OpenAPI spec"
	@echo "  generate-db   - Generate database code from SQL queries"
	@echo "  generate-templ- Generate Go code from templ templates"
	@echo "  migrate       - Run database migrations"
	@echo "  migrate-up    - Run database migrations (up)"
	@echo "  migrate-down  - Rollback last migration"
	@echo "  migrate-status- Show migration status"
	@echo "  dev           - Start development environment"
	@echo "  docker-up     - Start PostgreSQL with Docker"
	@echo "  docker-down   - Stop PostgreSQL Docker container"

# Go variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Version variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X pepo/internal/version.Version=$(VERSION) -X pepo/internal/version.Commit=$(COMMIT) -X pepo/internal/version.Date=$(DATE)"

# Build the application
build:
	@echo "Building application..."
	go build -o $(GOBIN)/pepo ./cmd/server

# Build the application with version information
build-release:
	@echo "Building application with version info..."
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"
	go build $(LDFLAGS) -o $(GOBIN)/pepo ./cmd/server

# Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

# Run the application
run: build
	@echo "Starting application..."
	$(GOBIN)/pepo

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run API integration tests
test-api: build
	@echo "Running API integration tests..."
	./test_api.sh

# Run form integration tests
test-forms: build
	@echo "Running form integration tests..."
	./test_forms.sh

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(GOBIN)
	go clean
	go mod tidy

# Set up development environment
setup:
	@echo "Setting up development environment..."
	go mod download
	go install github.com/amacneil/dbmate/v2@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/ogen-go/ogen/cmd/ogen@latest
	go install github.com/a-h/templ/cmd/templ@latest
	@echo "Development environment setup complete!"

# Generate all code
generate: generate-api generate-db generate-templ

# Generate API code from OpenAPI specification
generate-api:
	@echo "Generating API code..."
	mkdir -p internal/api
	~/go/bin/ogen --target internal/api --package api api/openapi.yaml

# Generate database code from SQL queries
generate-db:
	@echo "Generating database code..."
	mkdir -p internal/db
	~/go/bin/sqlc generate

# Generate template code from templ files
generate-templ:
	@echo "Generating template code..."
	~/go/bin/templ generate

# Database migration commands
migrate: migrate-up

migrate-up:
	@echo "Running database migrations..."
	~/go/bin/dbmate up

migrate-down:
	@echo "Rolling back last migration..."
	~/go/bin/dbmate down

migrate-status:
	@echo "Checking migration status..."
	~/go/bin/dbmate status

# Create new migration
migrate-new:
	@read -p "Enter migration name: " name; \
	~/go/bin/dbmate new $$name

# Start development environment
dev: docker-up migrate generate
	@echo "Development environment ready!"

# Docker commands for PostgreSQL
docker-up:
	@echo "Starting PostgreSQL container..."
	docker run --name pepo-postgres -d \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=password \
		-e POSTGRES_DB=pepo_dev \
		-p 5433:5432 \
		postgres:15-alpine || echo "Container already running"

docker-down:
	@echo "Stopping PostgreSQL container..."
	docker stop pepo-postgres || true
	docker rm pepo-postgres || true

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod verify

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Security check
security:
	@echo "Running security checks..."
	gosec ./...

# Update dependencies
update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Create database
createdb:
	@echo "Creating database..."
	~/go/bin/dbmate create

# Drop database
dropdb:
	@echo "Dropping database..."
	~/go/bin/dbmate drop

# Reset database (drop, create, migrate)
resetdb: dropdb createdb migrate
	@echo "Database reset complete!"
