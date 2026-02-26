package auth

import (
	"auth-as-a-service/app/http/httpkit"
	"auth-as-a-service/app/memory/redis"
	userStore "auth-as-a-service/app/memory/store/user"

	authMW "auth-as-a-service/app/http/middleware/auth"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	users *userStore.Store
	redis redis.Service
}

func New(users *userStore.Store, redis redis.Service) *Handler {
	return &Handler{users: users, redis: redis}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", httpkit.Handle(h.register))
		r.Post("/login", httpkit.Handle(h.login))
		r.Post("/refresh", httpkit.Handle(h.refresh))

		r.With(authMW.RequireAuth(h.redis)).
			Post("/logout", httpkit.Handle(h.logout))
	})
}
