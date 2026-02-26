# Task 01 — Add Redis to docker-compose and implement Redis client service

## Goal

Introduce Redis as the shared state layer for JWT blacklisting and OTP storage. Wire it into the server alongside the existing Postgres connection.

## Blocked by

None — can start immediately.

## Acceptance Criteria

- [ ] `docker-compose.yml` includes a `redis` service with a health check.
- [ ] `REDIS_HOST`, `REDIS_PORT`, and `REDIS_PASSWORD` are read from env vars (`.env`).
- [ ] `internal/cache/cache.go` defines a `Service` interface with `Get`, `Set`, `Delete`, `Health`, and `Close`.
- [ ] A concrete `go-redis`-backed implementation satisfies the interface.
- [ ] The `Server` struct in `internal/server/server.go` holds a `cache.Service` field.
- [ ] `GET /health` reflects Redis status alongside the existing Postgres status.

## Implementation Notes

### docker-compose addition

```yaml
redis:
  image: redis:7-alpine
  restart: unless-stopped
  ports:
    - "${REDIS_PORT}:6379"
  healthcheck:
    test: ["CMD", "redis-cli", "ping"]
    interval: 5s
    timeout: 3s
    retries: 3
  networks:
    - blueprint
```

### Service interface (`internal/cache/cache.go`)

```go
type Service interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Health() map[string]string
    Close() error
}
```

### Dependency

```
go get github.com/redis/go-redis/v9
```
