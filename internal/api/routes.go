package api

import (
	"log"
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