package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RouteMetadata contains metadata for a route
type RouteMetadata struct {
	Path                   string
	Method                 string
	RequiresAuthentication bool
	Handler                http.HandlerFunc
	Description            string
	RateLimit              int // requests per minute, 0 = no limit
}

// RouteRegistry manages route metadata and registration
type RouteRegistry struct {
	routes []RouteMetadata
}

// NewRouteRegistry creates a new route registry
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes: make([]RouteMetadata, 0),
	}
}

// RegisterRoute registers a route with metadata
func (rr *RouteRegistry) RegisterRoute(path, method string, requiresAuth bool, handler http.HandlerFunc, description string) {
	route := RouteMetadata{
		Path:                   path,
		Method:                 method,
		RequiresAuthentication: requiresAuth,
		Handler:                handler,
		Description:            description,
		RateLimit:              0, // Default: no rate limit
	}
	rr.routes = append(rr.routes, route)
}

// RegisterRouteWithRateLimit registers a route with rate limiting
func (rr *RouteRegistry) RegisterRouteWithRateLimit(path, method string, requiresAuth bool, handler http.HandlerFunc, description string, rateLimit int) {
	route := RouteMetadata{
		Path:                   path,
		Method:                 method,
		RequiresAuthentication: requiresAuth,
		Handler:                handler,
		Description:            description,
		RateLimit:              rateLimit,
	}
	rr.routes = append(rr.routes, route)
}

// GetRouteMetadata retrieves metadata for a specific route
func (rr *RouteRegistry) GetRouteMetadata(path, method string) (RouteMetadata, bool) {
	for _, route := range rr.routes {
		if route.Path == path && route.Method == method {
			return route, true
		}
	}
	return RouteMetadata{}, false
}

// GetAllRoutes returns all registered routes
func (rr *RouteRegistry) GetAllRoutes() []RouteMetadata {
	return rr.routes
}

// SetupRoutes configures all routes with their metadata
func (s *Server) SetupRoutes(router *mux.Router) *RouteRegistry {
	registry := NewRouteRegistry()

	// Create API v1 subrouter
	api := router.PathPrefix("/v1").Subrouter()

	// Health endpoint - public, no auth required
	registry.RegisterRoute("/v1/health", "GET", false, s.healthHandler, "API health check")
	api.HandleFunc("/health", s.healthHandler).Methods("GET")

	// Search endpoints - public, with rate limiting
	registry.RegisterRouteWithRateLimit("/v1/packages", "GET", false, s.searchPackagesHandler, "Search packages", 60)
	api.HandleFunc("/packages", s.searchPackagesHandler).Methods("GET")

	// Blob download - public, with rate limiting for abuse prevention
	registry.RegisterRouteWithRateLimit("/v1/blobs/{sha256}", "GET", false, s.downloadBlobHandler, "Download package blob", 30)
	api.HandleFunc("/blobs/{sha256}", s.downloadBlobHandler).Methods("GET")

	// Unscoped package routes (must come before scoped routes)
	registry.RegisterRouteWithRateLimit("/v1/packages/{name}/versions/{version}", "GET", false, s.getUnscopedPackageVersionHandler, "Get package version (unscoped)", 120)
	api.HandleFunc("/packages/{name}/versions/{version}", s.getUnscopedPackageVersionHandler).Methods("GET")
	
	registry.RegisterRouteWithRateLimit("/v1/packages/{name}", "GET", false, s.getUnscopedPackageHandler, "Get package details (unscoped)", 120)
	api.HandleFunc("/packages/{name}", s.getUnscopedPackageHandler).Methods("GET")

	// Scoped package routes (more specific)
	registry.RegisterRouteWithRateLimit("/v1/packages/{scope}/{name}/versions/{version}", "GET", false, s.getPackageVersionHandler, "Get package version", 120)
	api.HandleFunc("/packages/{scope}/{name}/versions/{version}", s.getPackageVersionHandler).Methods("GET")
	
	registry.RegisterRouteWithRateLimit("/v1/packages/{scope}/{name}", "GET", false, s.getPackageHandler, "Get package details", 120)
	api.HandleFunc("/packages/{scope}/{name}", s.getPackageHandler).Methods("GET")

	// Publishing - requires authentication, with rate limiting
	registry.RegisterRouteWithRateLimit("/v1/packages", "POST", true, s.publishPackageHandler, "Publish package", 10)
	api.HandleFunc("/packages", s.publishPackageHandler).Methods("POST")

	return registry
}