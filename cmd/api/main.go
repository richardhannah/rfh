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

	// Ensure root user exists
	if err := ensureRootUser(database); err != nil {
		log.Printf("Warning: Failed to ensure root user exists: %v", err)
		// Don't fail startup, just log the error
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

// ensureRootUser checks if a root user exists and creates one if not
func ensureRootUser(database *db.DB) error {
	// Check if any root user exists
	var exists bool
	err := database.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM rulestack.users WHERE role = 'root'
		)
	`).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		log.Println("Root user already exists")
		return nil
	}

	// Create the root user with hardcoded defaults
	log.Println("Creating default root user...")

	user := &db.CreateUserRequest{
		Username: "root",
		Email:    "root@rulestack.init",
		Password: "root1234",
		Role:     db.RoleRoot,
	}

	// Use the existing CreateUser method to create the root user
	_, err = database.CreateUser(*user)
	if err != nil {
		// Check if it's a duplicate error (race condition)
		if err.Error() == "username or email already exists" {
			log.Println("Root user was created by another process")
			return nil
		}
		return err
	}

	log.Println("Root user created successfully")
	return nil
}
