# Task 03 — Build the Forge: worker pool with Argon2 hashing

## Goal

Prevent CPU exhaustion by processing all Argon2 hash and verify operations through a fixed-size worker pool. This is the heart of the system's concurrency model.

## Blocked by

None — can start immediately.

## Acceptance Criteria

- [ ] `internal/forge/` package exists with `dispatcher.go` and `worker.go`.
- [ ] `HashJob` carries a plain-text password and a `result chan HashResult`.
- [ ] `HashResult` carries the resulting hash string and an error.
- [ ] `VerifyJob` carries a plain-text password, a stored hash, and a `result chan VerifyResult`.
- [ ] `Dispatcher` exposes `Start()`, `Stop()`, `Submit(job Job) error`.
- [ ] Worker count is `runtime.NumCPU()`.
- [ ] The internal job channel is buffered (size = `2 * runtime.NumCPU()`).
- [ ] `Submit` returns `ErrQueueFull` immediately if the channel is full (non-blocking send).
- [ ] `Stop()` closes the job channel and waits for all workers to finish via `sync.WaitGroup`.
- [ ] Unit tests cover: successful hash, successful verify, wrong password verify, and `ErrQueueFull` backpressure.
- [ ] Tests pass under `go test -race ./internal/forge/...`.

## Implementation Notes

### Package structure

```
internal/forge/
  dispatcher.go   — Dispatcher struct, Start/Stop/Submit
  worker.go       — worker goroutine, Argon2 params
  forge_test.go
```

### Argon2 parameters (worker.go)

```go
const (
    argonTime    = 1
    argonMemory  = 64 * 1024 // 64 MB
    argonThreads = 4
    argonKeyLen  = 32
)
```

Use `argon2.IDKey` from `golang.org/x/crypto/argon2` (already in `go.mod`).

Store the hash as: `$argon2id$v=19$m=65536,t=1,p=4$<base64-salt>$<base64-hash>`

### Backpressure

```go
func (d *Dispatcher) Submit(job Job) error {
    select {
    case d.jobs <- job:
        return nil
    default:
        return ErrQueueFull
    }
}
```

### Caller's 503 response

When `Submit` returns `ErrQueueFull`, the HTTP handler responds:

```json
{ "error": "server busy, try again later" }
```

with status `503 Service Unavailable`.
