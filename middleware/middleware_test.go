package middleware_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/namth2302/go-notify/message"
	"github.com/namth2302/go-notify/middleware"
)

func TestChainOrder(t *testing.T) {
	var executionOrder []string

	m1 := func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, msg message.Message) (*message.Result, error) {
			executionOrder = append(executionOrder, "m1_start")
			res, err := next(ctx, msg)
			executionOrder = append(executionOrder, "m1_end")
			return res, err
		}
	}

	m2 := func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, msg message.Message) (*message.Result, error) {
			executionOrder = append(executionOrder, "m2_start")
			res, err := next(ctx, msg)
			executionOrder = append(executionOrder, "m2_end")
			return res, err
		}
	}

	finalHandler := func(ctx context.Context, msg message.Message) (*message.Result, error) {
		executionOrder = append(executionOrder, "final")
		return &message.Result{StatusCode: 200}, nil
	}

	chained := middleware.Chain(m1, m2)(finalHandler)
	_, _ = chained(context.Background(), message.NewText("hello"))

	expected := []string{"m1_start", "m2_start", "final", "m2_end", "m1_end"}
	if len(executionOrder) != len(expected) {
		t.Fatalf("unexpected order length: got %v, want %v", executionOrder, expected)
	}
	for i := range expected {
		if executionOrder[i] != expected[i] {
			t.Errorf("at index %d: got %s, want %s", i, executionOrder[i], expected[i])
		}
	}
}

func TestRetryMiddleware(t *testing.T) {
	var attempts int32

	// Handler that fails 2 times and succeeds on 3rd attempt
	handler := func(ctx context.Context, msg message.Message) (*message.Result, error) {
		cur := atomic.AddInt32(&attempts, 1)
		if cur < 3 {
			return &message.Result{StatusCode: http.StatusServiceUnavailable}, errors.New("temporary error")
		}
		return &message.Result{StatusCode: http.StatusOK}, nil
	}

	retryMw := middleware.Retry(3,
		middleware.WithInitialInterval(5*time.Millisecond),
		middleware.WithMaxInterval(20*time.Millisecond),
	)

	wrapped := retryMw(handler)
	res, err := wrapped(context.Background(), message.NewText("test retry"))
	if err != nil {
		t.Fatalf("expected success on 3rd attempt, got error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestRetryContextCancelled(t *testing.T) {
	handler := func(ctx context.Context, msg message.Message) (*message.Result, error) {
		return &message.Result{StatusCode: 500}, errors.New("fail")
	}

	retryMw := middleware.Retry(5, middleware.WithInitialInterval(100*time.Millisecond))
	wrapped := retryMw(handler)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := wrapped(ctx, message.NewText("cancel"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded error, got %v", err)
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := middleware.NewRateLimiter(10, 2)
	mw := middleware.RateLimit(limiter)

	var count int32
	handler := mw(func(ctx context.Context, msg message.Message) (*message.Result, error) {
		atomic.AddInt32(&count, 1)
		return &message.Result{StatusCode: 200}, nil
	})

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, err := handler(ctx, message.NewText("rate limit"))
		if err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
	}
	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("expected 3 calls, got %d", atomic.LoadInt32(&count))
	}
}

func TestLoggingMiddleware(t *testing.T) {
	mw := middleware.Logging(slog.Default())
	handler := mw(func(ctx context.Context, msg message.Message) (*message.Result, error) {
		return &message.Result{Provider: "test", StatusCode: 200}, nil
	})

	_, err := handler(context.Background(), message.NewText("log test"))
	if err != nil {
		t.Fatalf("logging handler failed: %v", err)
	}

	// Test with error
	errHandler := mw(func(ctx context.Context, msg message.Message) (*message.Result, error) {
		return &message.Result{Provider: "test", StatusCode: 500}, errors.New("mock error")
	})
	_, err = errHandler(context.Background(), message.NewText("log error"))
	if err == nil {
		t.Errorf("expected error from handler")
	}
}
