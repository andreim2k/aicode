package proxy

import (
	"fmt"
	"net/http"
)

const (
	// MaxRequestBodySize is the maximum allowed request body size (10MB)
	MaxRequestBodySize = 10 * 1024 * 1024
	// MaxMessages is the maximum number of messages allowed
	MaxMessages = 100
	// MaxTokens is the maximum tokens allowed
	MaxTokens = 100000
)

// ValidateRequest validates the incoming request
func ValidateRequest(r *http.Request) error {
	// Check Content-Length header
	if r.ContentLength > MaxRequestBodySize {
		return fmt.Errorf("request body too large: %d bytes (max: %d)", r.ContentLength, MaxRequestBodySize)
	}

	return nil
}

// ValidateAnthropicRequest validates the Anthropic request structure
func ValidateAnthropicRequest(req *AnthropicRequest) error {
	// Validate model name
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	// Validate messages
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages array cannot be empty")
	}

	if len(req.Messages) > MaxMessages {
		return fmt.Errorf("too many messages: %d (max: %d)", len(req.Messages), MaxMessages)
	}

	// Validate max_tokens
	if req.MaxTokens > MaxTokens {
		return fmt.Errorf("max_tokens too large: %d (max: %d)", req.MaxTokens, MaxTokens)
	}

	if req.MaxTokens < 0 {
		return fmt.Errorf("max_tokens cannot be negative")
	}

	// Validate temperature
	if req.Temperature < 0 || req.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2, got: %f", req.Temperature)
	}

	// Validate top_p
	if req.TopP < 0 || req.TopP > 1 {
		return fmt.Errorf("top_p must be between 0 and 1, got: %f", req.TopP)
	}

	return nil
}
