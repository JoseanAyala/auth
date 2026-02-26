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

```go
func RequireAuth(tokenSvc *token.Service) func(http.Handler) http.Handler
```

Extracts Bearer token → calls `Validate` → on success sets `userID` in context via a typed key.

### Logout handler

Register at `POST /auth/logout`, protect it with `RequireAuth`.
