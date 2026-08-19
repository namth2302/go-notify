package notify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/namth/go-notify/message"
	"github.com/namth/go-notify/middleware"
	"github.com/namth/go-notify/provider/gchat"
	"github.com/namth/go-notify/provider/lark"
	"github.com/namth/go-notify/router"
	"github.com/namth/go-notify/template"
)

// Sender is the primary public interface for sending notifications.
type Sender interface {
	Send(ctx context.Context, msg message.Message) (*message.Result, error)
	SendRaw(ctx context.Context, payload any) (*message.Result, error)
	SendTemplate(ctx context.Context, templateName string, data any) (*message.Result, error)
	ProviderName() string
}

// BaseSender is the adapter contract implemented by individual providers.
type BaseSender interface {
	Send(ctx context.Context, msg message.Message) (*message.Result, error)
	SendRaw(ctx context.Context, payload any) (*message.Result, error)
	ProviderName() string
}

// Notifier is the core facade implementing Sender with middleware and template support.
type Notifier struct {
	base        BaseSender
	middlewares []middleware.Middleware
	registry    *template.Registry
	httpClient  *http.Client
	pipeline    middleware.Handler
}

// New wraps a BaseSender with optional middlewares and configurations.
func New(base BaseSender, opts ...Option) *Notifier {
	n := &Notifier{
		base:        base,
		middlewares: make([]middleware.Middleware, 0),
		registry:    template.DefaultRegistry(),
	}

	for _, opt := range opts {
		opt(n)
	}

	var baseHandler middleware.Handler
	if n.base != nil {
		baseHandler = n.base.Send
	} else {
		baseHandler = func(ctx context.Context, msg message.Message) (*message.Result, error) {
			return nil, errors.New("notifier: base sender is nil")
		}
	}

	// Build execution pipeline
	n.pipeline = middleware.Chain(n.middlewares...)(baseHandler)
	return n
}

// ProviderName returns the underlying provider's name.
func (n *Notifier) ProviderName() string {
	if n.base == nil {
		return "none"
	}
	return n.base.ProviderName()
}

// Send sends a Universal Message through the middleware pipeline.
func (n *Notifier) Send(ctx context.Context, msg message.Message) (*message.Result, error) {
	if n.pipeline == nil {
		return nil, errors.New("notifier: pipeline not initialized")
	}
	return n.pipeline(ctx, msg)
}

// SendRaw sends raw payload through the middleware pipeline.
func (n *Notifier) SendRaw(ctx context.Context, payload any) (*message.Result, error) {
	return n.Send(ctx, message.NewRaw(payload))
}

// SendTemplate renders a registered template with data and sends it.
func (n *Notifier) SendTemplate(ctx context.Context, templateName string, data any) (*message.Result, error) {
	if n.registry == nil {
		return nil, errors.New("notifier: template registry is not configured")
	}
	msg, err := n.registry.Render(templateName, data)
	if err != nil {
		return nil, fmt.Errorf("notifier: render template %q failed: %w", templateName, err)
	}
	return n.Send(ctx, msg)
}

// NewFromConfig builds a Notifier engine from a Config struct.
func NewFromConfig(cfg Config, opts ...Option) (*Notifier, error) {
	base, err := buildBaseSender(cfg)
	if err != nil {
		return nil, err
	}

	var mws []middleware.Middleware

	// Auto-configure rate limiting if specified
	if cfg.RateLimit > 0 {
		limiter := middleware.NewRateLimiter(cfg.RateLimit, int(cfg.RateLimit*2)+1)
		mws = append(mws, middleware.RateLimit(limiter))
	}

	// Auto-configure retries if specified
	if cfg.Retries > 1 {
		mws = append(mws, middleware.Retry(cfg.Retries))
	}

	allOpts := append([]Option{WithMiddleware(mws...)}, opts...)
	return New(base, allOpts...), nil
}

// NewFromEnv initializes a Notifier from environment variables:
// NOTIFY_PROVIDER: "gchat" | "lark"
// NOTIFY_ENDPOINT: Webhook URL
// NOTIFY_SECRET: Lark signing secret (optional)
// NOTIFY_TIMEOUT: "5s" (optional)
// NOTIFY_RETRIES: "3" (optional)
// NOTIFY_RATELIMIT: "5.0" (optional)
func NewFromEnv(opts ...Option) (*Notifier, error) {
	provider := os.Getenv("NOTIFY_PROVIDER")
	if provider == "" {
		provider = ProviderGChat
	}

	retries, _ := strconv.Atoi(os.Getenv("NOTIFY_RETRIES"))
	rateLimit, _ := strconv.ParseFloat(os.Getenv("NOTIFY_RATELIMIT"), 64)

	cfg := Config{
		Provider:  provider,
		Endpoint:  os.Getenv("NOTIFY_ENDPOINT"),
		Secret:    os.Getenv("NOTIFY_SECRET"),
		Timeout:   os.Getenv("NOTIFY_TIMEOUT"),
		Retries:   retries,
		RateLimit: rateLimit,
	}

	return NewFromConfig(cfg, opts...)
}

func buildBaseSender(cfg Config) (BaseSender, error) {
	var timeout time.Duration
	if cfg.Timeout != "" {
		var err error
		timeout, err = time.ParseDuration(cfg.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout format %q: %w", cfg.Timeout, err)
		}
	}

	switch cfg.Provider {
	case ProviderGChat:
		opts := []gchat.Option{gchat.WithWebhookURL(cfg.Endpoint)}
		if timeout > 0 {
			opts = append(opts, gchat.WithTimeout(timeout))
		}
		return gchat.New(opts...)

	case ProviderLark:
		opts := []lark.Option{
			lark.WithWebhookURL(cfg.Endpoint),
			lark.WithSecret(cfg.Secret),
		}
		if timeout > 0 {
			opts = append(opts, lark.WithTimeout(timeout))
		}
		return lark.New(opts...)

	case ProviderBroadcast:
		if len(cfg.Children) == 0 {
			return nil, errors.New("broadcast provider requires children configurations")
		}
		senders := make([]router.BaseSender, 0, len(cfg.Children))
		for _, child := range cfg.Children {
			s, err := buildBaseSender(child)
			if err != nil {
				return nil, err
			}
			senders = append(senders, s)
		}
		return router.NewBroadcast(senders...)

	case ProviderFallback:
		if len(cfg.Children) < 2 {
			return nil, errors.New("fallback provider requires at least 2 children configurations (1 primary + 1 fallback)")
		}
		primary, err := buildBaseSender(cfg.Children[0])
		if err != nil {
			return nil, err
		}
		fallbacks := make([]router.BaseSender, 0, len(cfg.Children)-1)
		for _, child := range cfg.Children[1:] {
			s, err := buildBaseSender(child)
			if err != nil {
				return nil, err
			}
			fallbacks = append(fallbacks, s)
		}
		return router.NewFallback(primary, fallbacks...)

	default:
		return nil, fmt.Errorf("unknown provider %q (supported: %s, %s, %s, %s)",
			cfg.Provider, ProviderGChat, ProviderLark, ProviderBroadcast, ProviderFallback)
	}
}

// RegisterTemplate registers a template to the default global registry.
func RegisterTemplate(name string, tpl template.Template) error {
	return template.Register(name, tpl)
}

// RegisterFunc registers a type-safe generic function template to the default global registry.
func RegisterFunc[T any](name string, fn func(data T) *message.Card) error {
	return template.RegisterFunc(name, fn)
}
