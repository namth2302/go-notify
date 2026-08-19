package middleware

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/namth/go-notify/message"
)

// RetryConfig configures retry behavior.
type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	IsRetryable     func(res *message.Result, err error) bool
}

// RetryOption configures retry settings.
type RetryOption func(*RetryConfig)

// WithInitialInterval sets starting backoff duration.
func WithInitialInterval(d time.Duration) RetryOption {
	return func(c *RetryConfig) {
		c.InitialInterval = d
	}
}

// WithMaxInterval sets maximum cap for backoff duration.
func WithMaxInterval(d time.Duration) RetryOption {
	return func(c *RetryConfig) {
		c.MaxInterval = d
	}
}

// WithRetryPredicate sets custom check for retryable errors.
func WithRetryPredicate(fn func(res *message.Result, err error) bool) RetryOption {
	return func(c *RetryConfig) {
		c.IsRetryable = fn
	}
}

// DefaultRetryable returns true if status code is 429 or 5xx or network error.
func DefaultRetryable(res *message.Result, err error) bool {
	if err != nil {
		if res == nil {
			return true // Network / IO error
		}
		if res.StatusCode == http.StatusTooManyRequests || (res.StatusCode >= 500 && res.StatusCode <= 599) {
			return true
		}
	}
	return false
}

// Retry returns an interceptor that retries failed requests with exponential backoff and jitter.
func Retry(maxAttempts int, opts ...RetryOption) Middleware {
	cfg := RetryConfig{
		MaxAttempts:     maxAttempts,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     3 * time.Second,
		Multiplier:      2.0,
		IsRetryable:     DefaultRetryable,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, msg message.Message) (*message.Result, error) {
			var lastRes *message.Result
			var lastErr error

			interval := cfg.InitialInterval

			for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
				lastRes, lastErr = next(ctx, msg)
				if lastErr == nil {
					return lastRes, nil
				}

				if attempt == cfg.MaxAttempts || !cfg.IsRetryable(lastRes, lastErr) {
					return lastRes, lastErr
				}

				// Calculate sleep duration with jitter
				sleepDur := calculateBackoff(interval, cfg.MaxInterval)
				interval = time.Duration(float64(interval) * cfg.Multiplier)

				select {
				case <-ctx.Done():
					return lastRes, ctx.Err()
				case <-time.After(sleepDur):
				}
			}

			return lastRes, lastErr
		}
	}
}

func calculateBackoff(base, max time.Duration) time.Duration {
	if base > max {
		base = max
	}
	if base <= 0 {
		return 0
	}
	// Full jitter: random between [base/2, base]
	half := base / 2
	jitter := time.Duration(rand.Int63n(int64(half + 1)))
	return half + jitter
}
