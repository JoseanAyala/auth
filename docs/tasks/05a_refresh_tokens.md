# Task 05a — Refresh tokens with rotation and dual-revoke logout

## Goal

Extend the JWT system with long-lived refresh tokens. On login the service issues both an access token (short-lived) and a refresh token (long-lived). `POST /auth/refresh` validates the refresh token, issues a new access + refresh pair, and blacklists the old refresh token. `POST /auth/logout` revokes both tokens.

## Blocked by

- **#05** — JWT generation, validation, and Redis blacklist must be in place.

## Acceptance Criteria

- [ ] `internal/token/token.go` gains `GenerateRefresh` and `ValidateRefresh`.
- [ ] `Generate` adds `token_type: "access"` claim; `GenerateRefresh` adds `token_type: "refresh"`.
- [ ] `Validate` (used by `RequireAuth`) rejects tokens where `token_type == "refresh"`.
- [ ] `ValidateRefresh` rejects tokens where `token_type != "refresh"`.
- [ ] `POST /auth/login` returns `{ access_token, refresh_token }`.
- [ ] `POST /auth/refresh` accepts `Authorization: Bearer <refresh_token>`, validates it, revokes the old refresh token, and returns a new `{ access_token, refresh_token }`.
- [ ] `POST /auth/logout` accepts `Authorization: Bearer <access_token>` and a JSON body `{ "refresh_token": "..." }`; revokes both tokens; returns `204 No Content`.
- [ ] `REFRESH_TOKEN_EXPIRY_DAYS` (default `30`) is read from env.
- [ ] Unit tests cover: valid refresh token, expired refresh token, refresh token used as access token (rejected), rotated token reuse (rejected), logout revokes both.

## Environment Variables

| Variable                    | Default | Purpose                        |
|-----------------------------|---------|--------------------------------|
| `JWT_SECRET`                | —       | Signing key (existing)         |
| `JWT_EXPIRY_HOURS`          | `24`    | Access token lifetime (existing) |
| `REFRESH_TOKEN_EXPIRY_DAYS` | `30`    | Refresh token lifetime         |

## Implementation Notes

### Package structure (changes only)

```
internal/token/
  token.go          ← add GenerateRefresh, ValidateRefresh; patch Generate + Validate
  token_test.go     ← new refresh token test cases

internal/server/handlers/auth/
  auth.go           ← update LoginHandler, LogoutHandler; add RefreshHandler
  model.go          ← update loginResponse; add refreshRequest, refreshResponse
  setup.go          ← add POST /auth/refresh route (public)
```

### token.go additions

```go
// GenerateRefresh signs a long-lived refresh token with token_type "refresh".
func GenerateRefresh(userID string) (string, error)

// ValidateRefresh verifies signature, checks expiry, checks blacklist,
// and requires token_type == "refresh". Returns userID.
func ValidateRefresh(ctx context.Context, tokenString string, cache redis.Service) (string, error)
```

`Generate` and `GenerateRefresh` both add a `token_type` claim (`"access"` / `"refresh"`).
`Validate` returns an error when `token_type == "refresh"` so refresh tokens cannot be used as bearer tokens.

### Rotation on POST /auth/refresh

```go
func (h *Handler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
    tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
    userID, err := token.ValidateRefresh(r.Context(), tokenString, h.redis)
    // on error → 401
    newAccess, _   := token.Generate(userID)
    newRefresh, _  := token.GenerateRefresh(userID)
    token.Revoke(r.Context(), tokenString, h.redis)   // blacklist old refresh token
    writeJSON(w, http.StatusOK, refreshResponse{AccessToken: newAccess, RefreshToken: newRefresh})
}
```

### Dual-revoke on POST /auth/logout

```go
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
    // access token already validated by RequireAuth middleware
    accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
    var body struct{ RefreshToken string `json:"refresh_token"` }
    json.NewDecoder(r.Body).Decode(&body)

    token.Revoke(r.Context(), accessToken, h.redis)
    if body.RefreshToken != "" {
        token.Revoke(r.Context(), body.RefreshToken, h.redis)
    }
    w.WriteHeader(http.StatusNoContent)
}
```

### Route registration (setup.go)

```go
func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Route("/auth", func(r chi.Router) {
        r.Post("/register", h.RegisterHandler)
        r.Post("/login",    h.LoginHandler)
        r.Post("/refresh",  h.RefreshHandler)   // public — refresh token is its own credential
        r.With(middleware.RequireAuth(h.redis)).Post("/logout", h.LogoutHandler)
    })
}
```

### model.go changes

```go
type loginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}
```
