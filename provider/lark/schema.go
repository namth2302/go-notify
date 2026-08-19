package lark

// Payload represents Lark Webhook JSON request envelope.
type Payload struct {
	Timestamp string       `json:"timestamp,omitempty"`
	Sign      string       `json:"sign,omitempty"`
	MsgType   string       `json:"msg_type"`
	Content   *TextContent `json:"content,omitempty"`
	Card      *Card        `json:"card,omitempty"`
}

// TextContent is the payload structure for msg_type="text".
type TextContent struct {
	Text string `json:"text"`
}

// Card represents Lark Card 2.0 interactive structure.
type Card struct {
	Schema string      `json:"schema,omitempty"`
	Header *CardHeader `json:"header,omitempty"`
	Body   *CardBody   `json:"body,omitempty"`
}

// CardHeader is the top header of a Lark Card.
type CardHeader struct {
	Template string    `json:"template,omitempty"` // "blue", "green", "orange", "red", "grey", etc.
	Title    *TextNode `json:"title,omitempty"`
	Subtitle *TextNode `json:"subtitle,omitempty"`
}

// TextNode represents a text or markdown node.
type TextNode struct {
	Tag     string `json:"tag"` // "plain_text" or "lark_md"
	Content string `json:"content"`
}

// CardBody contains card elements.
type CardBody struct {
	Elements []Element `json:"elements"`
}

// Element is a generic Lark Card UI component.
type Element struct {
	Tag             string       `json:"tag"` // "markdown", "div", "column_set", "action", "hr", "img"
	Content         string       `json:"content,omitempty"`
	FlexMode        string       `json:"flex_mode,omitempty"`
	BackgroundStyle string       `json:"background_style,omitempty"`
	Columns         []Column     `json:"columns,omitempty"`
	Actions         []ActionItem `json:"actions,omitempty"`
	ImgKey          string       `json:"img_key,omitempty"`
	Alt             *TextNode    `json:"alt,omitempty"`
	Text            *TextNode    `json:"text,omitempty"`
}

// Column represents a column in a ColumnSet.
type Column struct {
	Tag      string    `json:"tag"` // "column"
	Width    string    `json:"width,omitempty"`
	Weight   int       `json:"weight,omitempty"`
	Elements []Element `json:"elements"`
}

// ActionItem represents a button in Lark Card.
type ActionItem struct {
	Tag   string    `json:"tag"`            // "button"
	Text  *TextNode `json:"text"`           // button label
	Type  string    `json:"type,omitempty"` // "default", "primary", "danger"
	URL   string    `json:"url,omitempty"`
	MultiURL *MultiURL `json:"multi_url,omitempty"`
}

// MultiURL allows specifying desktop vs mobile URLs.
type MultiURL struct {
	URL        string `json:"url"`
	AndroidURL string `json:"android_url,omitempty"`
	IOSURL     string `json:"ios_url,omitempty"`
	PCURL      string `json:"pc_url,omitempty"`
}
