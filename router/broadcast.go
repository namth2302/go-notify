package router

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/namth2302/go-notify/message"
)

// BaseSender is the minimum interface needed for router dispatching.
type BaseSender interface {
	Send(ctx context.Context, msg message.Message) (*message.Result, error)
	SendRaw(ctx context.Context, payload any) (*message.Result, error)
	ProviderName() string
}

// BroadcastSender sends notifications to multiple providers concurrently.
type BroadcastSender struct {
	senders []BaseSender
}

// NewBroadcast creates a new BroadcastSender.
func NewBroadcast(senders ...BaseSender) (*BroadcastSender, error) {
	if len(senders) == 0 {
		return nil, errors.New("router: broadcast requires at least one sender")
	}
	return &BroadcastSender{senders: senders}, nil
}

// ProviderName returns "broadcast".
func (b *BroadcastSender) ProviderName() string {
	return "broadcast"
}

// Send sends message to all underlying senders in parallel.
func (b *BroadcastSender) Send(ctx context.Context, msg message.Message) (*message.Result, error) {
	start := time.Now()
	type sendOutput struct {
		res *message.Result
		err error
	}

	results := make([]sendOutput, len(b.senders))
	var wg sync.WaitGroup

	for i, sender := range b.senders {
		wg.Add(1)
		go func(idx int, s BaseSender) {
			defer wg.Done()
			res, err := s.Send(ctx, msg)
			results[idx] = sendOutput{res: res, err: err}
		}(i, sender)
	}

	wg.Wait()

	var errs []string
	var successCount int

	for _, out := range results {
		if out.err != nil {
			errs = append(errs, out.err.Error())
		} else {
			successCount++
		}
	}

	duration := time.Since(start)

	if len(errs) > 0 {
		return &message.Result{
			Provider: b.ProviderName(),
			Duration: duration,
		}, fmt.Errorf("broadcast errors (%d/%d succeeded): %s", successCount, len(b.senders), strings.Join(errs, "; "))
	}

	return &message.Result{
		Provider:   b.ProviderName(),
		StatusCode: 200,
		Duration:   duration,
	}, nil
}

// SendRaw broadcasts raw payload.
func (b *BroadcastSender) SendRaw(ctx context.Context, payload any) (*message.Result, error) {
	return b.Send(ctx, message.NewRaw(payload))
}
