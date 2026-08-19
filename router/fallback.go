package router

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/namth/go-notify/message"
)

// FallbackSender tries primary sender first, and falls back to secondary senders upon failure.
type FallbackSender struct {
	primary   BaseSender
	fallbacks []BaseSender
}

// NewFallback creates a new FallbackSender.
func NewFallback(primary BaseSender, fallbacks ...BaseSender) (*FallbackSender, error) {
	if primary == nil {
		return nil, errors.New("router: primary sender cannot be nil")
	}
	return &FallbackSender{
		primary:   primary,
		fallbacks: fallbacks,
	}, nil
}

// ProviderName returns "fallback".
func (f *FallbackSender) ProviderName() string {
	return "fallback"
}

// Send attempts to send with primary, falling back to next available sender on failure.
func (f *FallbackSender) Send(ctx context.Context, msg message.Message) (*message.Result, error) {
	start := time.Now()

	res, err := f.primary.Send(ctx, msg)
	if err == nil {
		return res, nil
	}

	errs := []string{fmt.Sprintf("primary (%s) failed: %v", f.primary.ProviderName(), err)}

	for _, fb := range f.fallbacks {
		fbRes, fbErr := fb.Send(ctx, msg)
		if fbErr == nil {
			return fbRes, nil
		}
		errs = append(errs, fmt.Sprintf("fallback (%s) failed: %v", fb.ProviderName(), fbErr))
	}

	return &message.Result{
		Provider: f.ProviderName(),
		Duration: time.Since(start),
	}, fmt.Errorf("all senders failed: %s", strings.Join(errs, "; "))
}

// SendRaw sends raw payload with fallback.
func (f *FallbackSender) SendRaw(ctx context.Context, payload any) (*message.Result, error) {
	return f.Send(ctx, message.NewRaw(payload))
}
