# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build       # Compile binary to ./main
make run         # Run the app (go run cmd/api/main.go)
make watch       # Live reload via air (auto-installs if missing)
make test        # Run all tests with verbose output
make itest       # Run database integration tests only (requires Docker)
make docker-run  # Start full stack (app + postgres + redis) via docker-compose
make docker-down # Stop docker-compose services
make clean       # Remove compiled binary
```

Run a single test:
```bash
go test ./internal/server/... -run TestHandlerName -v
```

## Architecture

This is a Go authentication microservice with three concurrency subsystems planned across phased tasks (see `docs/tasks/`):

1. **The Guardian** — Rate limiter middleware using `sync.Map` + token bucket algorithm, with a background sweeper goroutine
2. **The Hasher** — Worker pool for CPU-heavy Argon2 password hashing (`runtime.NumCPU()` workers, backpressure → 503)
3. **The Courier** — Fire-and-forget async notification system (email/SMS) using `context.WithTimeout`

### Current foundation (`internal/`)

- **`server/`** — HTTP server using `chi` router with CORS and request logging. Routes live in `routes.go`. The `Server` struct holds DB and cache dependencies injected at startup.
- **`database/`** — PostgreSQL via `pgx/v5`. Singleton `dbInstance`. `Service` interface enables testing. Integration tests use `testcontainers-go` to spin up real Postgres.
- **`cache/`** — Redis via `go-redis/v9`. Singleton `cacheInstance`. `Service` interface mirrors database pattern.

### Entry point

`cmd/api/main.go` wires everything together: loads `.env`, creates server, starts listening, handles SIGINT/SIGTERM with a 5-second graceful shutdown.

### Health check

`GET /health` returns combined DB + Redis stats (connection pool metrics, Redis INFO).

## Key dependencies

| Package | Purpose |
|---|---|
| `go-chi/chi/v5` | HTTP router |
| `go-chi/cors` | CORS middleware |
| `jackc/pgx/v5` | PostgreSQL driver |
| `redis/go-redis/v9` | Redis client |
| `joho/godotenv` | `.env` loading |
| `testcontainers/testcontainers-go` | Integration test containers |

## Environment

Copy `.env` for local development. Docker Compose uses service names as hostnames (`psql_bp`, `redis`). For local non-Docker runs, change `DB_HOST=localhost` and `REDIS_HOST=localhost`.

## Phased task roadmap

Implementation is tracked in `docs/tasks/` (00–10). Completed: foundation (server, DB, Redis, graceful shutdown). Upcoming: DB migrations (02), Hasher worker pool (03), auth handlers (04), JWT blacklisting (05), Guardian rate limiter (06), Courier notifications (07), Prometheus metrics (08), race detector tests (10).
