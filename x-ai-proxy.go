package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aicode/proxy/proxy"
)

var (
	port           = flag.String("port", "9001", "Port to listen on")
	xaiURL         = flag.String("xai-url", "https://api.x.ai/v1", "X.AI base URL")
	xaiToken       = flag.String("xai-token", "", "X.AI auth token (or use XAI_TOKEN env var)")
	proxyAuthToken = flag.String("proxy-auth-token", "", "Proxy authentication token (or use PROXY_AUTH_TOKEN env var)")
	authRequired   = flag.Bool("auth-required", false, "Require authentication for proxy endpoints")
)

func main() {
	flag.Parse()

	// Get token from environment variable if not provided via flag
	authToken := *xaiToken
	if authToken == "" {
		authToken = os.Getenv("XAI_TOKEN")
	}
	if authToken == "" {
		log.Fatal("Error: X.AI auth token is required (use -xai-token flag or XAI_TOKEN env var)")
	}

	// Get proxy auth token from environment variable if not provided via flag
	proxyAuth := *proxyAuthToken
	if proxyAuth == "" {
		proxyAuth = os.Getenv("PROXY_AUTH_TOKEN")
	}

	// Create proxy configuration
	config := proxy.Config{
		ProviderName:   "X.AI",
		BaseURL:        *xaiURL,
		Port:           *port,
		AuthToken:      authToken,
		AuthRequired:   *authRequired,
		ProxyAuthToken: proxyAuth,
	}

	// Create proxy instance
	p := proxy.NewProxy(config)

	// Create HTTP server with graceful shutdown support
	mux := http.NewServeMux()

	// Apply auth middleware if required
	if config.AuthRequired {
		mux.HandleFunc("/v1/messages", p.AuthMiddleware(p.HandleMessages))
		mux.HandleFunc("/health", p.HandleHealth)
	} else {
		mux.HandleFunc("/v1/messages", p.HandleMessages)
		mux.HandleFunc("/health", p.HandleHealth)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%s", *port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down proxy...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("X.AI Proxy listening on %s", server.Addr)
	log.Printf("  X.AI Base URL: %s", *xaiURL)
	if config.AuthRequired {
		log.Printf("  Proxy authentication: REQUIRED")
	} else {
		log.Printf("  Proxy authentication: DISABLED")
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Proxy shutdown complete")
}
