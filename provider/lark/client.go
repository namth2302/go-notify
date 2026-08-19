package lark

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/namth/go-notify/message"
)

const ProviderName = "lark"

// Client is a Lark/Feishu webhook sender.
type Client struct {
	webhookURL  string
	secret      string
	httpClient  *http.Client
	timeout     time.Duration
	transformer *Transformer
	signer      *Signer
}

// New creates a new Lark client.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		timeout:     10 * time.Second,
		transformer: NewTransformer(),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.webhookURL == "" {
		return nil, errors.New("lark: webhook URL is required")
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: c.timeout,
		}
	}

	c.signer = NewSigner(c.secret)
	return c, nil
}

// ProviderName returns provider identifier.
func (c *Client) ProviderName() string {
	return ProviderName
}

// Send sends a universal message to Lark webhook.
func (c *Client) Send(ctx context.Context, msg message.Message) (*message.Result, error) {
	payload, err := c.transformer.Transform(msg)
	if err != nil {
		return nil, fmt.Errorf("lark: transform failed: %w", err)
	}

	// Sign payload if secret is configured
	if c.secret != "" {
		ts, sign, err := c.signer.GenerateSignature(time.Now())
		if err != nil {
			return nil, fmt.Errorf("lark: generate signature failed: %w", err)
		}
		payload.Timestamp = ts
		payload.Sign = sign
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("lark: marshal payload failed: %w", err)
	}

	return c.post(ctx, payloadBytes)
}

// SendRaw sends a raw native payload.
func (c *Client) SendRaw(ctx context.Context, payload any) (*message.Result, error) {
	return c.Send(ctx, message.NewRaw(payload))
}

type responseEnvelope struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
}

func (c *Client) post(ctx context.Context, body []byte) (*message.Result, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("lark: build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		return &message.Result{
			Provider: ProviderName,
			Duration: duration,
		}, fmt.Errorf("lark: http request failed: %w", err)
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
		return res, fmt.Errorf("lark: unexpected http status %d: %s", resp.StatusCode, string(respBody))
	}

	// Check Lark specific response error code
	var env responseEnvelope
	if err := json.Unmarshal(respBody, &env); err == nil && env.Code != 0 {
		errMsg := env.Msg
		if errMsg == "" {
			errMsg = env.Message
		}
		return res, fmt.Errorf("lark: api returned error code %d: %s", env.Code, errMsg)
	}

	return res, nil
}
