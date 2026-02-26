# Task 05 — Implement JWT generation, validation, and Redis-backed blacklisting

## Goal

Issue signed JWTs on login, validate them on protected routes, and revoke them on logout using Redis as the blacklist store.

## Blocked by

- **#01** — Redis client must be available for blacklist storage.
- **#04** — Login handler must exist to call token generation.

## Acceptance Criteria

- [ ] `internal/token/token.go` exposes `Generate`, `Validate`, and `Revoke`.
- [ ] `Generate` signs a JWT with HS256 using `JWT_SECRET` from env; includes `sub` (user ID), `jti` (UUID), and `exp` claims.
- [ ] `Validate` verifies the signature, checks expiry, and rejects tokens whose `jti` is present in Redis.
- [ ] `Revoke` writes `jti → "1"` to Redis with a TTL equal to the token's remaining lifetime.
- [ ] `POST /auth/logout` reads the `Authorization: Bearer <token>` header, validates it, revokes the `jti`, and returns `204 No Content`.
- [ ] A `RequireAuth` chi middleware validates the token and injects the user ID into the request context.
- [ ] Unit tests cover: valid token, expired token, tampered signature, revoked token.
- [ ] `JWT_SECRET`, `JWT_EXPIRY_HOURS` (default 24) are read from env.

## Implementation Notes

### Package structure

```
internal/token/
  token.go
  token_test.go

internal/server/middleware/
  auth.go          ← RequireAuth chi middleware

internal/server/handlers/auth/
  setup.go         ← Handler struct gains cache dep; RegisterRoutes adds /logout
  auth.go          ← LogoutHandler added here
  model.go         ← (unchanged)
```

### Dependency

```
go get github.com/golang-jwt/jwt/v5
```

### Generate

```go
func Generate(userID string, cache cache.Service) (string, error)
```

Creates a `jti` via `github.com/google/uuid` (already indirect dep), signs with `jwt.SigningMethodHS256`.

### Revoke

```go
func Revoke(ctx context.Context, tokenString string, cache cache.Service) error
```

Parses (without validation) to extract `jti` and `exp`, then calls `cache.Set(ctx, "blacklist:"+jti, "1", ttl)`.

### RequireAuth middleware

Lives in `internal/server/middleware/auth.go`:

```go
func RequireAuth(cache cache.Service) func(http.Handler) http.Handler
```

Extracts Bearer token → calls `token.Validate` → on success sets `userID` in context via a typed key.

### Auth Handler changes

`internal/server/handlers/auth/setup.go` — add `cache cache.Service` to `Handler` and update `New`; apply `RequireAuth` to the logout route:

```go
func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Route("/auth", func(r chi.Router) {
        r.Post("/register", h.RegisterHandler)
        r.Post("/login", h.LoginHandler)
        r.With(middleware.RequireAuth(h.cache)).Post("/logout", h.LogoutHandler)
    })
}
```

`internal/server/handlers/auth/auth.go` — add `LogoutHandler`:

```go
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
    // token already validated by RequireAuth; revoke it
    tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
    if err := token.Revoke(r.Context(), tokenString, h.cache); err != nil {
        writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
```

### Wiring in routes.go

Pass `s.redis` when constructing the auth handler:

```go
authHandler.New(s.store.Users, s.hasher, s.redis).RegisterRoutes(r)
```
