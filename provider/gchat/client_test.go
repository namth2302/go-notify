package gchat_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/namth/go-notify/message"
	"github.com/namth/go-notify/provider/gchat"
)

func TestTransformer(t *testing.T) {
	tr := gchat.NewTransformer()

	// 1. Text Message
	txtMsg := message.NewText("Hello from unit test")
	data, err := tr.Transform(txtMsg)
	if err != nil {
		t.Fatalf("transform text failed: %v", err)
	}

	var payload gchat.Payload
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal payload failed: %v", err)
	}
	if payload.Text != "Hello from unit test" {
		t.Errorf("unexpected payload text: %q", payload.Text)
	}

	// 2. Card Message
	card := message.NewCard().
		SetStatus(message.StatusWarning).
		SetTitle("Warning Alert").
		SetSubtitle("Subtitle text").
		AddSection(
			message.NewSection().
				SetHeader("Section 1").
				SetText("Section desc").
				AddField("Key1", "Val1").
				SetImage("https://example.com/img.png", "Alt").
				AddButton(message.NewButton("Click", "https://example.com")),
		).
		AddAction(
			message.NewButton("Action 1", "https://example.com/a1").AsPrimary(),
			message.NewButton("Action 2", "https://example.com/a2").AsDanger(),
		)

	data, err = tr.Transform(card.Wrap())
	if err != nil {
		t.Fatalf("transform card failed: %v", err)
	}

	var cardPayload gchat.Payload
	if err := json.Unmarshal(data, &cardPayload); err != nil {
		t.Fatalf("unmarshal card payload failed: %v", err)
	}

	if len(cardPayload.CardsV2) != 1 {
		t.Fatalf("expected 1 cardsV2, got %d", len(cardPayload.CardsV2))
	}
	c := cardPayload.CardsV2[0].Card
	if c.Header.Title != "⚠️ Warning Alert" {
		t.Errorf("unexpected header title: %q", c.Header.Title)
	}
	if len(c.Sections) != 2 { // 1 content section + 1 action section
		t.Errorf("expected 2 sections, got %d", len(c.Sections))
	}

	// 3. Raw Message
	rawMsg := message.NewRaw(map[string]string{"custom": "data"})
	data, err = tr.Transform(rawMsg)
	if err != nil {
		t.Fatalf("transform raw failed: %v", err)
	}
	if string(data) != `{"custom":"data"}` {
		t.Errorf("unexpected raw output: %s", string(data))
	}
}

func TestClientSend(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}
		var err error
		buf := make([]byte, r.ContentLength)
		_, err = r.Body.Read(buf)
		receivedBody = buf
		_ = err

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messageId": "msg-123"}`))
	}))
	defer server.Close()

	client, err := gchat.New(
		gchat.WithWebhookURL(server.URL),
		gchat.WithTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("create client failed: %v", err)
	}

	if client.ProviderName() != "gchat" {
		t.Errorf("expected provider name 'gchat', got %q", client.ProviderName())
	}

	res, err := client.Send(context.Background(), message.NewText("Test message"))
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
	if len(receivedBody) == 0 {
		t.Errorf("expected server to receive body")
	}

	// Test SendRaw
	_, err = client.SendRaw(context.Background(), map[string]string{"text": "raw"})
	if err != nil {
		t.Fatalf("send raw failed: %v", err)
	}

	// Test Error response
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer errorServer.Close()

	errClient, _ := gchat.New(gchat.WithWebhookURL(errorServer.URL))
	_, err = errClient.Send(context.Background(), message.NewText("Fail"))
	if err == nil {
		t.Errorf("expected error on 400 response, got nil")
	}
}

func TestClientValidation(t *testing.T) {
	_, err := gchat.New()
	if err == nil {
		t.Errorf("expected error for empty webhook URL")
	}
}
