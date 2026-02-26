package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	authHandler "auth-as-a-service/internal/server/handlers/auth"
	"auth-as-a-service/internal/server/handlers/health"
)

func (s *Server) Setup() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(s.rateLimiter.Middleware())
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Setup health endpoint
	health.New(s.db, s.redis).RegisterRoutes(r)

	// Setup auth handler
	authHandler.New(s.store.Users, s.hasher, s.redis).RegisterRoutes(r)

	return r
}
