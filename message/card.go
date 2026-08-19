package message

// Field represents a key-value pair widget (e.g. in grid or key-value list).
type Field struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	IsShort   bool   `json:"is_short"`
	IsBoldKey bool   `json:"is_bold_key"`
}

// NewField creates a key-value field.
func NewField(key, value string) *Field {
	return &Field{
		Key:       key,
		Value:     value,
		IsShort:   true,
		IsBoldKey: true,
	}
}

// SetShort sets whether this field should be rendered in a side-by-side column.
func (f *Field) SetShort(short bool) *Field {
	f.IsShort = short
	return f
}

// Image represents an image widget in a card.
type Image struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}

// Section represents a logical grouping of content in a card.
type Section struct {
	Header   string    `json:"header,omitempty"`
	Text     string    `json:"text,omitempty"`
	Fields   []*Field  `json:"fields,omitempty"`
	Image    *Image    `json:"image,omitempty"`
	Buttons  []*Button `json:"buttons,omitempty"`
}

// NewSection creates a new card section.
func NewSection() *Section {
	return &Section{
		Fields:  make([]*Field, 0),
		Buttons: make([]*Button, 0),
	}
}

// SetHeader sets the header text of the section.
func (s *Section) SetHeader(header string) *Section {
	s.Header = header
	return s
}

// SetText sets descriptive markdown / text for the section.
func (s *Section) SetText(text string) *Section {
	s.Text = text
	return s
}

// AddField adds key-value fields to the section.
func (s *Section) AddField(key, value string) *Section {
	s.Fields = append(s.Fields, NewField(key, value))
	return s
}

// AddFields adds multiple Field pointers.
func (s *Section) AddFields(fields ...*Field) *Section {
	s.Fields = append(s.Fields, fields...)
	return s
}

// SetImage sets an image for the section.
func (s *Section) SetImage(url, alt string) *Section {
	s.Image = &Image{URL: url, Alt: alt}
	return s
}

// AddButton adds an action button to the section.
func (s *Section) AddButton(btn *Button) *Section {
	s.Buttons = append(s.Buttons, btn)
	return s
}

// Card represents a rich universal card message.
type Card struct {
	Status   Status     `json:"status"`
	Title    string     `json:"title"`
	Subtitle string     `json:"subtitle,omitempty"`
	Sections []*Section `json:"sections"`
	Actions  []*Button  `json:"actions,omitempty"`
}

// NewCard initializes an empty universal card.
func NewCard() *Card {
	return &Card{
		Status:   StatusDefault,
		Sections: make([]*Section, 0),
		Actions:  make([]*Button, 0),
	}
}

// SetStatus sets alert status level (Success, Warning, Danger, Info).
func (c *Card) SetStatus(status Status) *Card {
	c.Status = status
	return c
}

// SetTitle sets the main title of the card header.
func (c *Card) SetTitle(title string) *Card {
	c.Title = title
	return c
}

// SetSubtitle sets secondary header text.
func (c *Card) SetSubtitle(subtitle string) *Card {
	c.Subtitle = subtitle
	return c
}

// AddSection adds one or more content sections.
func (c *Card) AddSection(sections ...*Section) *Card {
	c.Sections = append(c.Sections, sections...)
	return c
}

// AddAction adds bottom action buttons.
func (c *Card) AddAction(buttons ...*Button) *Card {
	c.Actions = append(c.Actions, buttons...)
	return c
}
