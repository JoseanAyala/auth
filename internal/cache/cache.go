package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Health() map[string]string
	Close() error
}

type service struct {
	client *redis.Client
}

var (
	host          = os.Getenv("REDIS_HOST")
	port          = os.Getenv("REDIS_PORT")
	password      = os.Getenv("REDIS_PASSWORD")
	cacheInstance *service
)

func New() Service {
	if cacheInstance != nil {
		return cacheInstance
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})

	cacheInstance = &service{client: client}
	return cacheInstance
}

func (s *service) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *service) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *service) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	start := time.Now()
	_, err := s.client.Ping(ctx).Result()
	latency := time.Since(start)

	if err != nil {
		stats["redis_status"] = "down"
		stats["redis_error"] = fmt.Sprintf("redis down: %v", err)
		log.Printf("redis down: %v", err)
		return stats
	}

	stats["redis_status"] = "up"
	stats["redis_latency"] = latency.String()

	poolStats := s.client.PoolStats()
	stats["redis_connections"] = fmt.Sprintf("%d hits, %d misses, %d timeouts", poolStats.Hits, poolStats.Misses, poolStats.Timeouts)

	return stats
}

func (s *service) Close() error {
	log.Printf("Disconnected from Redis: %s:%s", host, port)
	return s.client.Close()
}
