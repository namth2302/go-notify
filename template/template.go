package template

import "github.com/namth/go-notify/message"

// Template is the common interface implemented by all notification templates.
type Template interface {
	// Render executes template with given data and produces a concrete Message.
	Render(data any) (message.Message, error)
}
