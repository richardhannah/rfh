package api

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// RouteMetadata contains metadata for a route
type RouteMetadata struct {
	Path                   string
	Method                 string
	RequiresAuthentication bool
	RequiredRole           string // "user", "publisher", "admin", or "" for public
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
	requiredRole := ""
	if requiresAuth {
		requiredRole = "user" // Default to user level access
	}
	route := RouteMetadata{
		Path:                   path,
		Method:                 method,
		RequiresAuthentication: requiresAuth,
		RequiredRole:           requiredRole,
		Handler:                handler,
		Description:            description,
		RateLimit:              0, // Default: no rate limit
	}
	rr.routes = append(rr.routes, route)
}

// RegisterRouteWithRole registers a route with specific role requirement
func (rr *RouteRegistry) RegisterRouteWithRole(path, method, requiredRole string, handler http.HandlerFunc, description string) {
	requiresAuth := requiredRole != ""
	route := RouteMetadata{
		Path:                   path,
		Method:                 method,
		RequiresAuthentication: requiresAuth,
		RequiredRole:           requiredRole,
		Handler:                handler,
		Description:            description,
		RateLimit:              0, // Default: no rate limit
	}
	rr.routes = append(rr.routes, route)
}

// RegisterRouteWithRateLimit registers a route with rate limiting
func (rr *RouteRegistry) RegisterRouteWithRateLimit(path, method string, requiresAuth bool, handler http.HandlerFunc, description string, rateLimit int) {
	requiredRole := ""
	if requiresAuth {
		requiredRole = "user" // Default to user level access
	}
	route := RouteMetadata{
		Path:                   path,
		Method:                 method,
		RequiresAuthentication: requiresAuth,
		RequiredRole:           requiredRole,
		Handler:                handler,
		Description:            description,
		RateLimit:              rateLimit,
	}
	rr.routes = append(rr.routes, route)
}

// RegisterRouteWithRoleAndRateLimit registers a route with role requirement and rate limiting
func (rr *RouteRegistry) RegisterRouteWithRoleAndRateLimit(path, method, requiredRole string, handler http.HandlerFunc, description string, rateLimit int) {
	requiresAuth := requiredRole != ""
	route := RouteMetadata{
		Path:                   path,
		Method:                 method,
		RequiresAuthentication: requiresAuth,
		RequiredRole:           requiredRole,
		Handler:                handler,
		Description:            description,
		RateLimit:              rateLimit,
	}
	rr.routes = append(rr.routes, route)
}

// GetRouteMetadata retrieves metadata for a specific route
func (rr *RouteRegistry) GetRouteMetadata(path, method string) (RouteMetadata, bool) {
	for _, route := range rr.routes {
		if matchRoute(route.Path, path) && route.Method == method {
			return route, true
		}
	}
	return RouteMetadata{}, false
}

