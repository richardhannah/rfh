package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"rulestack/internal/db"
)

// Enhanced auth middleware with route registry support
func (s *Server) enhancedAuthMiddleware(registry *RouteRegistry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip OPTIONS requests
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Check if route requires authentication
			if registry != nil {
				if metadata, found := registry.GetRouteMetadata(r.URL.Path, r.Method); found {
					if !metadata.RequiresAuthentication {
						// Route doesn't require authentication, proceed
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			// Validate Bearer token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			// Check Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, http.StatusUnauthorized, "Authorization header must be 'Bearer <token>'")
				return
			}

			token := parts[1]
			if token == "" {
				writeError(w, http.StatusUnauthorized, "Token cannot be empty")
				return
			}

			// Hash token and validate
			tokenHash := db.HashToken(token, s.Config.TokenSalt)
			dbToken, err := s.DB.ValidateToken(tokenHash)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Add token to context and continue
			ctx := context.WithValue(r.Context(), tokenContextKey, dbToken)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORS middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// JSON sanitization middleware
func (s *Server) jsonSanitizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle JSON POST/PUT requests
		if (r.Method == http.MethodPost || r.Method == http.MethodPut) &&
			strings.Contains(r.Header.Get("Content-Type"), "application/json") {

			// Read the request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				writeError(w, http.StatusBadRequest, "Failed to read request body")
				return
			}
			r.Body.Close()

			// Parse JSON
			var data interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				writeError(w, http.StatusBadRequest, "Invalid JSON")
				return
			}

			// Sanitize using strict policy
			policy := bluemonday.StrictPolicy()
			sanitized := sanitizeData(data, policy)

			// Re-encode sanitized data
			newBody, err := json.Marshal(sanitized)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to encode sanitized JSON")
				return
			}

			// Replace request body
			r.Body = io.NopCloser(bytes.NewReader(newBody))
			r.ContentLength = int64(len(newBody))
			r.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
		}

		next.ServeHTTP(w, r)
	})
}

// Rate limiting middleware
type rateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	cleanup  chan string
}

type visitor struct {
	tokens   int
	lastSeen time.Time
}

func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		cleanup:  make(chan string, 100),
	}
	
	// Cleanup goroutine
	go rl.cleanupVisitors()
	
	return rl
}

func (rl *rateLimiter) cleanupVisitors() {
	for {
		select {
		case ip := <-rl.cleanup:
			rl.mu.Lock()
			delete(rl.visitors, ip)
			rl.mu.Unlock()
		case <-time.After(time.Minute):
			// Periodic cleanup of old visitors
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 5*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

func (rl *rateLimiter) allow(ip string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:   limit - 1,
			lastSeen: time.Now(),
		}
		return true
	}

	// Token bucket refill
	now := time.Now()
	elapsed := now.Sub(v.lastSeen)
	tokensToAdd := int(elapsed.Minutes())
	
	v.tokens += tokensToAdd
	if v.tokens > limit {
		v.tokens = limit
	}
	v.lastSeen = now

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

func (s *Server) rateLimitMiddleware(registry *RouteRegistry) func(http.Handler) http.Handler {
	limiter := newRateLimiter()
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := getClientIP(r)
			
			// Check if route has rate limit
			if registry != nil {
				if metadata, found := registry.GetRouteMetadata(r.URL.Path, r.Method); found {
					if metadata.RateLimit > 0 {
						if !limiter.allow(ip, metadata.RateLimit) {
							writeError(w, http.StatusTooManyRequests, "Rate limit exceeded")
							return
						}
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Security headers middleware
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// Request logging middleware
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		log.Printf("[%s] %s %s - %d (%v) - %s",
			getClientIP(r),
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
			r.UserAgent(),
		)
	})
}

// Request size limiting middleware
func (s *Server) requestSizeLimitMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// Utility functions
func sanitizeData(v interface{}, policy *bluemonday.Policy) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, sub := range val {
			val[k] = sanitizeData(sub, policy)
		}
		return val
	case []interface{}:
		for i, sub := range val {
			val[i] = sanitizeData(sub, policy)
		}
		return val
	case string:
		return policy.Sanitize(val)
	default:
		return v
	}
}

func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (behind proxy)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check for X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip, _, _ = strings.Cut(ip, ":")
	}
	return ip
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}