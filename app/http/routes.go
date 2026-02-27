package server

import (
	"net/http"
	"os"
	"strings"

	authHandler "auth-as-a-service/app/http/handlers/auth"
	"auth-as-a-service/app/http/handlers/health"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) Setup() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(s.rateLimiter.Middleware())
	allowedOrigins := []string{"http://localhost:3000"}
	if origins := os.Getenv("CORS_ORIGINS"); origins != "" {
		allowedOrigins = strings.Split(origins, ",")
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Setup health endpoint
	health.New(s.db, s.redis).RegisterRoutes(r)

	// Setup auth handler
	authHandler.New(s.store.Users, s.redis).RegisterRoutes(r)

	return r
}
