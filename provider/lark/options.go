package lark

import (
	"net/http"
	"time"
)

// Option configures Lark client.
type Option func(*Client)

// WithWebhookURL sets the Lark incoming webhook URL.
func WithWebhookURL(url string) Option {
	return func(c *Client) {
		c.webhookURL = url
	}
}

// WithSecret sets the HMAC-SHA256 signing secret.
func WithSecret(secret string) Option {
	return func(c *Client) {
		c.secret = secret
	}
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets HTTP request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}
