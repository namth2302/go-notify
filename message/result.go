package message

import "time"

// Result holds the response metadata after sending a notification.
type Result struct {
	// Provider is the identifier of the sending provider (e.g. "gchat", "lark").
	Provider string `json:"provider"`
	// StatusCode is the HTTP status code returned by the provider API.
	StatusCode int `json:"status_code"`
	// MessageID is the unique ID returned by the provider (if available).
	MessageID string `json:"message_id,omitempty"`
	// Duration is the latency of the send operation.
	Duration time.Duration `json:"duration"`
	// RawResponse is the raw response body from the provider.
	RawResponse []byte `json:"raw_response,omitempty"`
}
