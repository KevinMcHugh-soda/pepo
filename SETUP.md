# Setup Instructions for Pepo Performance Tracking

This document provides step-by-step instructions to set up the Pepo performance tracking application for development.

## Prerequisites

Before you begin, ensure you have the following installed on your system:

- **Go 1.21+**: [Download and install Go](https://golang.org/dl/)
- **PostgreSQL 15+**: [Download PostgreSQL](https://www.postgresql.org/download/) or use Docker
- **Docker & Docker Compose** (optional but recommended): [Install Docker](https://docs.docker.com/get-docker/)
- **Make**: Usually pre-installed on Linux/macOS, for Windows use [chocolatey](https://chocolatey.org/) or [scoop](https://scoop.sh/)

## Quick Start

### 1. Clone and Navigate to Project

```bash
cd pepo
```

### 2. Install Development Tools

Run the setup command to install required development tools:

```bash
make setup
```

This will install:
- `dbmate` for database migrations
- `sqlc` for generating database code
- `ogen` for generating API code from OpenAPI specs

### 3. Start PostgreSQL Database

**Option A: Using Docker (Recommended)**
```bash
make docker-up
```

**Option B: Using Docker Compose**
```bash
docker-compose up -d postgres
```

**Option C: Local PostgreSQL Installation**
If you have PostgreSQL installed locally, create a database:
```bash
createdb pepo_dev
```

### 4. Configure Environment

Copy the example environment file and adjust if needed:
```bash
cp .env .env.local
```

Edit `.env.local` if your database configuration differs from the defaults.

### 5. Run Database Migrations

```bash
make migrate
```

### 6. Generate Code

Generate the API and database code:
```bash
make generate
```

### 7. Build and Run

```bash
make run
```

The application will be available at: http://localhost:8080

## Development Workflow

### Starting Development Environment

For a complete development setup in one command:
```bash
make dev
```

This will:
- Start PostgreSQL container
- Run database migrations
- Generate all code
- Display readiness message

### Common Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make build` | Build the application |
| `make run` | Build and run the application |
| `make test` | Run tests |
| `make clean` | Clean build artifacts |
| `make generate` | Generate API and database code |
| `make migrate` | Run database migrations |
| `make migrate-status` | Check migration status |
| `make migrate-new` | Create a new migration |
| `make docker-up` | Start PostgreSQL container |
| `make docker-down` | Stop PostgreSQL container |

### Database Management

**Create a new migration:**
```bash
make migrate-new
# Enter migration name when prompted
```

**Check migration status:**
```bash
make migrate-status
```

**Rollback last migration:**
```bash
make migrate-down
```

**Reset database (drop, create, migrate):**
```bash
make resetdb
```

### Code Generation

**Regenerate API code after OpenAPI changes:**
```bash
make generate-api
```

**Regenerate database code after SQL changes:**
```bash
make generate-db
```

## Project Structure

```
pepo/
├── api/                    # OpenAPI specifications
│   └── openapi.yaml
├── cmd/                    # Application entry points
│   └── server/
│       └── main.go
├── db/                     # Database related files
│   ├── migrations/         # Database migrations
│   └── queries/            # SQL queries for sqlc
├── internal/               # Generated and internal code
│   ├── api/               # Generated API code (ogen)
│   └── db/                # Generated database code (sqlc)
├── static/                 # Static files (CSS, JS, images)
├── templates/              # HTML templates
├── docker-compose.yml      # Docker development environment
├── Makefile               # Development commands
├── sqlc.yaml              # sqlc configuration
└── go.mod                 # Go module file
```

## Database Access

**Using pgAdmin (Web Interface):**
- URL: http://localhost:5050
- Email: admin@pepo.local
- Password: admin

**Using psql (Command Line):**
```bash
psql postgres://postgres:password@localhost:5432/pepo_dev
```

## API Documentation

The API is defined in `api/openapi.yaml` and follows OpenAPI 3.0 specification.

**Key Endpoints:**
- `GET /health` - Health check
- `GET /api/v1/persons` - List people
- `POST /api/v1/persons` - Create person
- `GET /api/v1/persons/{id}` - Get person by ID
- `PUT /api/v1/persons/{id}` - Update person
- `DELETE /api/v1/persons/{id}` - Delete person

## Testing

Run tests with:
```bash
make test
```

## Troubleshooting

### Database Connection Issues

1. **Check if PostgreSQL is running:**
   ```bash
   docker ps
   ```

2. **Check database logs:**
   ```bash
   docker logs pepo-postgres
   ```

3. **Restart PostgreSQL:**
   ```bash
   make docker-down
   make docker-up
   ```

### Port Already in Use

If port 8080 is already in use, set a different port:
```bash
export PORT=8081
make run
```

### Permission Issues on macOS/Linux

If you encounter permission issues:
```bash
sudo make setup
```

### Clean Start

For a completely clean start:
```bash
make clean
make docker-down
make docker-up
make migrate
make generate
make run
```

## Production Deployment

For production deployment considerations:
- Set `ENV=production` in environment variables
- Use proper PostgreSQL connection with SSL
- Configure proper logging
- Set up health monitoring
- Use proper secrets management for database credentials

## Contributing

1. Make changes to code
2. Run `make generate` if you modified OpenAPI spec or SQL queries
3. Run `make test` to ensure tests pass
4. Build and test locally with `make run`
5. Submit your changes

## Need Help?

- Check the main [README.MD](README.MD) for project overview
- Review the [Makefile](Makefile) for available commands
- Check logs with `docker logs pepo-postgres` for database issues
- Verify your Go installation: `go version`
