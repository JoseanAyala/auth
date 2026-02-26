# Task 02 — Set up database migrations tooling and initial schema

## Goal

Establish a repeatable migration workflow using `goose` and define the initial `users` table that the auth handlers will depend on.

## Blocked by

None — can start immediately.

## Acceptance Criteria

- [ ] `github.com/pressly/goose/v3` is added to `go.mod`.
- [ ] `db/migrations/` directory exists with a `00001_create_users.sql` migration.
- [ ] `users` table has: `id` (UUID, PK), `email` (unique, not null), `password_hash` (text, not null), `created_at`, `updated_at`.
- [ ] `cmd/migrate/main.go` runs `goose up` / `goose down` based on a CLI argument.
- [ ] `Makefile` has `migrate-up` and `migrate-down` targets.
- [ ] Migration runs cleanly against the Docker Compose Postgres instance.

## Implementation Notes

### Directory layout

```
db/
  migrations/
    00001_create_users.sql
cmd/
  migrate/
    main.go
```

### Migration SQL

```sql
-- +goose Up
CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email        TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE users;
```

### Makefile targets

```makefile
migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down
```

### Dependency

```
go get github.com/pressly/goose/v3
```
