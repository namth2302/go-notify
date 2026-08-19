package gchat

// Payload represents Google Chat Webhook JSON payload.
type Payload struct {
	Text    string       `json:"text,omitempty"`
	CardsV2 []CardV2Item `json:"cardsV2,omitempty"`
}

// CardV2Item wraps a Card with an optional CardId.
type CardV2Item struct {
	CardID string `json:"cardId,omitempty"`
	Card   *Card  `json:"card,omitempty"`
}

// Card represents Google Chat Card structure.
type Card struct {
	Header   *Header   `json:"header,omitempty"`
	Sections []Section `json:"sections,omitempty"`
}

// Header represents Card header.
type Header struct {
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle,omitempty"`
	ImageURL  string `json:"imageUrl,omitempty"`
	ImageType string `json:"imageType,omitempty"`
}

// Section represents a Card section.
type Section struct {
	Header                    string   `json:"header,omitempty"`
	Widgets                   []Widget `json:"widgets,omitempty"`
	Collapsible               bool     `json:"collapsible,omitempty"`
	UncollapsibleWidgetsCount int      `json:"uncollapsibleWidgetsCount,omitempty"`
}

// Widget represents a single widget inside a section.
type Widget struct {
	TextParagraph *TextParagraph `json:"textParagraph,omitempty"`
	DecoratedText *DecoratedText `json:"decoratedText,omitempty"`
	Image         *ImageWidget   `json:"image,omitempty"`
	ButtonList    *ButtonList    `json:"buttonList,omitempty"`
	Divider       *Divider       `json:"divider,omitempty"`
}

// TextParagraph displays formatted text.
type TextParagraph struct {
	Text string `json:"text"`
}

// DecoratedText displays a formatted text line with top label, bottom label, or icon.
type DecoratedText struct {
	TopLabel    string `json:"topLabel,omitempty"`
	Text        string `json:"text"`
	BottomLabel string `json:"bottomLabel,omitempty"`
	WrapText    bool   `json:"wrapText,omitempty"`
}

// ImageWidget represents an image.
type ImageWidget struct {
	ImageURL string `json:"imageUrl"`
	AltText  string `json:"altText,omitempty"`
}

// ButtonList holds a slice of buttons.
type ButtonList struct {
	Buttons []Button `json:"buttons"`
}

// Button represents a Google Chat button.
type Button struct {
	Text     string   `json:"text"`
	Color    *Color   `json:"color,omitempty"`
	OnClick  *OnClick `json:"onClick,omitempty"`
	Disabled bool     `json:"disabled,omitempty"`
}

// Color represents RGBA color in Google Chat (float 0.0 - 1.0).
type Color struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
	Alpha float64 `json:"alpha,omitempty"`
}

// OnClick defines action when button is clicked.
type OnClick struct {
	OpenLink *OpenLink `json:"openLink,omitempty"`
}

// OpenLink opens an external URL.
type OpenLink struct {
	URL string `json:"url"`
}

// Divider represents a line separator widget.
type Divider struct{}
