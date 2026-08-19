package middleware

import (
	"context"

	"github.com/namth2302/go-notify/message"
)

// Handler represents the function signature for sending a notification.
type Handler func(ctx context.Context, msg message.Message) (*message.Result, error)

// Middleware is an interceptor function wrapping a Handler.
type Middleware func(next Handler) Handler

// Chain combines multiple middlewares into a single middleware.
// Middleware execution order: first middleware in slice is the outermost (executes first).
func Chain(middlewares ...Middleware) Middleware {
	return func(final Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			if middlewares[i] != nil {
				final = middlewares[i](final)
			}
		}
		return final
	}
}
