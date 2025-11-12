package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Config holds the proxy configuration
type Config struct {
	ProviderName  string
	BaseURL       string
	Port          string
	AuthToken     string // Token for the provider API
	RequestID     string
	AuthRequired  bool   // Whether authentication is required for proxy endpoints
	ProxyAuthToken string // Token for proxy authentication (if AuthRequired is true)
}

// Proxy handles HTTP requests and converts between Anthropic and provider formats
type Proxy struct {
	config Config
	client *http.Client
}

// NewProxy creates a new proxy instance
func NewProxy(config Config) *Proxy {
	// Create HTTP client with proper timeouts
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}

	return &Proxy{
		config: config,
		client: client,
	}
}

// HandleMessages handles the /v1/messages endpoint
func (p *Proxy) HandleMessages(w http.ResponseWriter, r *http.Request) {
	// Generate or use request ID for tracing
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	w.Header().Set("X-Request-ID", requestID)

	// Validate request method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate request size
	if err := ValidateRequest(r); err != nil {
		log.Printf("[%s] Validation error: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}

	// Read request body with size limit
	body, err := io.ReadAll(io.LimitReader(r.Body, MaxRequestBodySize))
	if err != nil {
		log.Printf("[%s] Failed to read body: %v", requestID, err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse Anthropic request
	var anthropicReq AnthropicRequest
	if err := json.Unmarshal(body, &anthropicReq); err != nil {
		log.Printf("[%s] Invalid JSON: %v", requestID, err)
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Validate Anthropic request
	if err := ValidateAnthropicRequest(&anthropicReq); err != nil {
		log.Printf("[%s] Validation error: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert Anthropic messages to provider format
	providerReq, err := ConvertAnthropicToProvider(&anthropicReq)
	if err != nil {
		log.Printf("[%s] Conversion error: %v", requestID, err)
		http.Error(w, fmt.Sprintf("Failed to convert request: %v", err), http.StatusInternalServerError)
		return
	}

	// Marshal to JSON
	providerBody, err := json.Marshal(providerReq)
	if err != nil {
		log.Printf("[%s] Failed to marshal request: %v", requestID, err)
		http.Error(w, fmt.Sprintf("Failed to marshal request: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare provider API URL
	baseURL := p.config.BaseURL
	if baseURL != "" && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	providerURL := baseURL + "/chat/completions"

	// Create request to provider API
	req, err := http.NewRequest("POST", providerURL, bytes.NewBuffer(providerBody))
	if err != nil {
		log.Printf("[%s] Failed to create request: %v", requestID, err)
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.AuthToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", requestID)

	// Call provider API
	resp, err := p.client.Do(req)
	if err != nil {
		log.Printf("[%s] Failed to call %s: %v", requestID, p.config.ProviderName, err)
		http.Error(w, fmt.Sprintf("Failed to call %s: %v", p.config.ProviderName, err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read provider response
	providerRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%s] Failed to read %s response: %v", requestID, p.config.ProviderName, err)
		http.Error(w, fmt.Sprintf("Failed to read %s response", p.config.ProviderName), http.StatusBadGateway)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[%s] %s error (status %d): %s", requestID, p.config.ProviderName, resp.StatusCode, string(providerRespBody))
		http.Error(w, fmt.Sprintf("%s error: %s", p.config.ProviderName, string(providerRespBody)), resp.StatusCode)
		return
	}

	// Parse provider response
	var providerResp ProviderResponse
	if err := json.Unmarshal(providerRespBody, &providerResp); err != nil {
		log.Printf("[%s] Failed to parse %s response: %v", requestID, p.config.ProviderName, err)
		http.Error(w, fmt.Sprintf("Failed to parse %s response: %v", p.config.ProviderName, err), http.StatusBadGateway)
		return
	}

	// Check for provider errors
	if providerResp.Error != nil {
		log.Printf("[%s] %s returned error: %v", requestID, p.config.ProviderName, providerResp.Error)
		http.Error(w, fmt.Sprintf("%s returned error: %v", p.config.ProviderName, providerResp.Error), http.StatusBadGateway)
		return
	}

	// Convert to Anthropic format
	anthropicResp := ConvertProviderToAnthropic(&providerResp)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(anthropicResp); err != nil {
		log.Printf("[%s] Failed to encode response: %v", requestID, err)
		// Response may have already been sent, so we can't send another error
		return
	}

	log.Printf("[%s] Successfully processed request for model: %s", requestID, anthropicReq.Model)
}

// HandleHealth handles the /health endpoint
func (p *Proxy) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"provider": p.config.ProviderName,
	})
}
