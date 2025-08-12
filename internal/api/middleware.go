package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"rulestack/internal/db"
)

type contextKey string

const tokenContextKey contextKey = "token"

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

// getTokenFromContext retrieves token from request context
func getTokenFromContext(ctx context.Context) *db.Token {
	token, ok := ctx.Value(tokenContextKey).(*db.Token)
	if !ok {
		return nil
	}
	return token
}

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