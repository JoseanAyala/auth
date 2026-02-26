package redis

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func mustStartRedisContainer() (func(context.Context, ...testcontainers.TerminateOption) error, error) {
	redisContainer, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "redis:latest",
				ExposedPorts: []string{"6379/tcp"},
				WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(5 * time.Second),
			},
			Started: true,
		},
	)
	if err != nil {
		return nil, err
	}

	redisHost, err := redisContainer.Host(context.Background())
	if err != nil {
		return redisContainer.Terminate, err
	}

	redisPort, err := redisContainer.MappedPort(context.Background(), "6379/tcp")
	if err != nil {
		return redisContainer.Terminate, err
	}

	host = redisHost
	port = redisPort.Port()
	password = ""

	return redisContainer.Terminate, nil
}

func TestMain(m *testing.M) {
	teardown, err := mustStartRedisContainer()
	if err != nil {
		log.Fatalf("could not start redis container: %v", err)
	}

	m.Run()

	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown redis container: %v", err)
	}
}

func TestNew(t *testing.T) {
	srv := New()
	if srv == nil {
		t.Fatal("New() returned nil")
	}
}

func TestHealth(t *testing.T) {
	srv := New()

	stats := srv.Health()

	if stats["redis_status"] != "up" {
		t.Fatalf("expected redis_status to be up, got %s", stats["redis_status"])
	}

	if _, ok := stats["redis_latency"]; !ok {
		t.Fatal("expected redis_latency to be present")
	}
}

func TestSetAndGet(t *testing.T) {
	srv := New()
	ctx := context.Background()

	if err := srv.Set(ctx, "test-key", "test-value", time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}

	val, err := srv.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "test-value" {
		t.Fatalf("expected test-value, got %s", val)
	}
}

func TestGetMissing(t *testing.T) {
	srv := New()
	ctx := context.Background()

	_, err := srv.Get(ctx, "nonexistent-key")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestDelete(t *testing.T) {
	srv := New()
	ctx := context.Background()

	if err := srv.Set(ctx, "delete-me", "value", time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := srv.Delete(ctx, "delete-me"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := srv.Get(ctx, "delete-me")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestSetWithTTL(t *testing.T) {
	srv := New()
	ctx := context.Background()

	if err := srv.Set(ctx, "ttl-key", "expires", 100*time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	_, err := srv.Get(ctx, "ttl-key")
	if err == nil {
		t.Fatal("expected key to have expired")
	}
}

func TestClose(t *testing.T) {
	srv := New()

	if err := srv.Close(); err != nil {
		t.Fatalf("expected Close() to return nil, got %v", err)
	}

	// Reset singleton so other tests aren't affected.
	redisInstance = nil
}
