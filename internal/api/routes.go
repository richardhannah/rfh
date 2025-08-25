package api

import (
	"github.com/gorilla/mux"

	"rulestack/internal/config"
	"rulestack/internal/db"
)

// Server holds dependencies for API handlers
type Server struct {
	DB       *db.DB
	Config   config.Config
	Registry *RouteRegistry
}

// RegisterRoutes sets up all API routes with enhanced security
func RegisterRoutes(r *mux.Router, database *db.DB, cfg config.Config) {
	s := &Server{
		DB:     database,
		Config: cfg,
	}

	// Create route registry
	registry := s.SetupRoutes(r)
	s.Registry = registry

	// Apply middleware in order (outermost to innermost)
	r.Use(panicRecoveryMiddleware)           // Panic recovery (outermost)
	r.Use(s.securityHeadersMiddleware)       // Security headers
	r.Use(s.corsMiddleware)                  // CORS
	r.Use(s.loggingMiddleware)               // Request logging
	r.Use(s.requestSizeLimitMiddleware(50*1024*1024)) // 50MB max request size
	r.Use(s.rateLimitMiddleware(registry))   // Rate limiting
	r.Use(s.jsonSanitizeMiddleware)          // JSON sanitization
	r.Use(s.enhancedAuthMiddleware(registry)) // Authentication

	// API v1 routes are now set up in SetupRoutes method
}