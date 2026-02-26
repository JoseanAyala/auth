package ratelimiter

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	defaultSweepInterval = 60 * time.Second
	defaultStaleDuration = 5 * time.Minute
)

type bucket struct {
	mu       sync.Mutex
	tokens   float64
	lastSeen time.Time
}

// RateLimiter enforces per-IP token bucket rate limits with a background sweeper.
type RateLimiter struct {
	buckets       sync.Map
	rps           float64
	burst         float64
	sweepInterval time.Duration
	staleDuration time.Duration
	done          chan struct{}
}

// New creates a RateLimiter with production defaults (60s sweep, 5min stale).
func New(rps, burst float64) *RateLimiter {
	return newWithConfig(rps, burst, defaultSweepInterval, defaultStaleDuration)
}

func newWithConfig(rps, burst float64, sweepInterval, staleDuration time.Duration) *RateLimiter {
	return &RateLimiter{
		rps:           rps,
		burst:         burst,
		sweepInterval: sweepInterval,
		staleDuration: staleDuration,
		done:          make(chan struct{}),
	}
}

// Start launches the background sweeper goroutine.
func (rl *RateLimiter) Start() {
	go rl.sweep()
}

// Stop signals the sweeper goroutine to exit cleanly.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

// Middleware returns an http.Handler middleware that enforces the rate limit.
// Requests that exceed the limit receive 429 with a Retry-After header.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)

			val, _ := rl.buckets.LoadOrStore(ip, &bucket{
				tokens:   rl.burst,
				lastSeen: time.Now(),
			})
			b := val.(*bucket)

			b.mu.Lock()
			now := time.Now()
			elapsed := now.Sub(b.lastSeen).Seconds()
			b.tokens = min(rl.burst, b.tokens+elapsed*rl.rps)
			b.lastSeen = now

			if b.tokens < 1 {
				retryAfter := int(math.Ceil((1 - b.tokens) / rl.rps))
				if retryAfter < 1 {
					retryAfter = 1
				}
				b.mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded"}`))
				return
			}

			b.tokens--
			b.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) sweep() {
	ticker := time.NewTicker(rl.sweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.buckets.Range(func(key, val any) bool {
				b := val.(*bucket)
				b.mu.Lock()
				stale := time.Since(b.lastSeen) > rl.staleDuration
				b.mu.Unlock()
				if stale {
					rl.buckets.Delete(key)
				}
				return true
			})
		case <-rl.done:
			return
		}
	}
}

func (rl *RateLimiter) bucketCount() int {
	count := 0
	rl.buckets.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func extractIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
