package template

import (
	"fmt"

	"github.com/namth/go-notify/message"
)

// FuncTemplate wraps a Go function that produces a Message from any data.
type FuncTemplate struct {
	fn func(data any) (message.Message, error)
}

// NewFuncTemplate creates a Template from a function.
func NewFuncTemplate(fn func(data any) (message.Message, error)) *FuncTemplate {
	return &FuncTemplate{fn: fn}
}

// Render calls the wrapped function.
func (f *FuncTemplate) Render(data any) (message.Message, error) {
	if f.fn == nil {
		return nil, fmt.Errorf("nil template function")
	}
	return f.fn(data)
}

// FromCardFunc converts a type-safe `func(T) message.Card` or `func(T) *message.Card` into a Template.
func FromCardFunc[T any](fn func(data T) *message.Card) Template {
	return NewFuncTemplate(func(raw any) (message.Message, error) {
		typed, ok := raw.(T)
		if !ok {
			return nil, fmt.Errorf("expected template data type %T, got %T", *new(T), raw)
		}
		card := fn(typed)
		if card == nil {
			return nil, fmt.Errorf("template function returned nil card")
		}
		return card.Wrap(), nil
	})
}
