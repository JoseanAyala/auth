# Task 10 — Run race detector and fix any data races

## Goal

Verify the entire system is free of data races under concurrent load before considering the project production-ready.

## Blocked by

- **#03** — Forge must be complete.
- **#06** — Guardian must be complete.
- **#09** — Graceful shutdown must be complete (tests must be able to start/stop cleanly).

## Acceptance Criteria

- [ ] `go test -race ./...` passes with zero race conditions reported.
- [ ] A load test script (or `go test` benchmark) hits `POST /auth/register` and `POST /auth/login` with 50 concurrent goroutines for 10 seconds.
- [ ] The server is started and stopped under `-race` with no warnings.
- [ ] Any races found are documented with root cause and fix described in a comment or commit message.
- [ ] Known likely race candidates are reviewed and confirmed safe:
  - Token bucket state in the Guardian (`sync.Mutex` per bucket).
  - Job channel sends/receives in the Forge.
  - `sync.Map` usage in the Guardian sweeper.

## Implementation Notes

### Run the race detector

```bash
# Unit + integration tests
go test -race ./...

# Run server binary with race detector enabled
go run -race ./cmd/api
```

### Load test (using vegeta or hey)

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Hit the register endpoint with 50 concurrent workers for 10s
hey -n 10000 -c 50 -m POST \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  http://localhost:8080/auth/register
```

### Common race patterns to check

| Location | Risk | Mitigation |
|----------|------|------------|
| `guardian/ratelimiter.go` bucket | Concurrent token reads/writes | Per-bucket `sync.Mutex` |
| `forge/dispatcher.go` job channel | Concurrent submit + close | Channel semantics are safe; ensure `Stop()` only closes once |
| `metrics` goroutine sampler | Gauge write vs HTTP read | `promauto` gauges are thread-safe by design |
| `database` singleton | Double-init of `dbInstance` | Wrap in `sync.Once` |
