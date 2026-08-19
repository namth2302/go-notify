package testutil

import (
	"strings"

	"github.com/namth/go-notify/message"
)

// Recorder provides helper assertions and inspection over a MockSender.
type Recorder struct {
	mock *MockSender
}

// NewRecorder creates a Recorder wrapping a MockSender.
func NewRecorder(mock *MockSender) *Recorder {
	return &Recorder{mock: mock}
}

// HasTitle returns true if any sent card has the exact specified title.
func (r *Recorder) HasTitle(title string) bool {
	for _, msg := range r.mock.Messages() {
		if cardMsg, ok := msg.(*message.CardMessage); ok {
			if cardMsg.Card.Title == title {
				return true
			}
		}
	}
	return false
}

// HasStatus returns true if any sent card has the specified alert status.
func (r *Recorder) HasStatus(status message.Status) bool {
	for _, msg := range r.mock.Messages() {
		if cardMsg, ok := msg.(*message.CardMessage); ok {
			if cardMsg.Card.Status == status {
				return true
			}
		}
	}
	return false
}

// ContainsText returns true if any message contains the given substring in title or text.
func (r *Recorder) ContainsText(substr string) bool {
	for _, msg := range r.mock.Messages() {
		switch m := msg.(type) {
		case *message.TextMessage:
			if strings.Contains(m.Content, substr) {
				return true
			}
		case *message.CardMessage:
			if strings.Contains(m.Card.Title, substr) || strings.Contains(m.Card.Subtitle, substr) {
				return true
			}
			for _, sec := range m.Card.Sections {
				if strings.Contains(sec.Header, substr) || strings.Contains(sec.Text, substr) {
					return true
				}
				for _, f := range sec.Fields {
					if strings.Contains(f.Key, substr) || strings.Contains(f.Value, substr) {
						return true
					}
				}
			}
		}
	}
	return false
}
