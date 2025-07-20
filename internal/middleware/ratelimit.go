package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Simple in-memory rate limiter (for production, use Redis or similar)
type RateLimiter struct {
	clients map[string]*ClientLimiter
	mutex   sync.RWMutex
	limit   int
	window  time.Duration
}

type ClientLimiter struct {
	requests []time.Time
	mutex    sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*ClientLimiter),
		limit:   limit,
		window:  window,
	}
}

// RateLimit middleware
func (rl *RateLimiter) RateLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !rl.allowRequest(clientIP) {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
				w.Header().Set("X-RateLimit-Window", rl.window.String())
				w.Header().Set("Retry-After", "60")

				http.Error(w, `{"success":false,"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Rate limit exceeded"}}`,
					http.StatusTooManyRequests)
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
			w.Header().Set("X-RateLimit-Window", rl.window.String())

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) allowRequest(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	client, exists := rl.clients[clientIP]
	if !exists {
		client = &ClientLimiter{
			requests: make([]time.Time, 0),
		}
		rl.clients[clientIP] = client
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove old requests
	validRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	client.requests = validRequests

	// Check if limit exceeded
	if len(client.requests) >= rl.limit {
		return false
	}

	// Add current request
	client.requests = append(client.requests, now)
	return true
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
