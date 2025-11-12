package proxy

import (
	"net/http"
	"strings"
)

// AuthMiddleware provides authentication for proxy endpoints
func (p *Proxy) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoint
		if r.URL.Path == "/health" {
			next(w, r)
			return
		}

		// If auth is not required, proceed
		if !p.config.AuthRequired || p.config.ProxyAuthToken == "" {
			next(w, r)
			return
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format. Use: Bearer <token>", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token != p.config.ProxyAuthToken {
			http.Error(w, "Invalid authorization token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
