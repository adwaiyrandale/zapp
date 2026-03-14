package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

type Limiter struct {
	tokens   map[string]*tokenBucket
	mu       sync.RWMutex
	rate     float64
	capacity int64
	refill   time.Duration
}

type tokenBucket struct {
	tokens    int64
	lastCheck time.Time
}

func New(rate float64, capacity int64, refill time.Duration) *Limiter {
	l := &Limiter{
		tokens:   make(map[string]*tokenBucket),
		rate:     rate,
		capacity: capacity,
		refill:   refill,
	}
	go l.cleanup()
	return l
}

func (l *Limiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		l.mu.Lock()
		for key, tb := range l.tokens {
			if time.Since(tb.lastCheck) > 10*time.Minute {
				delete(l.tokens, key)
			}
		}
		l.mu.Unlock()
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	tb, exists := l.tokens[key]
	if !exists {
		l.tokens[key] = &tokenBucket{
			tokens:    l.capacity - 1,
			lastCheck: time.Now(),
		}
		return true
	}

	now := time.Now()
	elapsed := now.Sub(tb.lastCheck)
	refill := int64(elapsed.Seconds() * l.rate)
	tb.tokens = min(l.capacity, tb.tokens+refill)
	tb.lastCheck = now

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
			key = apiKey
		}

		if !l.Allow(key) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *Limiter) MiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
			key = apiKey
		}

		if !l.Allow(key) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func RateLimit(rate float64) func(chi.Router) {
	limiter := New(rate, 100, time.Second)
	return func(r chi.Router) {
		r.Use(limiter.Middleware)
	}
}
