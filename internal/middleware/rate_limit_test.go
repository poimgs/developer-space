package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_AllowsUpToMax(t *testing.T) {
	rl := NewRateLimiter(5, 1*time.Hour)

	for i := 0; i < 5; i++ {
		if !rl.Allow("user@example.com") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_BlocksAfterMax(t *testing.T) {
	rl := NewRateLimiter(5, 1*time.Hour)

	for i := 0; i < 5; i++ {
		rl.Allow("user@example.com")
	}

	if rl.Allow("user@example.com") {
		t.Error("6th request should be blocked")
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Hour)

	rl.Allow("user1@example.com")
	rl.Allow("user1@example.com")

	if rl.Allow("user1@example.com") {
		t.Error("user1 3rd request should be blocked")
	}

	// user2 should still be allowed
	if !rl.Allow("user2@example.com") {
		t.Error("user2 should be allowed (different key)")
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)

	rl.Allow("key")
	rl.Allow("key")

	if rl.Allow("key") {
		t.Error("should be blocked before window expires")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("key") {
		t.Error("should be allowed after window expires")
	}
}

func TestRateLimitMiddleware_Returns429(t *testing.T) {
	rl := NewRateLimiter(1, 1*time.Hour)

	mw := RateLimit(rl, func(r *http.Request) string {
		return "test-key"
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request: allowed
	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec.Code)
	}

	// Second request: blocked
	req = httptest.NewRequest("POST", "/test", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", rec.Code)
	}
}
