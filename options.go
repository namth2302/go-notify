package notify

import (
	"net/http"

	"github.com/namth/go-notify/middleware"
	"github.com/namth/go-notify/template"
)

// Option configures the Notifier engine.
type Option func(*Notifier)

// WithMiddleware appends middlewares to the execution pipeline.
func WithMiddleware(mws ...middleware.Middleware) Option {
	return func(n *Notifier) {
		n.middlewares = append(n.middlewares, mws...)
	}
}

// WithRegistry sets a custom template registry.
func WithRegistry(reg *template.Registry) Option {
	return func(n *Notifier) {
		n.registry = reg
	}
}

// WithHTTPClient sets a shared HTTP client across created providers.
func WithHTTPClient(client *http.Client) Option {
	return func(n *Notifier) {
		n.httpClient = client
	}
}
