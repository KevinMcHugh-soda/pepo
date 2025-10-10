This is a project to build a simple web application that supports tracking performance of direct reports.

The project is implemented in Golang.

It uses a RESTful API to communicate with the frontend. The API is defined using OpenAPI. We generate code for the API using https://ogen.dev/

We store the data in a PostgreSQL database. We generate code to interact with the database using https://sqlc.dev/

The database schema is managed using https://github.com/amacneil/dbmate

The core database model is a Person, who has a name. The primary key is an xid: https://github.com/rs/xid

The front-end will be implemented using HTMX: https://htmx.org/ and Tailwind CSS: https://tailwindcss.com/

## Talking to the Pepo Agent

You can start an interactive terminal session with the Pepo agent by building the CLI located in `cmd/agent`:

```
go build -o pepo-agent ./cmd/agent
```

Before running the binary, export your OpenAI API key:

```
export OPENAI_API_KEY=sk-...
```

The agent can then be launched with:

```
./pepo-agent
```

By default the tool will talk to a Pepo API server running at `http://localhost:8080`. Override this and other options with flags (for example `-api-base-url`) or the environment variables `PEPO_API_BASE_URL`, `PEPO_AGENT_SYSTEM`, `PEPO_AGENT_MAX_ROUNDS`, and `PEPO_AGENT_TIMEOUT`.
