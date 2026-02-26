# Task 06 — Build the Guardian: rate limiter middleware with token bucket and background sweeper

## Goal

Protect the service from brute-force and flood attacks by enforcing per-IP rate limits using a token bucket algorithm with a background goroutine to clean up stale state.

## Blocked by

None — can start immediately.

## Acceptance Criteria

- [ ] implemented the token bucket per IP using `sync.Map`.
- [ ] Each bucket refills at `RATE_LIMIT_RPS` tokens/second up to a max of `RATE_LIMIT_BURST`.
- [ ] Requests that exceed the limit receive `429 Too Many Requests` with a `Retry-After` header.
- [ ] A background sweeper goroutine removes entries not seen in the last 5 minutes; runs every 60 seconds.
- [ ] `RateLimiter` exposes `Middleware() func(http.Handler) http.Handler`, `Start()`, and `Stop()`.
- [ ] `Stop()` signals the sweeper goroutine to exit cleanly.
- [ ] Middleware is applied globally in `RegisterRoutes()`.
- [ ] `RATE_LIMIT_RPS` (default `10`) and `RATE_LIMIT_BURST` (default `20`) are read from env at startup.
- [ ] Unit tests cover: requests within limit pass, requests over limit get 429, sweeper removes stale entries.
- [ ] Tests pass under `go test -race ./internal/server/ratelimiter/...`.

## Implementation Notes

### Package structure

```
internal/server/ratelimiter/
  ratelimiter.go
  ratelimiter_test.go
```

### Bucket struct

```go
type bucket struct {
    mu       sync.Mutex
    tokens   float64
    lastSeen time.Time
}
```

### Refill logic (called inside the mutex on each request)

```go
now := time.Now()
elapsed := now.Sub(b.lastSeen).Seconds()
b.tokens = min(burst, b.tokens + elapsed*rps)
b.lastSeen = now
```

### IP extraction

Use `r.RemoteAddr` as the key, stripping the port. For production, also check `X-Forwarded-For` (but guard against spoofing — only trust if behind a known proxy).

### Sweeper

```go
func (rl *RateLimiter) sweep() {
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            rl.buckets.Range(func(key, val any) bool {
                b := val.(*bucket)
                b.mu.Lock()
                stale := time.Since(b.lastSeen) > 5*time.Minute
                b.mu.Unlock()
                if stale {
                    rl.buckets.Delete(key)
                }
                return true
            })
        case <-rl.done:
            return
        }
    }
}
```
