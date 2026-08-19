package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	notify "github.com/namth2302/go-notify"
	"github.com/namth2302/go-notify/message"
	"github.com/namth2302/go-notify/middleware"
	"github.com/namth2302/go-notify/template"
	"github.com/namth2302/go-notify/testutil"
)

func TestNotifierSendWithMiddleware(t *testing.T) {
	mock := testutil.NewMockSender().SetProviderName("gchat")

	var mwExecuted bool
	customMw := func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, msg message.Message) (*message.Result, error) {
			mwExecuted = true
			return next(ctx, msg)
		}
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}
	notifier := notify.New(mock,
		notify.WithMiddleware(customMw),
		notify.WithHTTPClient(httpClient),
	)

	if notifier.ProviderName() != "gchat" {
		t.Errorf("expected provider gchat, got %s", notifier.ProviderName())
	}

	res, err := notifier.Send(context.Background(), message.NewText("Hello from root"))
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}
	if !mwExecuted {
		t.Errorf("expected middleware to execute")
	}
	if mock.SentCount() != 1 {
		t.Errorf("expected mock sent count 1, got %d", mock.SentCount())
	}
}

func TestNotifierSendTemplate(t *testing.T) {
	mock := testutil.NewMockSender()
	reg := template.NewRegistry()

	_ = reg.Register("welcome", template.NewCardTemplate().
		SetTitle("Welcome {{ .Username }}").
		AddSection(template.NewSectionTemplate().AddField("Role", "{{ .Role }}")),
	)

	notifier := notify.New(mock, notify.WithRegistry(reg))

	res, err := notifier.SendTemplate(context.Background(), "welcome", map[string]string{
		"Username": "Alice",
		"Role":     "Admin",
	})
	if err != nil {
		t.Fatalf("send template failed: %v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}

	lastCard := mock.LastCard()
	if lastCard == nil || lastCard.Title != "Welcome Alice" {
		t.Errorf("unexpected card title: %+v", lastCard)
	}
	if len(lastCard.Sections) == 0 || lastCard.Sections[0].Fields[0].Value != "Admin" {
		t.Errorf("unexpected section field value")
	}

	// Test missing template
	_, err = notifier.SendTemplate(context.Background(), "non_existent", nil)
	if err == nil {
		t.Errorf("expected error for non existent template")
	}
}

func TestNewFromConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code": 0}`))
	}))
	defer server.Close()

	// 1. GChat Config
	gchatCfg := notify.Config{
		Provider:  notify.ProviderGChat,
		Endpoint:  server.URL,
		Timeout:   "3s",
		Retries:   2,
		RateLimit: 10,
	}
	nGChat, err := notify.NewFromConfig(gchatCfg)
	if err != nil {
		t.Fatalf("new gchat failed: %v", err)
	}
	if nGChat.ProviderName() != "gchat" {
		t.Errorf("expected provider gchat, got %s", nGChat.ProviderName())
	}

	// 2. Lark Config
	larkCfg := notify.Config{
		Provider: notify.ProviderLark,
		Endpoint: server.URL,
		Secret:   "test-secret",
		Timeout:  "5s",
	}
	nLark, err := notify.NewFromConfig(larkCfg)
	if err != nil {
		t.Fatalf("new lark failed: %v", err)
	}
	if nLark.ProviderName() != "lark" {
		t.Errorf("expected provider lark, got %s", nLark.ProviderName())
	}

	// 3. Broadcast Config
	broadcastCfg := notify.Config{
		Provider: notify.ProviderBroadcast,
		Children: []notify.Config{gchatCfg, larkCfg},
	}
	nBroadcast, err := notify.NewFromConfig(broadcastCfg)
	if err != nil {
		t.Fatalf("new broadcast failed: %v", err)
	}
	if nBroadcast.ProviderName() != "broadcast" {
		t.Errorf("expected provider broadcast, got %s", nBroadcast.ProviderName())
	}

	// 4. Fallback Config
	fallbackCfg := notify.Config{
		Provider: notify.ProviderFallback,
		Children: []notify.Config{gchatCfg, larkCfg},
	}
	nFallback, err := notify.NewFromConfig(fallbackCfg)
	if err != nil {
		t.Fatalf("new fallback failed: %v", err)
	}
	if nFallback.ProviderName() != "fallback" {
		t.Errorf("expected provider fallback, got %s", nFallback.ProviderName())
	}

	// 5. Unknown Provider
	_, err = notify.NewFromConfig(notify.Config{Provider: "unknown"})
	if err == nil {
		t.Errorf("expected error for unknown provider")
	}

	// 6. Invalid Timeout
	_, err = notify.NewFromConfig(notify.Config{Provider: notify.ProviderGChat, Timeout: "invalid"})
	if err == nil {
		t.Errorf("expected error for invalid timeout format")
	}

	// 7. Invalid broadcast children
	_, err = notify.NewFromConfig(notify.Config{Provider: notify.ProviderBroadcast})
	if err == nil {
		t.Errorf("expected error for broadcast without children")
	}

	// 8. Invalid fallback children
	_, err = notify.NewFromConfig(notify.Config{Provider: notify.ProviderFallback})
	if err == nil {
		t.Errorf("expected error for fallback without children")
	}
}

func TestNewFromEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = os.Setenv("NOTIFY_PROVIDER", "lark")
	_ = os.Setenv("NOTIFY_ENDPOINT", server.URL)
	_ = os.Setenv("NOTIFY_SECRET", "secret")
	_ = os.Setenv("NOTIFY_RETRIES", "2")
	_ = os.Setenv("NOTIFY_RATELIMIT", "5.0")
	defer func() {
		_ = os.Unsetenv("NOTIFY_PROVIDER")
		_ = os.Unsetenv("NOTIFY_ENDPOINT")
		_ = os.Unsetenv("NOTIFY_SECRET")
		_ = os.Unsetenv("NOTIFY_RETRIES")
		_ = os.Unsetenv("NOTIFY_RATELIMIT")
	}()

	n, err := notify.NewFromEnv()
	if err != nil {
		t.Fatalf("new from env failed: %v", err)
	}
	if n.ProviderName() != "lark" {
		t.Errorf("expected provider lark, got %s", n.ProviderName())
	}
}

func TestGlobalTemplateHelpers(t *testing.T) {
	_ = notify.RegisterTemplate("global_card_tpl", template.NewCardTemplate().SetTitle("Global Card"))

	var executed int32
	_ = notify.RegisterFunc("test_helper_func", func(name string) *message.Card {
		atomic.AddInt32(&executed, 1)
		return message.NewCard().SetTitle("Hi " + name)
	})

	mock := testutil.NewMockSender()
	n := notify.New(mock)

	_, err := n.SendTemplate(context.Background(), "test_helper_func", "Bob")
	if err != nil {
		t.Fatalf("send template failed: %v", err)
	}
	if atomic.LoadInt32(&executed) != 1 {
		t.Errorf("expected template func to execute")
	}

	// SendRaw test
	_, err = n.SendRaw(context.Background(), map[string]string{"k": "v"})
	if err != nil {
		t.Fatalf("send raw failed: %v", err)
	}

	// Nil Base test
	nilNotifier := notify.New(nil)
	if nilNotifier.ProviderName() != "none" {
		t.Errorf("expected 'none', got %s", nilNotifier.ProviderName())
	}
}
