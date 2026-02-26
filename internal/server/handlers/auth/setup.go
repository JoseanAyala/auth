package auth

import (
	"auth-as-a-service/internal/hasher"
	userstore "auth-as-a-service/internal/store/user"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	users  *userstore.Store
	hasher *hasher.Dispatcher
}

func New(users *userstore.Store, hasher *hasher.Dispatcher) *Handler {
	return &Handler{users: users, hasher: hasher}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.RegisterHandler)
		r.Post("/login", h.LoginHandler)
	})
}
