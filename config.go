package notify

// Supported provider names.
const (
	ProviderGChat     = "gchat"
	ProviderLark      = "lark"
	ProviderBroadcast = "broadcast"
	ProviderFallback  = "fallback"
)

// Config defines the configuration for the Notification client.
// Supports standard JSON, YAML, and Mapstructure (Viper) unmarshaling.
type Config struct {
	// Provider name: "gchat", "lark", "broadcast", "fallback"
	Provider string `json:"provider" yaml:"provider" mapstructure:"provider"`
	// Endpoint is the webhook URL (for gchat or lark)
	Endpoint string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`
	// Secret is the optional signing secret (for lark HMAC-SHA256)
	Secret string `json:"secret,omitempty" yaml:"secret,omitempty" mapstructure:"secret,omitempty"`
	// Timeout duration string (e.g. "5s", "10s")
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty" mapstructure:"timeout,omitempty"`
	// Retries specifies the number of retry attempts on failure (0 or 1 = no retry)
	Retries int `json:"retries,omitempty" yaml:"retries,omitempty" mapstructure:"retries,omitempty"`
	// RateLimit is requests per second limit (e.g. 5.0). 0 means no rate limiting.
	RateLimit float64 `json:"ratelimit,omitempty" yaml:"ratelimit,omitempty" mapstructure:"ratelimit,omitempty"`
	// Children is used for multi-provider setups (broadcast or fallback)
	Children []Config `json:"children,omitempty" yaml:"children,omitempty" mapstructure:"children,omitempty"`
}
