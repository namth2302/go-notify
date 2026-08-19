package gchat

import (
	"encoding/json"
	"fmt"

	"github.com/namth/go-notify/message"
)

// Transformer transforms universal message to Google Chat payload.
type Transformer struct{}

// NewTransformer creates a new GChat transformer.
func NewTransformer() *Transformer {
	return &Transformer{}
}

// Transform converts a message.Message into JSON bytes.
func (t *Transformer) Transform(msg message.Message) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	switch msg.Type() {
	case message.TypeText:
		txt, ok := msg.(*message.TextMessage)
		if !ok {
			return nil, fmt.Errorf("invalid text message type %T", msg)
		}
		return json.Marshal(&Payload{Text: txt.Content})

	case message.TypeCard:
		cardMsg, ok := msg.(*message.CardMessage)
		if !ok {
			return nil, fmt.Errorf("invalid card message type %T", msg)
		}
		payload := t.TransformCard(cardMsg.Card)
		return json.Marshal(payload)

	case message.TypeRaw:
		raw, ok := msg.(*message.RawMessage)
		if !ok {
			return nil, fmt.Errorf("invalid raw message type %T", msg)
		}
		if bytes, ok := raw.Payload.([]byte); ok {
			return bytes, nil
		}
		return json.Marshal(raw.Payload)

	default:
		return nil, fmt.Errorf("unsupported message type %q", msg.Type())
	}
}

// TransformCard converts universal Card into Google Chat Payload.
func (t *Transformer) TransformCard(card *message.Card) *Payload {
	if card == nil {
		return &Payload{}
	}

	titleWithEmoji := card.Title
	if card.Status != message.StatusDefault && card.Status != "" {
		titleWithEmoji = fmt.Sprintf("%s %s", card.Status.IconEmoji(), card.Title)
	}

	gchatCard := &Card{
		Header: &Header{
			Title:    titleWithEmoji,
			Subtitle: card.Subtitle,
		},
		Sections: make([]Section, 0),
	}

	for _, sec := range card.Sections {
		gsec := Section{
			Header:  sec.Header,
			Widgets: make([]Widget, 0),
		}

		if sec.Text != "" {
			gsec.Widgets = append(gsec.Widgets, Widget{
				TextParagraph: &TextParagraph{Text: sec.Text},
			})
		}

		for _, field := range sec.Fields {
			gsec.Widgets = append(gsec.Widgets, Widget{
				DecoratedText: &DecoratedText{
					TopLabel: field.Key,
					Text:     field.Value,
					WrapText: true,
				},
			})
		}

		if sec.Image != nil && sec.Image.URL != "" {
			gsec.Widgets = append(gsec.Widgets, Widget{
				Image: &ImageWidget{
					ImageURL: sec.Image.URL,
					AltText:  sec.Image.Alt,
				},
			})
		}

		if len(sec.Buttons) > 0 {
			gsec.Widgets = append(gsec.Widgets, Widget{
				ButtonList: &ButtonList{
					Buttons: t.transformButtons(sec.Buttons),
				},
			})
		}

		gchatCard.Sections = append(gchatCard.Sections, gsec)
	}

	// Bottom action buttons
	if len(card.Actions) > 0 {
		actionSec := Section{
			Widgets: []Widget{
				{
					ButtonList: &ButtonList{
						Buttons: t.transformButtons(card.Actions),
					},
				},
			},
		}
		gchatCard.Sections = append(gchatCard.Sections, actionSec)
	}

	return &Payload{
		CardsV2: []CardV2Item{
			{
				CardID: "card-1",
				Card:   gchatCard,
			},
		},
	}
}

func (t *Transformer) transformButtons(buttons []*message.Button) []Button {
	gbuttons := make([]Button, 0, len(buttons))
	for _, b := range buttons {
		btn := Button{
			Text: b.Text,
			OnClick: &OnClick{
				OpenLink: &OpenLink{URL: b.URL},
			},
		}

		switch b.Type {
		case message.ButtonTypePrimary:
			btn.Color = &Color{Red: 0.1, Green: 0.45, Blue: 0.9, Alpha: 1.0} // Blue
		case message.ButtonTypeDanger:
			btn.Color = &Color{Red: 0.85, Green: 0.15, Blue: 0.15, Alpha: 1.0} // Red
		}

		gbuttons = append(gbuttons, btn)
	}
	return gbuttons
}
