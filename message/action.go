package message

// ButtonType defines the visual style of a button.
type ButtonType string

const (
	// ButtonTypeDefault is standard neutral button.
	ButtonTypeDefault ButtonType = "default"
	// ButtonTypePrimary is highlighted / call-to-action button.
	ButtonTypePrimary ButtonType = "primary"
	// ButtonTypeDanger is red / destructive action button.
	ButtonTypeDanger ButtonType = "danger"
)

// Button represents an interactive or link button.
type Button struct {
	Text string     `json:"text"`
	URL  string     `json:"url"`
	Type ButtonType `json:"type"`
}

// NewButton creates a standard link button.
func NewButton(text, url string) *Button {
	return &Button{
		Text: text,
		URL:  url,
		Type: ButtonTypeDefault,
	}
}

// AsPrimary sets button style to primary.
func (b *Button) AsPrimary() *Button {
	b.Type = ButtonTypePrimary
	return b
}

// AsDanger sets button style to danger/destructive.
func (b *Button) AsDanger() *Button {
	b.Type = ButtonTypeDanger
	return b
}

// AsDefault sets button style to default.
func (b *Button) AsDefault() *Button {
	b.Type = ButtonTypeDefault
	return b
}
