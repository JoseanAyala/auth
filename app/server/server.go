package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"auth-as-a-service/internal/database"
	"auth-as-a-service/internal/hasher"
	redis "auth-as-a-service/internal/redis"
	"auth-as-a-service/internal/server/middleware/ratelimiter"
	"auth-as-a-service/internal/store"
)

type Server struct {
	db          database.Service
	redis       redis.Service
	hasher      *hasher.Dispatcher
	store       *store.Registry
	rateLimiter *ratelimiter.RateLimiter
}

func NewServer() *http.Server {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic("server port value is not valid")
	}

	// Setup memory
	db := database.New()
	redis := redis.New()

	// Setup Worker
	h := hasher.NewDispatcher()
	h.Start()

	// Setup rate limiter
	rps := parseFloat(os.Getenv("RATE_LIMIT_RPS"), 10)
	burst := parseFloat(os.Getenv("RATE_LIMIT_BURST"), 20)
	rl := ratelimiter.New(rps, burst)
	rl.Start()

	Server := &Server{
		db:          db,
		redis:       redis,
		store:       store.New(db.DB()),
		hasher:      h,
		rateLimiter: rl,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      Server.Setup(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	server.RegisterOnShutdown(rl.Stop)

	return server
}

func parseFloat(s string, def float64) float64 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v <= 0 {
		return def
	}
	return v
}
