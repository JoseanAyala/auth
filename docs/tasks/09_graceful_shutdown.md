# Task 09 — Extend graceful shutdown to drain the Hasher and stop the Guardian sweeper

## Goal

Ensure no goroutines, in-flight hashing jobs, or background sweepers are abandoned when the process receives SIGINT/SIGTERM. Everything must clean up within the existing 5-second window.

## Blocked by

- **#03** — Hasher dispatcher must expose `Stop()`.
- **#06** — Guardian rate limiter must expose `Stop()`.

## Acceptance Criteria

- [ ] `gracefulShutdown` in `cmd/api/main.go` shuts down in this order:
  1. Stop accepting new HTTP requests (`server.Shutdown`).
  2. Call `hasher.Stop()` — blocks until all in-flight jobs complete.
  3. Call `guardian.Stop()` — stops the background sweeper.
  4. Close the Redis connection (`cache.Close()`).
  5. Close the DB connection (`db.Close()`).
- [ ] The entire sequence runs within a `context.WithTimeout` of 5 seconds.
- [ ] If any step exceeds the timeout, the process logs the step that timed out and exits cleanly.
- [ ] A manual test (run server, fire a slow request, kill with Ctrl+C) confirms in-flight hashing jobs complete before exit.

## Implementation Notes

### Revised shutdown sequence

```go
func gracefulShutdown(deps *deps, done chan bool) {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()
    <-ctx.Done()

    log.Println("shutting down gracefully...")

    shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := deps.httpServer.Shutdown(shutCtx); err != nil {
            log.Printf("HTTP shutdown error: %v", err)
        }
    }()
    wg.Wait()

    deps.hasher.Stop()
    deps.guardian.Stop()
    deps.cache.Close()
    deps.db.Close()

    log.Println("shutdown complete")
    done <- true
}
```

### deps struct

Introduce a `deps` struct in `cmd/api/main.go` to group all long-lived resources passed to the shutdown function, keeping `main()` clean.

```go
type deps struct {
    httpServer *http.Server
    hasher   *hasher.Dispatcher
    guardian   *guardian.RateLimiter
    cache      cache.Service
    db         database.Service
}
```
