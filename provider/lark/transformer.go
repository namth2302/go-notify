package lark

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/namth/go-notify/message"
)

// Transformer transforms universal message to Lark Interactive Card payload.
type Transformer struct{}

// NewTransformer creates a new Lark transformer.
func NewTransformer() *Transformer {
	return &Transformer{}
}

// Transform converts a message.Message into a Lark Payload struct.
func (t *Transformer) Transform(msg message.Message) (*Payload, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	switch msg.Type() {
	case message.TypeText:
		txt, ok := msg.(*message.TextMessage)
		if !ok {
			return nil, fmt.Errorf("invalid text message type %T", msg)
		}
		return &Payload{
			MsgType: "text",
			Content: &TextContent{Text: txt.Content},
		}, nil

	case message.TypeCard:
		cardMsg, ok := msg.(*message.CardMessage)
		if !ok {
			return nil, fmt.Errorf("invalid card message type %T", msg)
		}
		card := t.TransformCard(cardMsg.Card)
		return &Payload{
			MsgType: "interactive",
			Card:    card,
		}, nil

	case message.TypeRaw:
		raw, ok := msg.(*message.RawMessage)
		if !ok {
			return nil, fmt.Errorf("invalid raw message type %T", msg)
		}
		if p, ok := raw.Payload.(*Payload); ok {
			return p, nil
		}
		// If raw map or struct, marshal and unmarshal into generic payload
		bytes, err := json.Marshal(raw.Payload)
		if err != nil {
			return nil, err
		}
		var p Payload
		if err := json.Unmarshal(bytes, &p); err == nil && p.MsgType != "" {
			return &p, nil
		}
		// Default fallback for raw
		return &Payload{
			MsgType: "interactive",
			Card: &Card{
				Schema: "2.0",
				Body: &CardBody{
					Elements: []Element{
						{
							Tag:     "markdown",
							Content: string(bytes),
						},
					},
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported message type %q", msg.Type())
	}
}

// TransformCard converts universal Card into Lark Card 2.0.
func (t *Transformer) TransformCard(card *message.Card) *Card {
	if card == nil {
		return &Card{Schema: "2.0"}
	}

	templateColor := card.Status.LarkTemplate()
	titleWithEmoji := card.Title
	if card.Status != message.StatusDefault && card.Status != "" {
		titleWithEmoji = fmt.Sprintf("%s %s", card.Status.IconEmoji(), card.Title)
	}

	larkHeader := &CardHeader{
		Template: templateColor,
		Title: &TextNode{
			Tag:     "plain_text",
			Content: titleWithEmoji,
		},
	}

	if card.Subtitle != "" {
		larkHeader.Subtitle = &TextNode{
			Tag:     "plain_text",
			Content: card.Subtitle,
		}
	}

	body := &CardBody{
		Elements: make([]Element, 0),
	}

	for _, sec := range card.Sections {
		if sec.Header != "" {
			body.Elements = append(body.Elements, Element{
				Tag:     "markdown",
				Content: fmt.Sprintf("**%s**", sec.Header),
			})
		}

		if sec.Text != "" {
			body.Elements = append(body.Elements, Element{
				Tag:     "markdown",
				Content: sec.Text,
			})
		}

		if len(sec.Fields) > 0 {
			body.Elements = append(body.Elements, t.transformFields(sec.Fields)...)
		}

		if len(sec.Buttons) > 0 {
			body.Elements = append(body.Elements, Element{
				Tag:     "action",
				Actions: t.transformActions(sec.Buttons),
			})
		}
	}

	// Bottom action buttons
	if len(card.Actions) > 0 {
		body.Elements = append(body.Elements, Element{
			Tag:     "action",
			Actions: t.transformActions(card.Actions),
		})
	}

	return &Card{
		Schema: "2.0",
		Header: larkHeader,
		Body:   body,
	}
}

func (t *Transformer) transformFields(fields []*message.Field) []Element {
	elements := make([]Element, 0)

	// Group short fields into pairs (2-column grid)
	var shortGroup []*message.Field
	for _, f := range fields {
		if f.IsShort {
			shortGroup = append(shortGroup, f)
			if len(shortGroup) == 2 {
				elements = append(elements, t.createColumnSet(shortGroup))
				shortGroup = nil
			}
		} else {
			if len(shortGroup) > 0 {
				elements = append(elements, t.createColumnSet(shortGroup))
				shortGroup = nil
			}
			elements = append(elements, Element{
				Tag:     "markdown",
				Content: formatFieldMarkdown(f),
			})
		}
	}
	if len(shortGroup) > 0 {
		elements = append(elements, t.createColumnSet(shortGroup))
	}

	return elements
}

func (t *Transformer) createColumnSet(fields []*message.Field) Element {
	cols := make([]Column, 0, len(fields))
	for _, f := range fields {
		cols = append(cols, Column{
			Tag:    "column",
			Width:  "weighted",
			Weight: 1,
			Elements: []Element{
				{
					Tag:     "markdown",
					Content: formatFieldMarkdown(f),
				},
			},
		})
	}
	return Element{
		Tag:             "column_set",
		FlexMode:        "none",
		BackgroundStyle: "default",
		Columns:         cols,
	}
}

func formatFieldMarkdown(f *message.Field) string {
	var sb strings.Builder
	if f.IsBoldKey {
		sb.WriteString("**")
		sb.WriteString(f.Key)
		sb.WriteString(":**\n")
	} else {
		sb.WriteString(f.Key)
		sb.WriteString(":\n")
	}
	sb.WriteString(f.Value)
	return sb.String()
}

func (t *Transformer) transformActions(buttons []*message.Button) []ActionItem {
	actions := make([]ActionItem, 0, len(buttons))
	for _, b := range buttons {
		btnType := "default"
		switch b.Type {
		case message.ButtonTypePrimary:
			btnType = "primary"
		case message.ButtonTypeDanger:
			btnType = "danger"
		}

		actions = append(actions, ActionItem{
			Tag:  "button",
			Type: btnType,
			Text: &TextNode{
				Tag:     "plain_text",
				Content: b.Text,
			},
			URL: b.URL,
		})
	}
	return actions
}
