package testutil_test

import (
	"context"
	"errors"
	"testing"

	"github.com/namth2302/go-notify/message"
	"github.com/namth2302/go-notify/template"
	"github.com/namth2302/go-notify/testutil"
)

func TestMockSenderAndRecorder(t *testing.T) {
	mock := testutil.NewMockSender().
		SetProviderName("test-mock").
		SetStatusCode(200)
	recorder := testutil.NewRecorder(mock)

	if mock.ProviderName() != "test-mock" {
		t.Errorf("expected provider name 'test-mock', got %s", mock.ProviderName())
	}

	ctx := context.Background()

	// 1. Send Text
	_, err := mock.Send(ctx, message.NewText("Database connected"))
	if err != nil {
		t.Fatalf("send text failed: %v", err)
	}

	if mock.LastText() != "Database connected" {
		t.Errorf("expected last text 'Database connected', got %q", mock.LastText())
	}

	// 2. Send Card
	card := message.NewCard().
		SetStatus(message.StatusDanger).
		SetTitle("🔥 Out of Memory").
		SetSubtitle("Service: auth").
		AddSection(message.NewSection().AddField("Pod", "auth-99"))

	_, err = mock.Send(ctx, card.Wrap())
	if err != nil {
		t.Fatalf("send card failed: %v", err)
	}

	if mock.SentCount() != 2 {
		t.Errorf("expected 2 sent messages, got %d", mock.SentCount())
	}
	if mock.LastCard() == nil || mock.LastCard().Title != "🔥 Out of Memory" {
		t.Errorf("unexpected last card: %+v", mock.LastCard())
	}
	if mock.LastText() != "🔥 Out of Memory" {
		t.Errorf("expected last text from card title, got %q", mock.LastText())
	}

	// 3. SendRaw
	_, err = mock.SendRaw(ctx, map[string]string{"raw": "test"})
	if err != nil {
		t.Fatalf("send raw failed: %v", err)
	}

	// 4. SendTemplate
	_ = template.Register("mock_tpl", template.NewCardTemplate().SetTitle("Mock Template {{ .Val }}"))
	_, err = mock.SendTemplate(ctx, "mock_tpl", map[string]string{"Val": "123"})
	if err != nil {
		t.Fatalf("send template failed: %v", err)
	}

	// Test Recorder Assertions
	if !recorder.HasTitle("🔥 Out of Memory") {
		t.Errorf("expected recorder to have title")
	}
	if !recorder.HasStatus(message.StatusDanger) {
		t.Errorf("expected recorder to have status danger")
	}
	if !recorder.ContainsText("Database connected") {
		t.Errorf("expected text 'Database connected'")
	}
	if !recorder.ContainsText("auth-99") {
		t.Errorf("expected field value 'auth-99'")
	}
	if recorder.ContainsText("non-existent") {
		t.Errorf("expected false for non-existent text")
	}

	// Test Failure mode
	mock.SetFail(true, errors.New("forced fail"))
	_, err = mock.Send(ctx, message.NewText("fail"))
	if err == nil {
		t.Errorf("expected error when mock fail is true")
	}

	// Test Mock Reset
	mock.Reset()
	if mock.SentCount() != 0 {
		t.Errorf("expected 0 after reset, got %d", mock.SentCount())
	}
	if mock.LastMessage() != nil {
		t.Errorf("expected nil last message after reset")
	}
	if mock.LastText() != "" {
		t.Errorf("expected empty last text after reset")
	}
	if mock.LastCard() != nil {
		t.Errorf("expected nil last card after reset")
	}
}
