package health

import (
	"encoding/json"
	"net/http"

	"auth-as-a-service/internal/database"
	"auth-as-a-service/internal/redis"

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
	r.Get("/health", h.healthHandler)
}

func (s *Handler) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"database": s.db.Health(),
		"redis":    s.redis.Health(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
