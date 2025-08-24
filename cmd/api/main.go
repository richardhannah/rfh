package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"

	"rulestack/internal/api"
	"rulestack/internal/config"
	"rulestack/internal/db"
)

func main() {
	// Load configuration from system environment variables
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
