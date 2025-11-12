package proxy

// AnthropicMessage represents a message in Anthropic API format
type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or array of content blocks
}

// AnthropicRequest represents a request in Anthropic API format
type AnthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []AnthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	System      interface{}        `json:"system,omitempty"` // Can be string or array
}

// ProviderMessage represents a message in provider API format (OpenAI-compatible)
type ProviderMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ProviderRequest represents a request in provider API format
type ProviderRequest struct {
	Model       string           `json:"model"`
	Messages    []ProviderMessage `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
}

// ProviderChoice represents a choice in provider API response
type ProviderChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

// ProviderUsage represents usage information in provider API response
type ProviderUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// ProviderResponse represents a response from provider API
type ProviderResponse struct {
	ID      string          `json:"id"`
	Model   string          `json:"model"`
	Choices []ProviderChoice `json:"choices"`
	Usage   ProviderUsage   `json:"usage"`
	Error   interface{}     `json:"error,omitempty"`
}

// AnthropicResponse represents a response in Anthropic API format
type AnthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string      `json:"model"`
	StopReason   string      `json:"stop_reason"`
	StopSequence interface{} `json:"stop_sequence"`
	Usage        struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	} `json:"usage"`
}
