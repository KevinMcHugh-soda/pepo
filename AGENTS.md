# AGENTS Instructions

This repository contains a Go-based web application for tracking performance of direct reports. Key technologies and workflows:

- **Language & Frameworks:** Go for server-side code; frontend leverages HTMX and Tailwind CSS.
- **API:** RESTful API described with OpenAPI. The spec lives at `api/openapi.yaml` and code is generated via [ogen](https://ogen.dev/) into `internal/api`. Run `make generate-api` after modifying the spec to refresh the generated code.
- **Database:** PostgreSQL with interactions generated through [sqlc](https://sqlc.dev/); schema managed by [dbmate](https://github.com/amacneil/dbmate). Primary keys use [xid](https://github.com/rs/xid).

## Data Flow from Forms to Database

- HTMX forms in `templates/` submit data with `hx-post` or `hx-put` to API routes.
- `FormToJSONAdapter` middleware (`internal/middleware/form_adapter.go`) intercepts form submissions and converts `application/x-www-form-urlencoded` or multipart form data into JSON.
- API handlers (`internal/handlers`) unmarshal this JSON into request structs and call the appropriate sqlc-generated query functions.
- These query functions (`internal/db`) execute SQL against the PostgreSQL database, persisting or retrieving the submitted data.

## Development Guidelines

- Format Go code with `go fmt ./...` before committing.
- Run tests with `go test ./...` (or `make test`) to ensure all tests pass.
- After updating `api/openapi.yaml`, run `make generate-api` to regenerate API code.
- Table names are singular, eg action, person.
- API routes are plural, eg GET /actions, GET /people
- Name migrations using the current timestamp, e.g. `20240115093000_add_user_table.sql`.
- Use named SQL parameters (e.g., `sqlc.arg(name)`) instead of positional placeholders like `$1`.
