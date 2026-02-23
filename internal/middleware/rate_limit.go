package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

func newIPLimiter(r rate.Limit, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
}

func (l *ipLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.limiters[ip] = limiter
	}
	return limiter
}

func RateLimit(rps float64, burst int) func(http.Handler) http.Handler {
	limiter := newIPLimiter(rate.Limit(rps), burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !limiter.getLimiter(ip).Allow() {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
