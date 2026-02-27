package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	Health() map[string]string
	Close() error
	DB() *sqlx.DB
}

type service struct {
	db *sqlx.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	sslmode    = os.Getenv("DB_SSLMODE")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	if sslmode == "" {
		sslmode = "require"
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s", username, password, host, port, database, sslmode, schema)
	db, err := sqlx.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	start := time.Now()
	err := s.db.PingContext(ctx)
	latency := time.Since(start)

	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["latency"] = latency.String()

	dbStats := s.db.Stats()
	stats["connections"] = fmt.Sprintf("%d open, %d in use, %d idle", dbStats.OpenConnections, dbStats.InUse, dbStats.Idle)

	if dbStats.WaitCount > 0 {
		stats["wait_count"] = fmt.Sprintf("%d (total wait: %s)", dbStats.WaitCount, dbStats.WaitDuration)
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}

func (s *service) DB() *sqlx.DB {
	return s.db
}
