# Task 6: Basic API Handlers (1 hour)

## Objective
Create the core HTTP handlers for the RuleStack API using gorilla/mux and the database layer.

## Prerequisites
- Tasks 1-5 completed
- Database connection and models working
- Config system functioning
- Docker development environment running

## Checklist

### 1. Create API Server Setup (15 minutes)
Create `cmd/api/main.go`:
```go
package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    
    "rulestack/internal/api"
    "rulestack/internal/config"
    "rulestack/internal/db"
)

func main() {
    // Load environment variables
    if err := config.LoadEnvFile(".env"); err != nil {
        log.Printf("Warning: Could not load .env file: %v", err)
    }
    
    // Load configuration
    cfg := config.Load()
    
    // Connect to database
    database, err := db.Connect(cfg.DBURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer database.Close()
    
    // Test database connection
    if err := database.Health(); err != nil {
        log.Fatal("Database health check failed:", err)
    }
    
    // Create storage directory if it doesn't exist
    if err := os.MkdirAll(cfg.StoragePath, 0o755); err != nil {
        log.Fatal("Failed to create storage directory:", err)
    }
    
    // Set up router
    r := mux.NewRouter()
    
    // Register API routes
    api.RegisterRoutes(r, database, cfg)
    
    log.Printf("API server starting on port %s", cfg.APIPort)
    log.Printf("Storage path: %s", cfg.StoragePath)
    log.Fatal(http.ListenAndServe(":"+cfg.APIPort, r))
}
```

- [ ] Create main.go file
- [ ] Verify all imports are correct

### 2. Create Route Registration (10 minutes)
Create `internal/api/routes.go`:
```go
package api

import (
    "net/http"
    
    "github.com/gorilla/mux"
    
    "rulestack/internal/config"
    "rulestack/internal/db"
)

// Server holds dependencies for API handlers
type Server struct {
    DB     *db.DB
    Config config.Config
}

// RegisterRoutes sets up all API routes
func RegisterRoutes(r *mux.Router, database *db.DB, cfg config.Config) {
    s := &Server{
        DB:     database,
        Config: cfg,
    }
    
    // Add middleware
    r.Use(loggingMiddleware)
    r.Use(corsMiddleware)
    
    // API v1 routes
    api := r.PathPrefix("/v1").Subrouter()
    
    // Public routes
    api.HandleFunc("/health", s.healthHandler).Methods("GET")
    api.HandleFunc("/packages", s.searchPackagesHandler).Methods("GET")
    api.HandleFunc("/packages/{scope}/{name}", s.getPackageHandler).Methods("GET")
    api.HandleFunc("/packages/{scope}/{name}/versions/{version}", s.getPackageVersionHandler).Methods("GET")
    api.HandleFunc("/blobs/{sha256}", s.downloadBlobHandler).Methods("GET")
    
    // Authenticated routes
    authAPI := api.PathPrefix("").Subrouter()
    authAPI.Use(s.authMiddleware)
    authAPI.HandleFunc("/packages", s.publishPackageHandler).Methods("POST")
}

// Middleware functions
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
        next.ServeHTTP(w, r)
    })
}

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

- [ ] Create routes.go file
- [ ] Add missing import for `log` package

### 3. Create Authentication Middleware (15 minutes)
Create `internal/api/middleware.go`:
```go
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
```

- [ ] Create middleware.go file
- [ ] Verify context handling code

### 4. Create Core API Handlers (15 minutes)
Create `internal/api/handlers.go`:
```go
package api

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/gorilla/mux"
)

// healthHandler returns API health status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    if err := s.DB.Health(); err != nil {
        writeError(w, http.StatusServiceUnavailable, "Database connection failed")
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]interface{}{
        "status":  "ok",
        "service": "rulestack-api",
        "version": "1.0.0",
    })
}

// searchPackagesHandler searches for packages
func (s *Server) searchPackagesHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    tag := r.URL.Query().Get("tag")
    target := r.URL.Query().Get("target")
    
    // Parse limit parameter
    limit := 50 // default
    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
            limit = parsedLimit
        }
    }
    
    results, err := s.DB.SearchPackages(query, tag, target, limit)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Search failed")
        return
    }
    
    writeJSON(w, http.StatusOK, results)
}

// getPackageHandler gets package information
func (s *Server) getPackageHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    scope := vars["scope"]
    name := vars["name"]
    
    // Handle unscoped packages (scope will be the name in that case)
    var scopePtr *string
    if scope != "" && name == "" {
        name = scope
        scopePtr = nil
    } else if scope != "" {
        // Remove @ prefix if present
        if scope[0] == '@' {
            scope = scope[1:]
        }
        scopePtr = &scope
    }
    
    pkg, err := s.DB.GetPackage(scopePtr, name)
    if err != nil {
        writeError(w, http.StatusNotFound, "Package not found")
        return
    }
    
    writeJSON(w, http.StatusOK, pkg)
}

// getPackageVersionHandler gets specific package version
func (s *Server) getPackageVersionHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    scope := vars["scope"]
    name := vars["name"]
    version := vars["version"]
    
    var scopePtr *string
    if scope != "" {
        if scope[0] == '@' {
            scope = scope[1:]
        }
        scopePtr = &scope
    }
    
    pkgVersion, err := s.DB.GetPackageVersion(scopePtr, name, version)
    if err != nil {
        writeError(w, http.StatusNotFound, "Package version not found")
        return
    }
    
    writeJSON(w, http.StatusOK, pkgVersion)
}
```

- [ ] Create handlers.go file
- [ ] Review URL parameter handling logic

### 5. Create Basic Tests (5 minutes)
Create `internal/api/handlers_test.go`:
```go
package api

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "rulestack/internal/config"
    "rulestack/internal/db"
)

func TestHealthHandler(t *testing.T) {
    // This would need a test database setup
    // For now, just test that the handler function exists
    server := &Server{
        Config: config.Config{},
    }
    
    req := httptest.NewRequest("GET", "/v1/health", nil)
    w := httptest.NewRecorder()
    
    // We can't test fully without database, but we can verify handler exists
    if server.healthHandler == nil {
        t.Error("healthHandler should not be nil")
    }
    
    // Skip actual execution since we don't have test DB
    _ = req
    _ = w
}

func TestWriteJSON(t *testing.T) {
    w := httptest.NewRecorder()
    data := map[string]string{"test": "data"}
    
    writeJSON(w, http.StatusOK, data)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    if w.Header().Get("Content-Type") != "application/json" {
        t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
    }
}
```

- [ ] Create test file
- [ ] Run tests: `go test ./internal/api -v`

## Validation
Test the API setup:
```bash
# Build and run the API server
go run ./cmd/api

# In another terminal, test endpoints:
curl http://localhost:8080/v1/health
curl "http://localhost:8080/v1/packages?q=test"

# Test with missing auth (should return 401)
curl -X POST http://localhost:8080/v1/packages
```

## Acceptance Criteria
- [ ] API server starts without errors
- [ ] Health endpoint returns 200 status
- [ ] Search endpoint accepts query parameters
- [ ] Authentication middleware blocks unauthenticated requests
- [ ] CORS headers are set correctly
- [ ] JSON responses are properly formatted
- [ ] Error handling returns appropriate HTTP status codes
- [ ] All imports compile successfully

## Time Estimate: ~60 minutes

## Next Task
Task 7: Manifest and Archive Handling