package auth

import (
	"auth-as-a-service/internal/hasher"
	"auth-as-a-service/internal/redis"
	"auth-as-a-service/internal/server/handler"
	authmw "auth-as-a-service/internal/server/middleware/auth"
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
		r.Post("/register", handler.Handle(h.register))
		r.Post("/login", handler.Handle(h.login))
		r.Post("/refresh", handler.Handle(h.refresh))

		r.With(authmw.RequireAuth(h.redis)).
			Post("/logout", handler.Handle(h.logout))
	})
}
