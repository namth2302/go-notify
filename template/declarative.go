package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/namth2302/go-notify/message"
)

// FieldTemplate represents a template for a Key-Value field.
type FieldTemplate struct {
	Key       string
	Value     string
	IsShort   bool
	IsBoldKey bool
}

// NewFieldTemplate creates a key-value template.
func NewFieldTemplate(key, value string) *FieldTemplate {
	return &FieldTemplate{
		Key:       key,
		Value:     value,
		IsShort:   true,
		IsBoldKey: true,
	}
}

// SetShort sets whether this field is short/columnar.
func (f *FieldTemplate) SetShort(short bool) *FieldTemplate {
	f.IsShort = short
	return f
}

// ImageTemplate represents template for an image widget.
type ImageTemplate struct {
	URL string
	Alt string
}

// ButtonTemplate represents template for an action button.
type ButtonTemplate struct {
	Text string
	URL  string
	Type message.ButtonType
}

// NewButtonTemplate creates a button template.
func NewButtonTemplate(text, url string) *ButtonTemplate {
	return &ButtonTemplate{
		Text: text,
		URL:  url,
		Type: message.ButtonTypeDefault,
	}
}

// AsPrimary sets button style to primary.
func (b *ButtonTemplate) AsPrimary() *ButtonTemplate {
	b.Type = message.ButtonTypePrimary
	return b
}

// AsDanger sets button style to danger.
func (b *ButtonTemplate) AsDanger() *ButtonTemplate {
	b.Type = message.ButtonTypeDanger
	return b
}

// SectionTemplate represents a template for a card section.
type SectionTemplate struct {
	Header   string
	Text     string
	Fields   []*FieldTemplate
	Image    *ImageTemplate
	Buttons  []*ButtonTemplate
}

// NewSectionTemplate creates a new section template.
func NewSectionTemplate() *SectionTemplate {
	return &SectionTemplate{
		Fields:  make([]*FieldTemplate, 0),
		Buttons: make([]*ButtonTemplate, 0),
	}
}

// SetHeader sets section header template.
func (s *SectionTemplate) SetHeader(header string) *SectionTemplate {
	s.Header = header
	return s
}

// SetText sets section text / markdown template.
func (s *SectionTemplate) SetText(text string) *SectionTemplate {
	s.Text = text
	return s
}

// AddField adds a key-value field template.
func (s *SectionTemplate) AddField(key, value string) *SectionTemplate {
	s.Fields = append(s.Fields, NewFieldTemplate(key, value))
	return s
}

// AddFields adds multiple field templates.
func (s *SectionTemplate) AddFields(fields ...*FieldTemplate) *SectionTemplate {
	s.Fields = append(s.Fields, fields...)
	return s
}

// SetImage sets image template.
func (s *SectionTemplate) SetImage(url, alt string) *SectionTemplate {
	s.Image = &ImageTemplate{URL: url, Alt: alt}
	return s
}

// AddButton adds button template.
func (s *SectionTemplate) AddButton(btn *ButtonTemplate) *SectionTemplate {
	s.Buttons = append(s.Buttons, btn)
	return s
}

// CardTemplate represents a declarative universal card template.
type CardTemplate struct {
	Status   message.Status
	Title    string
	Subtitle string
	Sections []*SectionTemplate
	Actions  []*ButtonTemplate
}

// NewCardTemplate creates a new declarative card template.
func NewCardTemplate() *CardTemplate {
	return &CardTemplate{
		Status:   message.StatusDefault,
		Sections: make([]*SectionTemplate, 0),
		Actions:  make([]*ButtonTemplate, 0),
	}
}

// SetStatus sets status level.
func (c *CardTemplate) SetStatus(status message.Status) *CardTemplate {
	c.Status = status
	return c
}

// SetTitle sets title template.
func (c *CardTemplate) SetTitle(title string) *CardTemplate {
	c.Title = title
	return c
}

// SetSubtitle sets subtitle template.
func (c *CardTemplate) SetSubtitle(subtitle string) *CardTemplate {
	c.Subtitle = subtitle
	return c
}

// AddSection adds section templates.
func (c *CardTemplate) AddSection(sections ...*SectionTemplate) *CardTemplate {
	c.Sections = append(c.Sections, sections...)
	return c
}

// AddAction adds button templates.
func (c *CardTemplate) AddAction(buttons ...*ButtonTemplate) *CardTemplate {
	c.Actions = append(c.Actions, buttons...)
	return c
}

// Render executes data binding on the CardTemplate to produce a message.Card.
func (c *CardTemplate) Render(data any) (message.Message, error) {
	card := message.NewCard().SetStatus(c.Status)

	var err error
	card.Title, err = renderString("title", c.Title, data)
	if err != nil {
		return nil, fmt.Errorf("render title failed: %w", err)
	}

	if c.Subtitle != "" {
		card.Subtitle, err = renderString("subtitle", c.Subtitle, data)
		if err != nil {
			return nil, fmt.Errorf("render subtitle failed: %w", err)
		}
	}

	for i, sTpl := range c.Sections {
		sec := message.NewSection()
		if sTpl.Header != "" {
			sec.Header, err = renderString(fmt.Sprintf("section[%d].header", i), sTpl.Header, data)
			if err != nil {
				return nil, err
			}
		}
		if sTpl.Text != "" {
			sec.Text, err = renderString(fmt.Sprintf("section[%d].text", i), sTpl.Text, data)
			if err != nil {
				return nil, err
			}
		}
		for j, fTpl := range sTpl.Fields {
			k, err := renderString(fmt.Sprintf("section[%d].field[%d].key", i, j), fTpl.Key, data)
			if err != nil {
				return nil, err
			}
			v, err := renderString(fmt.Sprintf("section[%d].field[%d].value", i, j), fTpl.Value, data)
			if err != nil {
				return nil, err
			}
			field := message.NewField(k, v)
			field.IsShort = fTpl.IsShort
			field.IsBoldKey = fTpl.IsBoldKey
			sec.AddFields(field)
		}
		if sTpl.Image != nil {
			imgURL, err := renderString(fmt.Sprintf("section[%d].image.url", i), sTpl.Image.URL, data)
			if err != nil {
				return nil, err
			}
			imgAlt, err := renderString(fmt.Sprintf("section[%d].image.alt", i), sTpl.Image.Alt, data)
			if err != nil {
				return nil, err
			}
			sec.SetImage(imgURL, imgAlt)
		}
		for j, bTpl := range sTpl.Buttons {
			btnText, err := renderString(fmt.Sprintf("section[%d].button[%d].text", i, j), bTpl.Text, data)
			if err != nil {
				return nil, err
			}
			btnURL, err := renderString(fmt.Sprintf("section[%d].button[%d].url", i, j), bTpl.URL, data)
			if err != nil {
				return nil, err
			}
			btn := message.NewButton(btnText, btnURL)
			btn.Type = bTpl.Type
			sec.AddButton(btn)
		}
		card.AddSection(sec)
	}

	for i, bTpl := range c.Actions {
		btnText, err := renderString(fmt.Sprintf("action[%d].text", i), bTpl.Text, data)
		if err != nil {
			return nil, err
		}
		btnURL, err := renderString(fmt.Sprintf("action[%d].url", i), bTpl.URL, data)
		if err != nil {
			return nil, err
		}
		btn := message.NewButton(btnText, btnURL)
		btn.Type = bTpl.Type
		card.AddAction(btn)
	}

	return card.Wrap(), nil
}

func renderString(name, tplText string, data any) (string, error) {
	if !strings.Contains(tplText, "{{") {
		return tplText, nil
	}
	t, err := template.New(name).Parse(tplText)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
