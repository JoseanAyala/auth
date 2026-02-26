package health

import (
	"net/http"

	"auth-as-a-service/app/http/httpkit"
	"auth-as-a-service/app/memory/database"
	"auth-as-a-service/app/memory/redis"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	db    database.Service
	redis redis.Service
}

func New(db database.Service, redis redis.Service) *Handler {
	return &Handler{
		db:    db,
		redis: redis,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", httpkit.Handle(h.healthHandler))
}

func (h *Handler) healthHandler(r *http.Request) (*httpkit.Response, error) {
	resp := map[string]any{
		"database": h.db.Health(),
		"redis":    h.redis.Health(),
	}

	return &httpkit.Response{
		Status: http.StatusOK,
		Body:   resp,
	}, nil
}