// matchRoute checks if a template path matches an actual path
func matchRoute(template, actual string) bool {
	// Simple exact match first
	if template == actual {
		return true
	}
	
	// Split paths into segments
	templateParts := strings.Split(strings.Trim(template, "/"), "/")
	actualParts := strings.Split(strings.Trim(actual, "/"), "/")
	
	// Must have same number of segments
	if len(templateParts) != len(actualParts) {
		return false
	}
	
	// Check each segment
	for i, templatePart := range templateParts {
		actualPart := actualParts[i]
		
		// If template part is a parameter (enclosed in {}), it matches any value
		if strings.HasPrefix(templatePart, "{") && strings.HasSuffix(templatePart, "}") {
			continue
		}
		
		// Otherwise, must be exact match
		if templatePart != actualPart {
			return false
		}
	}
	
	return true
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
	registry.RegisterRouteWithRateLimit("/v1/packages", "GET", false, s.searchPackagesHandler, "Search packages", 300)
	api.HandleFunc("/packages", s.searchPackagesHandler).Methods("GET")

	// Blob download - public, with rate limiting for abuse prevention
	registry.RegisterRouteWithRateLimit("/v1/blobs/{sha256}", "GET", false, s.downloadBlobHandler, "Download package blob", 150)
	api.HandleFunc("/blobs/{sha256}", s.downloadBlobHandler).Methods("GET")

	// Unscoped package routes (must come before scoped routes)
	registry.RegisterRouteWithRateLimit("/v1/packages/{name}/versions/{version}", "GET", false, s.getUnscopedPackageVersionHandler, "Get package version (unscoped)", 600)
	api.HandleFunc("/packages/{name}/versions/{version}", s.getUnscopedPackageVersionHandler).Methods("GET")
	
	registry.RegisterRouteWithRateLimit("/v1/packages/{name}", "GET", false, s.getUnscopedPackageHandler, "Get package details (unscoped)", 600)
	api.HandleFunc("/packages/{name}", s.getUnscopedPackageHandler).Methods("GET")

	// Scoped package routes (more specific)
	registry.RegisterRouteWithRateLimit("/v1/packages/{scope}/{name}/versions/{version}", "GET", false, s.getPackageVersionHandler, "Get package version", 600)
	api.HandleFunc("/packages/{scope}/{name}/versions/{version}", s.getPackageVersionHandler).Methods("GET")
	
	registry.RegisterRouteWithRateLimit("/v1/packages/{scope}/{name}", "GET", false, s.getPackageHandler, "Get package details", 600)
	api.HandleFunc("/packages/{scope}/{name}", s.getPackageHandler).Methods("GET")

	// Publishing - requires publisher role, with rate limiting
	registry.RegisterRouteWithRoleAndRateLimit("/v1/packages", "POST", "publisher", s.publishPackageHandler, "Publish package", 50)
	api.HandleFunc("/packages", s.publishPackageHandler).Methods("POST")

	// Authentication endpoints - public for registration and login
	registry.RegisterRouteWithRateLimit("/v1/auth/register", "POST", false, s.registerHandler, "User registration", 25)
	api.HandleFunc("/auth/register", s.registerHandler).Methods("POST")
	
	registry.RegisterRouteWithRateLimit("/v1/auth/login", "POST", false, s.loginHandler, "User login", 50)
	api.HandleFunc("/auth/login", s.loginHandler).Methods("POST")

	// User management endpoints - require authentication
	registry.RegisterRouteWithRoleAndRateLimit("/v1/auth/logout", "POST", "user", s.logoutHandler, "User logout", 30)
	api.HandleFunc("/auth/logout", s.logoutHandler).Methods("POST")
	
	registry.RegisterRouteWithRoleAndRateLimit("/v1/auth/profile", "GET", "user", s.profileHandler, "Get user profile", 60)
	api.HandleFunc("/auth/profile", s.profileHandler).Methods("GET")
	
	registry.RegisterRouteWithRoleAndRateLimit("/v1/auth/change-password", "POST", "user", s.changePasswordHandler, "Change password", 5)
	api.HandleFunc("/auth/change-password", s.changePasswordHandler).Methods("POST")
	
	registry.RegisterRouteWithRoleAndRateLimit("/v1/auth/delete-account", "DELETE", "user", s.deleteAccountHandler, "Delete account", 2)
	api.HandleFunc("/auth/delete-account", s.deleteAccountHandler).Methods("DELETE")

	// Admin endpoints - require admin role
	registry.RegisterRouteWithRoleAndRateLimit("/v1/admin/users", "GET", "admin", s.listUsersHandler, "List all users", 30)
	api.HandleFunc("/admin/users", s.listUsersHandler).Methods("GET")
	
	registry.RegisterRouteWithRoleAndRateLimit("/v1/admin/users/{id}", "DELETE", "admin", s.adminDeleteUserHandler, "Admin delete user", 5)
	api.HandleFunc("/admin/users/{id}", s.adminDeleteUserHandler).Methods("DELETE")

	return registry
}