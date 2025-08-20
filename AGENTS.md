# AGENTS Instructions

This repository contains a Go-based web application for tracking performance of direct reports. Key technologies and workflows:

- **Language & Frameworks:** Go for server-side code; frontend leverages HTMX and Tailwind CSS.
- **API:** RESTful API described with OpenAPI, code generated via [ogen](https://ogen.dev/).
- **Database:** PostgreSQL with interactions generated through [sqlc](https://sqlc.dev/); schema managed by [dbmate](https://github.com/amacneil/dbmate). Primary keys use [xid](https://github.com/rs/xid).

## Development Guidelines

- Format Go code with `go fmt ./...` before committing.
- Run tests with `go test ./...` (or `make test`) to ensure all tests pass.

