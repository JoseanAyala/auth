package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestRateLimiter_AllowsRequestsWithinLimit(t *testing.T) {
	t.Parallel()
	const burst = 5
	rl := newWithConfig(10, burst, time.Hour, time.Hour)
	rl.Start()
	defer rl.Stop()

	handler := rl.Middleware()(okHandler)

	for i := range burst {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}
}

func TestRateLimiter_BlocksRequestsOverLimit(t *testing.T) {
	t.Parallel()
	const burst = 3
	rl := newWithConfig(10, burst, time.Hour, time.Hour)
	rl.Start()
	defer rl.Stop()

	handler := rl.Middleware()(okHandler)

	// Exhaust the token bucket.
	for i := range burst {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:5678"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200 while filling, got %d", i+1, rr.Code)
		}
	}

	// Next request must be rate-limited.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:5678"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header to be set")
	}
}

func TestRateLimiter_SweeperRemovesStaleEntries(t *testing.T) {
	t.Parallel()
	const staleDur = 100 * time.Millisecond
	const sweepInterval = 30 * time.Millisecond

	rl := newWithConfig(10, 5, sweepInterval, staleDur)
	rl.Start()
	defer rl.Stop()

	handler := rl.Middleware()(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:9999"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if rl.bucketCount() != 1 {
		t.Fatal("expected one bucket entry after the request")
	}

	// Wait for the entry to become stale and for the sweeper to run at least once.
	time.Sleep(staleDur + sweepInterval*5)

	if rl.bucketCount() != 0 {
		t.Fatal("expected sweeper to remove the stale bucket entry")
	}
}
