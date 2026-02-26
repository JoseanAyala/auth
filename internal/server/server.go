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
	"auth-as-a-service/internal/store"
)

type Server struct {
	db     database.Service
	redis  redis.Service
	hasher *hasher.Dispatcher
	store  *store.Registry
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

	Server := &Server{
		db:     db,
		redis:  redis,
		store:  store.New(db.DB()),
		hasher: h,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      Server.Setup(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
