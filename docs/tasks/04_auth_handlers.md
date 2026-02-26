# Task 04 — Implement register and login HTTP handlers backed by the Forge

## Goal

Expose the first real auth endpoints, wiring the HTTP layer to Postgres (for user persistence) and the Forge (for safe password hashing/verification).

## Blocked by

- **#02** — users table must exist before inserting/querying users.
- **#03** — Forge dispatcher must exist to process hash/verify jobs.

## Acceptance Criteria

- [ ] `POST /auth/register` creates a new user and returns `201 Created`.
- [ ] `POST /auth/login` returns `200 OK` with a placeholder token field (JWT wired in task #05).
- [ ] Both handlers return `503` when the Forge queue is full.
- [ ] `POST /auth/register` returns `409 Conflict` if email already exists.
- [ ] `POST /auth/login` returns `401 Unauthorized` on wrong password.
- [ ] Input validation: email format, password minimum length (8 chars).
- [ ] Request/response structs and handlers live in `internal/server/handlers_auth.go`.
- [ ] Routes are registered under `/auth` group in `RegisterRoutes()`.
- [ ] Integration test covers register → login happy path using testcontainers (Postgres already set up in `database_test.go`).

## Implementation Notes

### Request / Response shapes

**Register**
```json
// POST /auth/register
// Request
{ "email": "user@example.com", "password": "supersecret" }

// Response 201
{ "id": "<uuid>", "email": "user@example.com" }
```

**Login**
```json
// POST /auth/login
// Request
{ "email": "user@example.com", "password": "supersecret" }

// Response 200
{ "token": "" }   // token filled in task #05
```

### Handler flow — Register

1. Decode and validate request body.
2. Submit `HashJob` to the Forge; block on `result` channel.
3. `INSERT INTO users (email, password_hash) VALUES ($1, $2)`.
4. Return user ID and email.

### Handler flow — Login

1. Decode and validate request body.
2. `SELECT id, password_hash FROM users WHERE email = $1`.
3. Submit `VerifyJob` to the Forge; block on `result` channel.
4. Return token (stubbed until task #05).

### database.Service extension

Add `CreateUser` and `GetUserByEmail` methods to the `database.Service` interface and implement them on the `service` struct.
