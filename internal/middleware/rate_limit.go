package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/developer-space/api/internal/response"
)

type rateLimitEntry struct {
	count    int
	resetAt  time.Time
}

// RateLimiter is an in-memory per-key rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	max     int
	window  time.Duration
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		max:     max,
		window:  window,
	}
}

// Allow checks if a request for the given key is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists || now.After(entry.resetAt) {
		rl.entries[key] = &rateLimitEntry{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return true
	}

	if entry.count >= rl.max {
		return false
	}

	entry.count++
	return true
}

// RateLimit middleware wraps a handler with rate limiting by a key extracted from the request.
func RateLimit(rl *RateLimiter, keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if !rl.Allow(key) {
				slog.Warn("rate limit exceeded", "key", key)
				response.Error(w, http.StatusTooManyRequests, "Too many requests. Try again later.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
