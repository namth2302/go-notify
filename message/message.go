package message

// Type defines the message format type.
type Type string

const (
	// TypeCard indicates a rich Universal Card message.
	TypeCard Type = "card"
	// TypeText indicates a plain text / simple markdown message.
	TypeText Type = "text"
	// TypeRaw indicates raw platform-specific JSON payload.
	TypeRaw Type = "raw"
)

// Message is the common interface implemented by all message types.
type Message interface {
	Type() Type
}

// CardMessage wraps a rich Universal Card.
type CardMessage struct {
	Card *Card
}

// Type returns TypeCard.
func (c *CardMessage) Type() Type {
	return TypeCard
}

// Wrap returns Message interface for Card.
func (c *Card) Wrap() Message {
	return &CardMessage{Card: c}
}

// TextMessage represents a simple text or markdown notification.
type TextMessage struct {
	Content string `json:"content"`
}

// NewText creates a plain text message.
func NewText(content string) *TextMessage {
	return &TextMessage{Content: content}
}

// Type returns TypeText.
func (t *TextMessage) Type() Type {
	return TypeText
}

// RawMessage represents a native payload bypass (escape hatch).
type RawMessage struct {
	Payload any
}

// NewRaw creates a raw message.
func NewRaw(payload any) *RawMessage {
	return &RawMessage{Payload: payload}
}

// Type returns TypeRaw.
func (r *RawMessage) Type() Type {
	return TypeRaw
}
