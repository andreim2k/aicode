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
	port           = flag.String("port", "9000", "Port to listen on")
	zaiURL         = flag.String("z-ai-url", "https://api.z.ai/api/paas/v4", "Z.AI base URL")
	zaiToken       = flag.String("z-ai-token", "", "Z.AI auth token (or use ZAI_TOKEN env var)")
	proxyAuthToken = flag.String("proxy-auth-token", "", "Proxy authentication token (or use PROXY_AUTH_TOKEN env var)")
	authRequired   = flag.Bool("auth-required", false, "Require authentication for proxy endpoints")
)

func main() {
	flag.Parse()

	// Get token from environment variable if not provided via flag
	authToken := *zaiToken
	if authToken == "" {
		authToken = os.Getenv("ZAI_TOKEN")
	}
	if authToken == "" {
		log.Fatal("Error: Z.AI auth token is required (use -z-ai-token flag or ZAI_TOKEN env var)")
	}

	// Get proxy auth token from environment variable if not provided via flag
	proxyAuth := *proxyAuthToken
	if proxyAuth == "" {
		proxyAuth = os.Getenv("PROXY_AUTH_TOKEN")
	}

	// Create proxy configuration
	config := proxy.Config{
		ProviderName:   "Z.AI",
		BaseURL:        *zaiURL,
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

	log.Printf("Z.AI Proxy listening on %s", server.Addr)
	log.Printf("  Z.AI Base URL: %s", *zaiURL)
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
