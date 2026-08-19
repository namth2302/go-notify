package gchat

import (
	"net/http"
	"time"
)

// Option configures Google Chat client.
type Option func(*Client)

// WithWebhookURL sets the incoming webhook URL.
func WithWebhookURL(url string) Option {
	return func(c *Client) {
		c.webhookURL = url
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
