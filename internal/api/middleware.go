package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Context types moved to security_middleware.go

// Legacy authMiddleware removed - using security_middleware.go instead

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
