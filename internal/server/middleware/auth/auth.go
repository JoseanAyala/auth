package middleware

import (
	"context"
	"net/http"
	"strings"

	"auth-as-a-service/internal/redis"
	"auth-as-a-service/internal/token"
)

type contextKey string

const UserIDKey contextKey = "userID"

func RequireAuth(cache redis.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := token.Validate(r.Context(), tokenString, cache)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
