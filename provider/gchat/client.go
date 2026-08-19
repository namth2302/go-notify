package gchat

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/namth2302/go-notify/message"
)

const ProviderName = "gchat"

// Client is a Google Chat webhook sender.
type Client struct {
	webhookURL  string
	httpClient  *http.Client
	timeout     time.Duration
	transformer *Transformer
}

// New creates a new Google Chat client.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		timeout:     10 * time.Second,
		transformer: NewTransformer(),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.webhookURL == "" {
		return nil, errors.New("gchat: webhook URL is required")
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: c.timeout,
		}
	}

	return c, nil
}

// ProviderName returns provider identifier.
func (c *Client) ProviderName() string {
	return ProviderName
}

// Send sends a universal message to Google Chat webhook.
func (c *Client) Send(ctx context.Context, msg message.Message) (*message.Result, error) {
	payloadBytes, err := c.transformer.Transform(msg)
	if err != nil {
		return nil, fmt.Errorf("gchat: transform failed: %w", err)
	}

	return c.post(ctx, payloadBytes)
}

// SendRaw sends a raw native payload.
func (c *Client) SendRaw(ctx context.Context, payload any) (*message.Result, error) {
	return c.Send(ctx, message.NewRaw(payload))
}

func (c *Client) post(ctx context.Context, body []byte) (*message.Result, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("gchat: build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		return &message.Result{
			Provider: ProviderName,
			Duration: duration,
		}, fmt.Errorf("gchat: http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	res := &message.Result{
		Provider:    ProviderName,
		StatusCode:  resp.StatusCode,
		Duration:    duration,
		RawResponse: respBody,
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return res, fmt.Errorf("gchat: unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	return res, nil
}
