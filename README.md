# goHTTPServer

A minimal Go HTTP backend with static file serving, user authentication, refresh-token support, webhook handling, chirp storage, and admin observability.

This repository includes:
- Static assets served under `/app/` and `/app/assets/`
- Health check endpoint at `/api/healthz`
- User creation, login, and update APIs
- JWT access tokens and refresh tokens
- Chirp creation, retrieval, and deletion
- External webhook handling for user upgrades
- Admin metrics and local reset behavior
- PostgreSQL schema and SQLC-generated database access

## Requirements

- Go 1.26+
- PostgreSQL database available via `DB_URL`
- Environment variables available in `.env` or shell environment

## Setup

1. Create a `.env` file or export these variables:
   - `DB_URL` — PostgreSQL connection string
   - `SECRET` — JWT signing secret
   - `POLKA_KEY` — API key for webhook authorization
   - `PLATFORM` — set to `dev` to enable `/admin/reset`

2. Build and run the server:
```bash
go build -o goHTTPServer .
./goHTTPServer
```

Or run directly:
```bash
go run .
```

The server listens on port `8080`.

## Environment variables

- `DB_URL` — PostgreSQL DSN for the backend database
- `SECRET` — shared secret for signing JWT access tokens
- `POLKA_KEY` — API key used by `/api/polka/webhooks`
- `PLATFORM` — if set to `dev`, local reset is allowed via `/admin/reset`

## Routes and URL patterns

### Static content

- `GET /app/`
  - Serves static assets from the repository root.
- `GET /app/assets/`
  - Serves static assets from `app/assets`.

### Health

- `GET /api/healthz`
  - Returns `200 OK` with plain text `OK`.

### User management

- `POST /api/users`
  - Create a new user.
  - Request body: `{"email":"...","password":"..."}`
  - Response: `201 Created`
  - Returns user metadata.

- `POST /api/login`
  - Authenticate a user.
  - Request body: `{"email":"...","password":"..."}`
  - Response: `200 OK`
  - Returns:
    - `token` — JWT bearer access token
    - `refresh_token` — refresh token to renew access
    - `is_chirpy_red` — upgrade status

- `PUT /api/users`
  - Update the authenticated user's email and password.
  - Requires `Authorization: Bearer <access_token>`.
  - Request body: `{"email":"...","password":"..."}`
  - Response: `200 OK`

### Refresh token management

- `POST /api/refresh`
  - Exchange a refresh token for a new access token.
  - Requires `Authorization: Bearer <refresh_token>`.
  - Response: `200 OK`
  - Returns `token`.

- `POST /api/revoke`
  - Revoke an existing refresh token.
  - Requires `Authorization: Bearer <refresh_token>`.
  - Response: `204 No Content`.

### Chirp API

- `POST /api/chirps`
  - Create a new chirp.
  - Requires `Authorization: Bearer <access_token>`.
  - Request body: `{"body":"your chirp text"}`
  - Chirps are limited to 140 characters.
  - Profanity filter censors configured banned words.
  - Response: `200 OK`

- `GET /api/chirps`
  - List chirps.
  - Optional query parameter: `author_id=<uuid>` to filter by author.
  - Response: `200 OK`

- `GET /api/chirps/{chirpID}`
  - Retrieve a chirp by its UUID.
  - Response: `200 OK` or `404 Not Found`.

- `DELETE /api/chirps/{chirpID}`
  - Delete a chirp by its UUID.
  - Requires `Authorization: Bearer <access_token>`.
  - Only the chirp owner may delete the chirp.
  - Response: `204 No Content`.

### Webhook handling

- `POST /api/polka/webhooks`
  - Receives external webhook events.
  - Requires `Authorization: Apikey <POLKA_KEY>`.
  - Supported event: `user.upgraded`.
  - When received, upgrades the user to `is_chirpy_red`.

### Admin endpoints

- `GET /admin/metrics`
  - Returns an HTML page showing the file server visit count.

- `POST /admin/reset`
  - Resets the user table and file server hit counter.
  - Only allowed when `PLATFORM=dev`.
  - Response: `200 OK`.

## Authentication and security

- Access tokens are JWTs signed with the `SECRET` environment variable.
- JWTs are passed with `Authorization: Bearer <token>`.
- Refresh tokens are random hex strings stored in the database.
- Webhook requests use `Authorization: Apikey <POLKA_KEY>`.
- Passwords are hashed with Argon2id before persisting.

## Database schema

The backend uses a PostgreSQL database and SQLC-generated queries.

### `users` table
- `id` UUID primary key
- `created_at` timestamp
- `updated_at` timestamp
- `email` unique
- `hashed_password`
- `is_chirpy_red` boolean

### `chirps` table
- `id` UUID primary key
- `created_at` timestamp
- `updated_at` timestamp
- `body` text
- `user_id` UUID foreign key

### `refresh_tokens` table
- `token` string primary key
- `created_at` timestamp
- `updated_at` timestamp
- `user_id` UUID foreign key
- `expires_at` timestamp
- `revoked_at` nullable timestamp

## SQLC-generated queries

- `CreateUser`
- `GetUserByEmail`
- `UpdateUserEmailAndPassword`
- `DeleteUsers`
- `UpgradeUserToChirpyRed`
- `CreateChirps`
- `GetAllChirps`
- `GetChirpsForAuthor`
- `RetriveOneChirp`
- `DeleteChirpById`
- `CreateRefreshToken`
- `GetUserFromRefreshToken`
- `RevokeRefreshToken`

## Notes

- `POST /api/users` creates a user but does not issue tokens.
- Login via `POST /api/login` is required to receive access and refresh tokens.
- `POST /admin/reset` is safe only in development when `PLATFORM=dev`.
- The backend currently censors the words `kerfuffle`, `sharbert`, and `fornax` in chirp content.

## Dependencies

- `github.com/alexedwards/argon2id`
- `github.com/golang-jwt/jwt/v4`
- `github.com/google/uuid`
- `github.com/joho/godotenv`
- `github.com/lib/pq`

---

Enjoy working with the `goHTTPServer` backend! Keep this documentation synced with code changes.
