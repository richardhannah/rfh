package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"rulestack/internal/db"
)

// Context types moved to security_middleware.go

// authMiddleware validates Bearer tokens
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Authorization header must be 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token == "" {
			http.Error(w, "Token cannot be empty", http.StatusUnauthorized)
			return
		}

		// Hash token and validate
		tokenHash := db.HashToken(token, s.Config.TokenSalt)
		dbToken, err := s.DB.ValidateToken(tokenHash)
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, "Token validation failed", http.StatusInternalServerError)
			return
		}

		// Add token to context and continue
		ctx := context.WithValue(r.Context(), tokenContextKey, dbToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getTokenFromContext retrieves token from request context (legacy support)
// Note: This function is now defined in security_middleware.go

// writeJSON writes JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes JSON error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// panicRecoveryMiddleware recovers from panics and returns a 500 error
func panicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic (in a real app, use proper logging)
				// Print to stderr so it shows in Docker logs
				fmt.Fprintf(os.Stderr, "PANIC in %s %s: %v\n", r.Method, r.URL.Path, err)
				
				// Return 500 error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}