package testutil

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/namth/go-notify/message"
	"github.com/namth/go-notify/template"
)

// MockSender is a test double implementation of notifications.Sender for consumer tests.
type MockSender struct {
	mu           sync.RWMutex
	messages     []message.Message
	rawPayloads  []any
	results      []*message.Result
	providerName string
	shouldFail   bool
	failErr      error
	statusCode   int
	latency      time.Duration
	registry     *template.Registry
}

// NewMockSender creates an initialized MockSender.
func NewMockSender() *MockSender {
	return &MockSender{
		messages:     make([]message.Message, 0),
		rawPayloads:  make([]any, 0),
		results:      make([]*message.Result, 0),
		providerName: "mock",
		statusCode:   200,
		registry:     template.DefaultRegistry(),
	}
}

// ProviderName returns the mock provider name.
func (m *MockSender) ProviderName() string {
	return m.providerName
}

// SetProviderName changes provider name tag.
func (m *MockSender) SetProviderName(name string) *MockSender {
	m.providerName = name
	return m
}

// SetFail configures the mock to return an error on Send.
func (m *MockSender) SetFail(shouldFail bool, err error) *MockSender {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
	if err == nil && shouldFail {
		m.failErr = errors.New("mock send failed")
	} else {
		m.failErr = err
	}
	return m
}

// SetStatusCode configures HTTP status code returned in Result.
func (m *MockSender) SetStatusCode(code int) *MockSender {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusCode = code
	return m
}

// Send records the sent message in memory.
func (m *MockSender) Send(ctx context.Context, msg message.Message) (*message.Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.latency > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.latency):
		}
	}

	m.messages = append(m.messages, msg)

	res := &message.Result{
		Provider:   m.providerName,
		StatusCode: m.statusCode,
		Duration:   m.latency,
	}
	m.results = append(m.results, res)

	if m.shouldFail {
		return res, m.failErr
	}

	return res, nil
}

// SendRaw records raw payload in memory.
func (m *MockSender) SendRaw(ctx context.Context, payload any) (*message.Result, error) {
	m.mu.Lock()
	m.rawPayloads = append(m.rawPayloads, payload)
	m.mu.Unlock()
	return m.Send(ctx, message.NewRaw(payload))
}

// SendTemplate renders template and records sent message.
func (m *MockSender) SendTemplate(ctx context.Context, templateName string, data any) (*message.Result, error) {
	msg, err := m.registry.Render(templateName, data)
	if err != nil {
		return nil, fmt.Errorf("mock render template failed: %w", err)
	}
	return m.Send(ctx, msg)
}

// SentCount returns the total number of messages sent.
func (m *MockSender) SentCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages)
}

// Messages returns all sent messages.
func (m *MockSender) Messages() []message.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copied := make([]message.Message, len(m.messages))
	copy(copied, m.messages)
	return copied
}

// LastMessage returns the most recent message sent, or nil if none.
func (m *MockSender) LastMessage() message.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.messages) == 0 {
		return nil
	}
	return m.messages[len(m.messages)-1]
}

// LastCard returns the most recent Card, or nil if not a Card.
func (m *MockSender) LastCard() *message.Card {
	last := m.LastMessage()
	if last == nil {
		return nil
	}
	if cardMsg, ok := last.(*message.CardMessage); ok {
		return cardMsg.Card
	}
	return nil
}

// LastText returns the text content of the last message.
func (m *MockSender) LastText() string {
	last := m.LastMessage()
	if last == nil {
		return ""
	}
	if txtMsg, ok := last.(*message.TextMessage); ok {
		return txtMsg.Content
	}
	if card := m.LastCard(); card != nil {
		return card.Title
	}
	return ""
}

// Reset clears all recorded messages.
func (m *MockSender) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = m.messages[:0]
	m.rawPayloads = m.rawPayloads[:0]
	m.results = m.results[:0]
	m.shouldFail = false
	m.failErr = nil
}
