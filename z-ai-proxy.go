package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or array of content blocks
}

type AnthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []AnthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	System      interface{}        `json:"system,omitempty"` // Can be string or array
}

type Z_AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Z_AIRequest struct {
	Model       string        `json:"model"`
	Messages    []Z_AIMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
}

type Z_AIChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type Z_AIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type Z_AIResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Choices []Z_AIChoice `json:"choices"`
	Usage   Z_AIUsage    `json:"usage"`
	Error   interface{}  `json:"error,omitempty"`
}

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

var (
	port      = flag.String("port", "9000", "Port to listen on")
	z_aiURL   = flag.String("z-ai-url", "https://api.z.ai/api/paas/v4", "Z.AI base URL")
	z_aiToken = flag.String("z-ai-token", "", "Z.AI auth token")
)

func handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse Anthropic request
	var anthropicReq AnthropicRequest
	if err := json.Unmarshal(body, &anthropicReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Convert Anthropic messages to Z.AI format
	z_aiMessages := []Z_AIMessage{}

	// Add system message if present
	if anthropicReq.System != nil {
		var systemStr string
		switch v := anthropicReq.System.(type) {
		case string:
			systemStr = v
		case []interface{}:
			// Extract text from system content blocks
			for _, block := range v {
				if m, ok := block.(map[string]interface{}); ok {
					if t, ok := m["text"].(string); ok {
						systemStr += t
					}
				}
			}
		default:
			systemStr = fmt.Sprintf("%v", v)
		}

		if systemStr != "" {
			z_aiMessages = append(z_aiMessages, Z_AIMessage{
				Role:    "system",
				Content: systemStr,
			})
		}
	}

	// Process each message and convert content
	for _, msg := range anthropicReq.Messages {
		var contentStr string

		// Handle content as either string or array
		switch v := msg.Content.(type) {
		case string:
			contentStr = v
		case []interface{}:
			// Extract text from content blocks
			for _, block := range v {
				if m, ok := block.(map[string]interface{}); ok {
					if t, ok := m["text"].(string); ok {
						contentStr += t
					}
				}
			}
		default:
			contentStr = fmt.Sprintf("%v", v)
		}

		z_aiMessages = append(z_aiMessages, Z_AIMessage{
			Role:    msg.Role,
			Content: contentStr,
		})
	}

	// Create Z.AI request
	z_aiReq := Z_AIRequest{
		Model:       anthropicReq.Model,
		Messages:    z_aiMessages,
		MaxTokens:   anthropicReq.MaxTokens,
		Temperature: anthropicReq.Temperature,
		TopP:        anthropicReq.TopP,
	}

	// Marshal to JSON
	z_aiBody, err := json.Marshal(z_aiReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal request: %v", err), http.StatusInternalServerError)
		return
	}

	// Call Z.AI API
	baseURL := *z_aiURL
	if baseURL != "" && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	z_aiFullURL := baseURL + "/chat/completions"
	req, err := http.NewRequest("POST", z_aiFullURL, bytes.NewBuffer(z_aiBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *z_aiToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to call Z.AI: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read Z.AI response
	z_aiRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read Z.AI response", http.StatusBadGateway)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Z.AI error: %s", string(z_aiRespBody)), resp.StatusCode)
		return
	}

	// Parse Z.AI response
	var z_aiResp Z_AIResponse
	if err := json.Unmarshal(z_aiRespBody, &z_aiResp); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse Z.AI response: %v", err), http.StatusBadGateway)
		return
	}

	// Check for Z.AI errors
	if z_aiResp.Error != nil {
		http.Error(w, fmt.Sprintf("Z.AI returned error: %v", z_aiResp.Error), http.StatusBadGateway)
		return
	}

	// Convert to Anthropic format
	anthropicResp := AnthropicResponse{
		ID:    fmt.Sprintf("msg_%s", z_aiResp.ID),
		Type:  "message",
		Role:  "assistant",
		Model: z_aiResp.Model,
	}

	// Extract message content
	if len(z_aiResp.Choices) > 0 {
		anthropicResp.Content = []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			{
				Type: "text",
				Text: z_aiResp.Choices[0].Message.Content,
			},
		}

		// Set stop reason
		if z_aiResp.Choices[0].FinishReason == "stop" {
			anthropicResp.StopReason = "end_turn"
		} else {
			anthropicResp.StopReason = "max_tokens"
		}
	}

	anthropicResp.StopSequence = nil
	anthropicResp.Usage.InputTokens = z_aiResp.Usage.PromptTokens
	anthropicResp.Usage.OutputTokens = z_aiResp.Usage.CompletionTokens
	anthropicResp.Usage.CacheCreationInputTokens = 0
	anthropicResp.Usage.CacheReadInputTokens = 0

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(anthropicResp)
}

func main() {
	flag.Parse()

	if *z_aiToken == "" {
		log.Fatal("Error: -z-ai-token is required")
	}

	http.HandleFunc("/v1/messages", handleMessages)

	addr := fmt.Sprintf("127.0.0.1:%s", *port)
	log.Printf("Z.AI Proxy listening on %s", addr)
	log.Printf("  Z.AI Base URL: %s", *z_aiURL)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down proxy...")
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
