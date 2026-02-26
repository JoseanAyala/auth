package auth

import (
	"auth-as-a-service/internal/hasher"
	"auth-as-a-service/internal/redis"
	authmw "auth-as-a-service/internal/server/middleware"
	userstore "auth-as-a-service/internal/store/user"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	users  *userstore.Store
	hasher *hasher.Dispatcher
	redis  redis.Service
}

func New(users *userstore.Store, hasher *hasher.Dispatcher, redis redis.Service) *Handler {
	return &Handler{users: users, hasher: hasher, redis: redis}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.RegisterHandler)
		r.Post("/login", h.LoginHandler)
		r.With(authmw.RequireAuth(h.redis)).Post("/logout", h.LogoutHandler)
	})
}
