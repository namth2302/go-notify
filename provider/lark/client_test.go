package lark_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/namth2302/go-notify/message"
	"github.com/namth2302/go-notify/provider/lark"
)

func TestSigner(t *testing.T) {
	signer := lark.NewSigner("my-test-secret")
	fixedTime := time.Unix(1600000000, 0)

	ts, sign, err := signer.GenerateSignature(fixedTime)
	if err != nil {
		t.Fatalf("generate signature failed: %v", err)
	}
	if ts != "1600000000" {
		t.Errorf("expected timestamp '1600000000', got %q", ts)
	}
	if sign == "" {
		t.Errorf("expected non-empty signature")
	}

	// Empty secret returns empty sign
	emptySigner := lark.NewSigner("")
	ts, sign, err = emptySigner.GenerateSignature(fixedTime)
	if err != nil || ts != "" || sign != "" {
		t.Errorf("expected empty result for empty secret")
	}
}

func TestTransformer(t *testing.T) {
	tr := lark.NewTransformer()

	// 1. Text Message
	txtMsg := message.NewText("Lark text alert")
	payload, err := tr.Transform(txtMsg)
	if err != nil {
		t.Fatalf("transform text failed: %v", err)
	}
	if payload.MsgType != "text" || payload.Content.Text != "Lark text alert" {
		t.Errorf("unexpected text payload: %+v", payload)
	}

	// 2. Card Message
	card := message.NewCard().
		SetStatus(message.StatusDanger).
		SetTitle("Critical DB Alert").
		SetSubtitle("Production").
		AddSection(
			message.NewSection().
				SetHeader("Metrics").
				SetText("CPU high").
				AddField("CPU", "98%").
				AddFields(message.NewField("Host", "db-1").SetShort(true)).
				AddFields(message.NewField("Long Detail", "Very long text detail").SetShort(false)).
				AddButton(message.NewButton("Section Button", "https://example.com/s")),
		).
		AddAction(
			message.NewButton("Acknowledge", "https://example.com/ack").AsPrimary(),
			message.NewButton("Cancel", "https://example.com/cancel").AsDanger(),
		)

	payload, err = tr.Transform(card.Wrap())
	if err != nil {
		t.Fatalf("transform card failed: %v", err)
	}

	if payload.MsgType != "interactive" || payload.Card == nil {
		t.Fatalf("expected interactive card, got %+v", payload)
	}

	c := payload.Card
	if c.Header.Template != "red" {
		t.Errorf("expected red template header for StatusDanger, got %q", c.Header.Template)
	}
	if c.Header.Title.Content != "🚨 Critical DB Alert" {
		t.Errorf("expected title '🚨 Critical DB Alert', got %q", c.Header.Title.Content)
	}
	if len(c.Body.Elements) == 0 {
		t.Errorf("expected body elements")
	}

	// 3. Raw Message
	rawMsg := message.NewRaw(map[string]any{"msg_type": "text", "content": map[string]string{"text": "raw text"}})
	payload, err = tr.Transform(rawMsg)
	if err != nil {
		t.Fatalf("transform raw failed: %v", err)
	}
	if payload.MsgType != "text" {
		t.Errorf("expected raw payload type text, got %q", payload.MsgType)
	}
}

func TestClientSend(t *testing.T) {
	var receivedPayload lark.Payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code": 0, "msg": "success"}`))
	}))
	defer server.Close()

	client, err := lark.New(
		lark.WithWebhookURL(server.URL),
		lark.WithSecret("test-secret"),
		lark.WithTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("create client failed: %v", err)
	}

	if client.ProviderName() != "lark" {
		t.Errorf("expected provider name 'lark', got %q", client.ProviderName())
	}

	res, err := client.Send(context.Background(), message.NewText("Test Lark"))
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
	if receivedPayload.Sign == "" || receivedPayload.Timestamp == "" {
		t.Errorf("expected signed payload, got %+v", receivedPayload)
	}

	// Test SendRaw
	_, err = client.SendRaw(context.Background(), map[string]any{"msg_type": "text", "content": map[string]string{"text": "raw lark"}})
	if err != nil {
		t.Fatalf("send raw failed: %v", err)
	}

	// Test Lark API Error code (e.g. invalid signature)
	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code": 19001, "msg": "sign match fail"}`))
	}))
	defer errServer.Close()

	errClient, _ := lark.New(lark.WithWebhookURL(errServer.URL))
	_, err = errClient.Send(context.Background(), message.NewText("Fail test"))
	if err == nil {
		t.Errorf("expected error on Lark error code 19001, got nil")
	}

	// Test HTTP 500 error
	http500Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`internal server error`))
	}))
	defer http500Server.Close()

	http500Client, _ := lark.New(lark.WithWebhookURL(http500Server.URL))
	_, err = http500Client.Send(context.Background(), message.NewText("Fail 500"))
	if err == nil {
		t.Errorf("expected error on HTTP 500, got nil")
	}
}

func TestClientValidation(t *testing.T) {
	_, err := lark.New()
	if err == nil {
		t.Errorf("expected error when webhook URL is missing")
	}
}
