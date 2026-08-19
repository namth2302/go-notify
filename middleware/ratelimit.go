package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/namth/go-notify/message"
)

// RateLimiter implements a standard token bucket rate limiter.
type RateLimiter struct {
	mu         sync.Mutex
	rate       float64       // Tokens added per second
	capacity   float64       // Maximum burst capacity
	tokens     float64       // Current token count
	lastRefill time.Time     // Last refill timestamp
}

// NewRateLimiter creates a new Token Bucket Rate Limiter.
// rps: tokens generated per second (e.g. 5.0).
// burst: maximum token capacity (e.g. 10).
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	if rps <= 0 {
		rps = 1.0
	}
	if burst <= 0 {
		burst = 1
	}
	return &RateLimiter{
		rate:       rps,
		capacity:   float64(burst),
		tokens:     float64(burst),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(r.lastRefill).Seconds()
		r.tokens = min(r.capacity, r.tokens+elapsed*r.rate)
		r.lastRefill = now

		if r.tokens >= 1.0 {
			r.tokens -= 1.0
			r.mu.Unlock()
			return nil
		}

		// Calculate wait time for 1 token
		missing := 1.0 - r.tokens
		waitTime := time.Duration((missing / r.rate) * float64(time.Second))
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}
}

// RateLimit returns a middleware that throttles requests through the provided limiter.
func RateLimit(limiter *RateLimiter) Middleware {
	if limiter == nil {
		return func(next Handler) Handler { return next }
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, msg message.Message) (*message.Result, error) {
			if err := limiter.Wait(ctx); err != nil {
				return nil, fmt.Errorf("ratelimit wait: %w", err)
			}
			return next(ctx, msg)
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
