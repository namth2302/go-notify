package message_test

import (
	"testing"

	"github.com/namth/go-notify/message"
)

func TestCardBuilder(t *testing.T) {
	card := message.NewCard().
		SetStatus(message.StatusDanger).
		SetTitle("🚨 High CPU Load").
		SetSubtitle("Host: node-1").
		AddSection(
			message.NewSection().
				SetHeader("Metrics").
				SetText("CPU is above 90% threshold.").
				AddField("Current", "94.5%").
				AddFields(message.NewField("Threshold", "85%").SetShort(false)).
				SetImage("https://example.com/chart.png", "CPU Chart").
				AddButton(message.NewButton("Section Link", "https://example.com/s")),
		).
		AddAction(
			message.NewButton("Dashboard", "https://example.com/dash").AsPrimary(),
			message.NewButton("Dismiss", "https://example.com/dismiss").AsDanger(),
			message.NewButton("Default", "https://example.com/def").AsDefault(),
		)

	if card.Status != message.StatusDanger {
		t.Errorf("expected StatusDanger, got %v", card.Status)
	}
	if card.Title != "🚨 High CPU Load" {
		t.Errorf("expected title '🚨 High CPU Load', got %v", card.Title)
	}
	if len(card.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(card.Sections))
	}
	if len(card.Sections[0].Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(card.Sections[0].Fields))
	}
	if len(card.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(card.Actions))
	}
	if card.Actions[0].Type != message.ButtonTypePrimary {
		t.Errorf("expected first action to be primary, got %v", card.Actions[0].Type)
	}
	if card.Actions[1].Type != message.ButtonTypeDanger {
		t.Errorf("expected second action to be danger, got %v", card.Actions[1].Type)
	}

	msg := card.Wrap()
	if msg.Type() != message.TypeCard {
		t.Errorf("expected TypeCard, got %v", msg.Type())
	}
}

func TestStatusHelpers(t *testing.T) {
	tests := []struct {
		status       message.Status
		wantTemplate string
		wantEmoji    string
	}{
		{message.StatusSuccess, "green", "✅"},
		{message.StatusWarning, "orange", "⚠️"},
		{message.StatusDanger, "red", "🚨"},
		{message.StatusInfo, "blue", "ℹ️"},
		{message.StatusDefault, "grey", "📢"},
	}

	for _, tt := range tests {
		if got := tt.status.LarkTemplate(); got != tt.wantTemplate {
			t.Errorf("status %v: got template %q, want %q", tt.status, got, tt.wantTemplate)
		}
		if got := tt.status.IconEmoji(); got != tt.wantEmoji {
			t.Errorf("status %v: got emoji %q, want %q", tt.status, got, tt.wantEmoji)
		}
	}
}

func TestTextAndRawMessage(t *testing.T) {
	txt := message.NewText("hello world")
	if txt.Type() != message.TypeText || txt.Content != "hello world" {
		t.Errorf("unexpected text message: %+v", txt)
	}

	raw := message.NewRaw(map[string]string{"foo": "bar"})
	if raw.Type() != message.TypeRaw {
		t.Errorf("unexpected raw message type: %v", raw.Type())
	}
}
