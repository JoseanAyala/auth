# Task 07 — Build the Courier: async notification service with context timeout

## Goal

Decouple slow external calls (email, SMS) from the request path using fire-and-forget goroutines, with a context timeout to prevent goroutine leaks if a provider hangs.

## Blocked by

- **#04** — Auth handlers must exist to call the Courier after successful register/login.

## Acceptance Criteria

- [ ] `internal/courier/courier.go` defines a `Notifier` interface with `SendWelcome` and `SendMFA`.
- [ ] A `LogNotifier` stub satisfies the interface by logging the event (no real provider needed now).
- [ ] `Send(n Notifier, fn func(context.Context) error)` fires a goroutine, wraps the call with `context.WithTimeout`, and logs any error — it never blocks the caller.
- [ ] `NOTIF_TIMEOUT_SEC` env var controls the timeout (default `5`).
- [ ] Register handler calls `courier.Send` for `SendWelcome` after a successful registration.
- [ ] Login handler calls `courier.Send` for `SendMFA` after a successful login.
- [ ] Unit test verifies that a slow notifier (sleeps longer than timeout) does not block the caller and logs a timeout error.
- [ ] Tests pass under `go test -race ./internal/courier/...`.

## Implementation Notes

### Package structure

```
internal/courier/
  courier.go
  courier_test.go
```

### Notifier interface

```go
type Notifier interface {
    SendWelcome(ctx context.Context, email string) error
    SendMFA(ctx context.Context, email, code string) error
}
```

### Send helper

```go
func Send(n Notifier, fn func(context.Context) error) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), timeout)
        defer cancel()
        if err := fn(ctx); err != nil {
            log.Printf("courier: notification failed: %v", err)
        }
    }()
}
```

### Usage in handler

```go
courier.Send(s.notifier, func(ctx context.Context) error {
    return s.notifier.SendWelcome(ctx, req.Email)
})
```

### Future extension

When a real provider is needed, implement a new type (e.g. `SendGridNotifier`) satisfying `Notifier` and swap it in via the `Server` struct — no handler changes required.
