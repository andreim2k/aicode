package proxy

import (
	"fmt"
	"strings"
)

// ConvertAnthropicToProvider converts Anthropic request format to provider format
func ConvertAnthropicToProvider(anthropicReq *AnthropicRequest) (*ProviderRequest, error) {
	providerMessages := []ProviderMessage{}

	// Add system message if present
	if anthropicReq.System != nil {
		systemStr, err := extractContentString(anthropicReq.System)
		if err != nil {
			return nil, fmt.Errorf("failed to extract system message: %w", err)
		}

		if systemStr != "" {
			providerMessages = append(providerMessages, ProviderMessage{
				Role:    "system",
				Content: systemStr,
			})
		}
	}

	// Process each message and convert content
	for _, msg := range anthropicReq.Messages {
		contentStr, err := extractContentString(msg.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to extract message content: %w", err)
		}

		providerMessages = append(providerMessages, ProviderMessage{
			Role:    msg.Role,
			Content: contentStr,
		})
	}

	// Create provider request
	providerReq := &ProviderRequest{
		Model:       anthropicReq.Model,
		Messages:    providerMessages,
		MaxTokens:   anthropicReq.MaxTokens,
		Temperature: anthropicReq.Temperature,
		TopP:        anthropicReq.TopP,
	}

	return providerReq, nil
}

// extractContentString extracts text content from various formats
func extractContentString(content interface{}) (string, error) {
	var parts []string

	switch v := content.(type) {
	case string:
		return v, nil
	case []interface{}:
		// Extract text from content blocks
		for _, block := range v {
			if m, ok := block.(map[string]interface{}); ok {
				if t, ok := m["text"].(string); ok {
					parts = append(parts, t)
				}
			}
		}
		return strings.Join(parts, "\n"), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// ConvertProviderToAnthropic converts provider response format to Anthropic format
func ConvertProviderToAnthropic(providerResp *ProviderResponse) *AnthropicResponse {
	anthropicResp := &AnthropicResponse{
		ID:    fmt.Sprintf("msg_%s", providerResp.ID),
		Type:  "message",
		Role:  "assistant",
		Model: providerResp.Model,
	}

	// Extract message content
	if len(providerResp.Choices) > 0 {
		anthropicResp.Content = []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			{
				Type: "text",
				Text: providerResp.Choices[0].Message.Content,
			},
		}

		// Set stop reason
		if providerResp.Choices[0].FinishReason == "stop" {
			anthropicResp.StopReason = "end_turn"
		} else {
			anthropicResp.StopReason = "max_tokens"
		}
	}

	anthropicResp.StopSequence = nil
	anthropicResp.Usage.InputTokens = providerResp.Usage.PromptTokens
	anthropicResp.Usage.OutputTokens = providerResp.Usage.CompletionTokens
	anthropicResp.Usage.CacheCreationInputTokens = 0
	anthropicResp.Usage.CacheReadInputTokens = 0

	return anthropicResp
}
