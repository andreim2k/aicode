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

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
}

type OpenAIChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type OpenAIResponse struct {
	ID      string        `json:"id"`
	Model   string        `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage   `json:"usage"`
	Error   interface{}   `json:"error,omitempty"`
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
	port    = flag.String("port", "9001", "Port to listen on")
	xaiURL  = flag.String("xai-url", "https://api.x.ai/v1", "X.AI base URL")
	xaiToken = flag.String("xai-token", "", "X.AI auth token")
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

	// Convert Anthropic messages to OpenAI format
	openaiMessages := []OpenAIMessage{}

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
			openaiMessages = append(openaiMessages, OpenAIMessage{
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

		openaiMessages = append(openaiMessages, OpenAIMessage{
			Role:    msg.Role,
			Content: contentStr,
		})
	}

	// Create OpenAI request
	openaiReq := OpenAIRequest{
		Model:       anthropicReq.Model,
		Messages:    openaiMessages,
		MaxTokens:   anthropicReq.MaxTokens,
		Temperature: anthropicReq.Temperature,
		TopP:        anthropicReq.TopP,
	}

	// Marshal to JSON
	openaiBody, err := json.Marshal(openaiReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal request: %v", err), http.StatusInternalServerError)
		return
	}

	// Call X.AI API
	baseURL := *xaiURL
	if baseURL != "" && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	xaiFullURL := baseURL + "/chat/completions"
	req, err := http.NewRequest("POST", xaiFullURL, bytes.NewBuffer(openaiBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *xaiToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to call X.AI: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read X.AI response
	xaiRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read X.AI response", http.StatusBadGateway)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("X.AI error: %s", string(xaiRespBody)), resp.StatusCode)
		return
	}

	// Parse X.AI response
	var xaiResp OpenAIResponse
	if err := json.Unmarshal(xaiRespBody, &xaiResp); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse X.AI response: %v", err), http.StatusBadGateway)
		return
	}

	// Check for X.AI errors
	if xaiResp.Error != nil {
		http.Error(w, fmt.Sprintf("X.AI returned error: %v", xaiResp.Error), http.StatusBadGateway)
		return
	}

	// Convert to Anthropic format
	anthropicResp := AnthropicResponse{
		ID:    fmt.Sprintf("msg_%s", xaiResp.ID),
		Type:  "message",
		Role:  "assistant",
		Model: xaiResp.Model,
	}

	// Extract message content
	if len(xaiResp.Choices) > 0 {
		anthropicResp.Content = []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			{
				Type: "text",
				Text: xaiResp.Choices[0].Message.Content,
			},
		}

		// Set stop reason
		if xaiResp.Choices[0].FinishReason == "stop" {
			anthropicResp.StopReason = "end_turn"
		} else {
			anthropicResp.StopReason = "max_tokens"
		}
	}

	anthropicResp.StopSequence = nil
	anthropicResp.Usage.InputTokens = xaiResp.Usage.PromptTokens
	anthropicResp.Usage.OutputTokens = xaiResp.Usage.CompletionTokens
	anthropicResp.Usage.CacheCreationInputTokens = 0
	anthropicResp.Usage.CacheReadInputTokens = 0

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(anthropicResp)
}

func main() {
	flag.Parse()

	if *xaiToken == "" {
		log.Fatal("Error: -xai-token is required")
	}

	http.HandleFunc("/v1/messages", handleMessages)

	addr := fmt.Sprintf("127.0.0.1:%s", *port)
	log.Printf("X.AI Proxy listening on %s", addr)
	log.Printf("  X.AI Base URL: %s", *xaiURL)

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

