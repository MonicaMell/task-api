# task-api

[![CI](https://github.com/MonicaMell/task-api/actions/workflows/ci.yml/badge.svg)](https://github.com/MonicaMell/task-api/actions/workflows/ci.yml)

A production-style task management REST API in Go, built with a clean layered architecture, JWT authentication, PostgreSQL, full integration tests, and continuous deployment.

**Live demo:** https://task-api-production-bb11.up.railway.app

---

## Features

- **Layered architecture** — strict separation of handler, service, and repository concerns, wired together with dependency injection and interfaces at the boundaries.
- **JWT authentication** — register and log in to receive a signed token; protected routes require a valid `Bearer` token.
- **Secure password storage** — passwords are hashed with bcrypt and never stored or returned in plain text.
- **Owner-scoped data** — every task is bound to its owner at the database layer, so users can only ever access their own data.
- **Request validation** — declarative input validation with structured, field-level error responses.
- **PostgreSQL persistence** — accessed via the `pgx` driver with a connection pool, with versioned schema migrations.
- **Pagination** — list endpoints support `limit` and `offset`, clamped to safe bounds.
- **Graceful shutdown** — in-flight requests are allowed to finish on `SIGTERM`/`SIGINT`.
- **Integration tested** — the full stack is tested against a real PostgreSQL spun up in a throwaway container.
- **CI/CD** — every push runs build, vet, and the full test suite via GitHub Actions, and deploys to Railway.

## Architecture

```
HTTP request
     │
     ▼
┌─────────────┐
│   handler   │  HTTP concerns: decode request, validate, map errors to status codes, write JSON
└─────────────┘
     │ calls
     ▼
┌─────────────┐
│   service   │  business logic and rules; defines the repository interface it depends on
└─────────────┘
     │ calls
     ▼
┌─────────────┐
│ repository  │  data access only: SQL queries against PostgreSQL
└─────────────┘
     │
     ▼
 PostgreSQL
```

Each layer depends only on the layer directly below it, and the service depends on a repository *interface* rather than the concrete implementation. This keeps the layers decoupled and makes the business logic unit-testable without a database.

## Tech stack

| Concern         | Choice                                        |
| --------------- | --------------------------------------------- |
| Language        | Go (standard-library `net/http`, no framework)|
| Database        | PostgreSQL                                    |
| DB driver       | `jackc/pgx` with `pgxpool`                    |
| Auth            | `golang-jwt/jwt` (HS256), `bcrypt`            |
| Validation      | `go-playground/validator`                     |
| Migrations      | `golang-migrate`                              |
| Tests           | `testify`, `testcontainers-go`                |
| Containerization| Docker (multi-stage build)                    |
| CI              | GitHub Actions                                |
| Hosting         | Railway                                       |

## API reference

All request and response bodies are JSON. Protected routes require an `Authorization: Bearer <token>` header.

| Method | Path                | Auth | Description                          |
| ------ | ------------------- | ---- | ------------------------------------ |
| GET    | `/healthz`          | No   | Health check (also verifies the DB)  |
| POST   | `/auth/register`    | No   | Create an account                    |
| POST   | `/auth/login`       | No   | Log in, returns a JWT                 |
| GET    | `/tasks`            | Yes  | List your tasks (`?limit=&offset=`)  |
| POST   | `/tasks`            | Yes  | Create a task                        |
| GET    | `/tasks/{id}`       | Yes  | Get one of your tasks                |
| PUT    | `/tasks/{id}`       | Yes  | Replace one of your tasks            |
| DELETE | `/tasks/{id}`       | Yes  | Delete one of your tasks             |

### Example

```bash
# Register
curl -X POST https://task-api-production-bb11.up.railway.app/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"supersecret123"}'

# Log in -> { "token": "..." }
curl -X POST https://task-api-production-bb11.up.railway.app/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"supersecret123"}'

# Create a task
curl -X POST https://task-api-production-bb11.up.railway.app/tasks \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Buy milk","status":"todo"}'
```

Task `status` must be one of `todo`, `in_progress`, or `done`.

## Getting started (local)

### Prerequisites

- Go (see the version in `go.mod`)
- Docker and Docker Compose
- [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI

### Run it

```bash
# 1. Clone
git clone https://github.com/MonicaMell/task-api.git
cd task-api

# 2. Start PostgreSQL
docker compose up -d

# 3. Create your local env file
cat > .env <<EOF
PORT=8080
DATABASE_URL=postgres://taskuser:taskpass@localhost:5432/taskdb?sslmode=disable
JWT_SECRET=$(openssl rand -base64 32)
EOF

# 4. Apply migrations
export DATABASE_URL='postgres://taskuser:taskpass@localhost:5432/taskdb?sslmode=disable'
migrate -path migrations -database "$DATABASE_URL" up

# 5. Run the server
go run ./cmd/api
```

The API is now available at `http://localhost:8080`.

## Running the tests

The integration tests start a real PostgreSQL in a container, so Docker must be running.

```bash
go test ./...
```

This runs both the fast service unit tests and the full integration suite (register, login, task CRUD, ownership isolation, and input validation) against a real database.

## Project structure

```
task-api/
├── cmd/api/                # entry point, HTTP handlers, routing, middleware
├── internal/
│   ├── auth/               # JWT token generation and verification
│   ├── config/             # environment configuration loading
│   ├── handler/            # (HTTP layer lives in cmd/api)
│   ├── model/              # domain types (User, Task)
│   ├── repository/         # PostgreSQL data access
│   └── service/            # business logic and repository interfaces
├── migrations/             # versioned SQL schema migrations
├── .github/workflows/      # CI pipeline
├── Dockerfile              # multi-stage production build
└── docker-compose.yml      # local PostgreSQL
```

## Deployment

The service is containerized with a multi-stage Dockerfile and deployed to Railway, which builds the image on every push to `main`. Configuration (`DATABASE_URL`, `JWT_SECRET`, `PORT`) is provided through environment variables; no secrets live in the repository.
