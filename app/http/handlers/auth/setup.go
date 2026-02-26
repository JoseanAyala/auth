package auth

import (
	"auth-as-a-service/app/hasher"
	"auth-as-a-service/app/http/util"
	"auth-as-a-service/app/redis"
	userStore "auth-as-a-service/app/store/user"

	authMW "auth-as-a-service/app/http/middleware/auth"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	users  *userStore.Store
	hasher *hasher.Dispatcher
	redis  redis.Service
}

func New(users *userStore.Store, hasher *hasher.Dispatcher, redis redis.Service) *Handler {
	return &Handler{users: users, hasher: hasher, redis: redis}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", util.Handle(h.register))
		r.Post("/login", util.Handle(h.login))
		r.Post("/refresh", util.Handle(h.refresh))

		r.With(authMW.RequireAuth(h.redis)).
			Post("/logout", util.Handle(h.logout))
	})
}
