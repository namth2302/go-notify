package template

import (
	"errors"
	"fmt"
	"sync"

	"github.com/namth/go-notify/message"
)

var (
	// ErrTemplateNotFound is returned when rendering an unregistered template.
	ErrTemplateNotFound = errors.New("template not found")
	// ErrTemplateAlreadyExists is returned when registering duplicate template without overwrite.
	ErrTemplateAlreadyExists = errors.New("template already exists")
)

// Registry manages thread-safe storage and retrieval of templates.
type Registry struct {
	mu        sync.RWMutex
	templates map[string]Template
}

// NewRegistry creates an empty Template Registry.
func NewRegistry() *Registry {
	return &Registry{
		templates: make(map[string]Template),
	}
}

// Register adds a template to the registry.
func (r *Registry) Register(name string, tpl Template) error {
	if name == "" {
		return errors.New("template name cannot be empty")
	}
	if tpl == nil {
		return errors.New("cannot register nil template")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.templates[name] = tpl
	return nil
}

// Get retrieves a registered template by name.
func (r *Registry) Get(name string) (Template, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tpl, ok := r.templates[name]
	return tpl, ok
}

// Render executes the named template with given data.
func (r *Registry) Render(name string, data any) (message.Message, error) {
	tpl, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
	}
	return tpl.Render(data)
}

// Global default registry
var defaultRegistry = NewRegistry()

// Register registers a template to the global default registry.
func Register(name string, tpl Template) error {
	return defaultRegistry.Register(name, tpl)
}

// RegisterFunc registers a type-safe generic function template globally.
func RegisterFunc[T any](name string, fn func(data T) *message.Card) error {
	return defaultRegistry.Register(name, FromCardFunc(fn))
}

// Get retrieves a template from the global default registry.
func Get(name string) (Template, bool) {
	return defaultRegistry.Get(name)
}

// Render renders a template from the global default registry.
func Render(name string, data any) (message.Message, error) {
	return defaultRegistry.Render(name, data)
}

// DefaultRegistry returns the singleton global registry.
func DefaultRegistry() *Registry {
	return defaultRegistry
}
