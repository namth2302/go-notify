package router_test

import (
	"context"
	"errors"
	"testing"

	"github.com/namth/go-notify/message"
	"github.com/namth/go-notify/router"
)

type mockBaseSender struct {
	name       string
	shouldFail bool
	called     bool
}

func (m *mockBaseSender) Send(ctx context.Context, msg message.Message) (*message.Result, error) {
	m.called = true
	if m.shouldFail {
		return &message.Result{Provider: m.name, StatusCode: 500}, errors.New("mock fail: " + m.name)
	}
	return &message.Result{Provider: m.name, StatusCode: 200}, nil
}

func (m *mockBaseSender) SendRaw(ctx context.Context, payload any) (*message.Result, error) {
	return m.Send(ctx, message.NewRaw(payload))
}

func (m *mockBaseSender) ProviderName() string {
	return m.name
}

func TestBroadcastSender(t *testing.T) {
	s1 := &mockBaseSender{name: "s1"}
	s2 := &mockBaseSender{name: "s2"}

	b, err := router.NewBroadcast(s1, s2)
	if err != nil {
		t.Fatalf("create broadcast failed: %v", err)
	}

	if b.ProviderName() != "broadcast" {
		t.Errorf("unexpected provider name: %s", b.ProviderName())
	}

	res, err := b.Send(context.Background(), message.NewText("hello"))
	if err != nil {
		t.Fatalf("broadcast failed: %v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
	if !s1.called || !s2.called {
		t.Errorf("expected all senders to be called")
	}

	// Test SendRaw
	_, err = b.SendRaw(context.Background(), map[string]string{"type": "raw"})
	if err != nil {
		t.Fatalf("broadcast raw failed: %v", err)
	}

	// Test partial failure
	s2Fail := &mockBaseSender{name: "s2", shouldFail: true}
	bFail, _ := router.NewBroadcast(s1, s2Fail)
	_, err = bFail.Send(context.Background(), message.NewText("hello"))
	if err == nil {
		t.Errorf("expected error when one sender fails")
	}

	// Test validation
	_, err = router.NewBroadcast()
	if err == nil {
		t.Errorf("expected error for empty senders")
	}
}

func TestFallbackSender(t *testing.T) {
	s1 := &mockBaseSender{name: "primary", shouldFail: true}
	s2 := &mockBaseSender{name: "fallback1", shouldFail: false}

	fb, err := router.NewFallback(s1, s2)
	if err != nil {
		t.Fatalf("create fallback failed: %v", err)
	}

	if fb.ProviderName() != "fallback" {
		t.Errorf("unexpected provider name: %s", fb.ProviderName())
	}

	res, err := fb.Send(context.Background(), message.NewText("hello"))
	if err != nil {
		t.Fatalf("fallback send failed: %v", err)
	}
	if res.Provider != "fallback1" {
		t.Errorf("expected result from fallback1, got %s", res.Provider)
	}
	if !s1.called || !s2.called {
		t.Errorf("expected both primary and fallback to be called")
	}

	// Test all fail
	s2Fail := &mockBaseSender{name: "fallback1", shouldFail: true}
	fbAllFail, _ := router.NewFallback(s1, s2Fail)
	_, err = fbAllFail.Send(context.Background(), message.NewText("hello"))
	if err == nil {
		t.Errorf("expected error when all senders fail")
	}

	// Test nil primary
	_, err = router.NewFallback(nil)
	if err == nil {
		t.Errorf("expected error for nil primary")
	}
}
