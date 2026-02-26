package health

import (
	"net/http"

	"auth-as-a-service/app/database"
	"auth-as-a-service/app/http/util"
	"auth-as-a-service/app/redis"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	db    database.Service
	redis redis.Service
}

func New(db database.Service, cache redis.Service) *Handler {
	return &Handler{
		db:    db,
		redis: cache,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", util.Handle(h.healthHandler))
}

func (h *Handler) healthHandler(r *http.Request) (*util.Response, error) {
	resp := map[string]any{
		"database": h.db.Health(),
		"redis":    h.redis.Health(),
	}

	return &util.Response{
		Status: http.StatusOK,
		Body:   resp,
	}, nil
}
