# goHTTPServer

A minimal, production-minded Go HTTP server that serves a small static webapp and exposes a tiny JSON API for validating short "chirps" (max 140 characters). It's designed for simplicity, observability and easy local development.

Features
- Static file serving under `/app/` and `/app/assets/`
- Health check endpoint at `/api/healthz`
- JSON API to validate chirp payloads at `/api/validate_chirp`
- Simple admin endpoints for metrics and counter reset

Table of Contents
- Quick Start
- Endpoints
- Development
- Contributing
- License

Quick Start

Prerequisites
- Go 1.18+ installed

Build & Run
```bash
# build
go build -o goHTTPServer .

# run (binary)
./goHTTPServer

# or run directly with the Go tool
go run .
```

The server listens on port `8080` by default and serves the static app directory at `/app/`.

Endpoints

- Health check
	- `GET /api/healthz`
	- Response: `200 OK` with text body `OK`
	- Example:
		```bash
		curl -i http://localhost:8080/api/healthz
		```

- Validate chirp
	- `POST /api/validate_chirp`
	- Request JSON: `{"body":"your chirp text"}`
	- Success: `200 OK` with JSON `{ "valid": true }`
	- Failure: `400 Bad Request` with JSON `{ "error": "..." }` (currently used when chirp > 140 chars or on bad JSON)
	- Example:
		```bash
		curl -sS -X POST http://localhost:8080/api/validate_chirp \
			-H 'Content-Type: application/json' \
			-d '{"body":"Hello, world!"}'
		```

- Static assets
	- `GET /app/` serves files from the project root (useful for single-file app assets)
	- `GET /app/assets/` serves files from `app/assets`

- Admin
	- `GET /admin/metrics` — simple HTML page showing a visit counter for the file server
	- `POST /admin/reset` — reset the file server visit counter to `0`
	- Example (reset):
		```bash
		curl -X POST http://localhost:8080/admin/reset
		```

Development

- The project is intentionally minimal and dependency-free outside the Go standard library.
- Build and run using the commands in Quick Start.
- If you want to extend the server, follow Go best practices: add unit tests, small interfaces around handlers for easier mocking, and document new endpoints in this README.

Contributing

- Contributions are welcome. Please open an issue or a pull request with a clear description and rationale for changes.

License

This repository is provided under the MIT License. See the LICENSE file for details, or contact the maintainer.

Maintainer
- Project: `goHTTPServer`
- Location: local development

Enjoy — and happy hacking!