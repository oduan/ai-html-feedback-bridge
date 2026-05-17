# AI HTML Feedback Bridge

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](docker-compose.yml)

A lightweight HTTP service that stores user feedback, notes, or confirmations submitted from AI-generated HTML pages.

It does one thing: **save raw content by `interaction_id` and return it as-is on request**. That's it.

Every AI HTML page you generate can now collect user input and persist it — no database setup, no frontend framework, no auth boilerplate.

## Features

- **Save** user content via `POST /interactions/{interaction_id}`
- **Read** saved content via `GET /interactions/{interaction_id}`
- **Overwrite** on repeat submissions (same `interaction_id` → updates content, keeps `created_at`)
- **Content-Type preservation** — saved headers are returned on read
- **SQLite persistence** — zero external dependencies, just a single file
- **Configurable size limit** via `MAX_CONTENT_SIZE` (default 1 MB)
- **CORS enabled** by default — browser HTML pages can call the API directly
- **Docker Compose** or bare Go — choose your workflow

## Quick Start

### Docker

```bash
docker compose up --build
```

The service starts at `http://localhost:8080`. Data persists to `./data/app.db`.

### Go

```bash
cd backend
go run ./cmd/server
```

PowerShell:

```powershell
$env:PORT="8080"
$env:SQLITE_PATH="./data/app/data.db"
$env:MAX_CONTENT_SIZE="1048576"
go run ./cmd/server
```

Bash:

```bash
PORT=8080 SQLITE_PATH=./data/app/data.db MAX_CONTENT_SIZE=1048576 go run ./cmd/server
```

## Use Cases

- AI generates a requirements confirmation page → user fills in details → saved via this service
- AI generates a review page → user adds revision notes, checkboxes, or comments → persisted
- AI outputs a structured form → user submits JSON, Markdown, or plain text → retrievable later
- Local demos, internal tools, or controlled environments where AI-generated pages need simple writeback

The service stores raw content. It does **not** execute, render, validate, or transform user submissions.

## API Reference

Base URL: `http://localhost:8080`

### Save Content

```http
POST /interactions/{interaction_id}
```

Body can be any text — JSON, HTML, Markdown, plain text, etc. The service saves it as-is.

```bash
curl -i \
  -X POST \
  -H "Content-Type: application/json" \
  --data '{"feedback":"Need to add project timeline","approved":true}' \
  http://localhost:8080/interactions/demo-001
```

**Success response:**

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"ok":true,"interaction_id":"demo-001"}
```

If no `Content-Type` header is provided, the service defaults to:

```
text/plain; charset=utf-8
```

### Read Content

```http
GET /interactions/{interaction_id}
```

```bash
curl -i http://localhost:8080/interactions/demo-001
```

**Success response:**

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"feedback":"Need to add project timeline","approved":true}
```

The response body is the raw saved content. The `Content-Type` header matches what was sent during save.

### Error Responses

All errors return JSON:

```json
{
  "ok": false,
  "error": "interaction not found"
}
```

| Status Code | Description |
|---|---|
| `200 OK` | Save or read succeeded |
| `400 Bad Request` | `interaction_id` is empty or invalid |
| `404 Not Found` | `interaction_id` does not exist |
| `413 Payload Too Large` | Request body exceeds `MAX_CONTENT_SIZE` |
| `500 Internal Server Error` | Server-side failure |

## Configuration

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `SQLITE_PATH` | `./data/app.db` | Path to the SQLite database file |
| `MAX_CONTENT_SIZE` | `1048576` | Max request body size in bytes (1 MB) |

If `MAX_CONTENT_SIZE` is not a valid number, the service exits with an error on startup.

## Interaction IDs

An `interaction_id` identifies a single feedback target — one generated page, one confirmation step, or one review task.

**Constraints:**

- Length: 1–128 characters
- Allowed: letters, digits, underscores (`_`), hyphens (`-`), dots (`.`)

**Examples:**

```
demo-001
task-2026-001.r1.html1
review.ab12cd
```

## CORS

Cross-origin requests are enabled by default so browser-based HTML pages can call the API directly.

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

The service does **not** use cookies or credentials, and does **not** return `Access-Control-Allow-Credentials: true`.

## Data Storage

Content is stored in a SQLite table called `interactions`:

| Column | Type | Description |
|---|---|---|
| `interaction_id` | `TEXT PRIMARY KEY` | Unique feedback target ID |
| `content_type` | `TEXT NOT NULL` | `Content-Type` from the save request |
| `content` | `TEXT NOT NULL` | Raw user-submitted content |
| `created_at` | `TEXT NOT NULL` | First save time (UTC / RFC3339) |
| `updated_at` | `TEXT NOT NULL` | Last update time (UTC / RFC3339) |

Re-saving the same `interaction_id`:

- Overwrites `content` and `content_type`
- Preserves `created_at`
- Updates `updated_at`

## Project Structure

```
.
├── backend/
│   ├── cmd/server/                 # Service entrypoint
│   ├── internal/config/            # Environment variable configuration
│   ├── internal/httpapi/           # Routing, handlers, CORS, error responses
│   ├── internal/storage/           # SQLite connection, migrations, CRUD
│   ├── migrations/                 # Database schema files
│   ├── go.mod
│   └── go.sum
├── examples/requests/              # Ready-to-use curl examples
├── Dockerfile
├── docker-compose.yml
├── SPEC.md                         # Full specification
├── PLAN.md                         # Development plan
└── README.md                       # This file
```

## Deployment Notes

This service is designed for **local, intranet, demo, or controlled environments**.

For public deployment, consider adding:

- Access control (tokens, API keys, or login)
- Rate limiting
- Audit logging
- Data retention policies
- HTTPS / TLS termination

The service does **not** include user management, permissions, version history, an admin UI, or content review. It is not an HTML page generator — HTML pages are created by AI, frontend code, or other tools.

## Development

```bash
# Run tests
cd backend
go test ./...

# Format code
cd backend
gofmt -w .

# Static analysis
cd backend
go vet ./...
```

Before committing:

```bash
cd backend
gofmt -w .
go test ./...
go vet ./...
```

## License

This project does not currently include a `LICENSE` file. Please add one before publishing as open source.
