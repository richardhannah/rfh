package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"rulestack/internal/config"
	"rulestack/internal/db"
)

func main() {
	// Load environment
	config.LoadEnvFile(".env")
	cfg := config.Load()

	// Connect to database
	database, err := db.Connect(cfg.DBURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Create a test token
	tokenValue := "dev-token-12345"
	tokenHash := db.HashToken(tokenValue, cfg.TokenSalt)

	// Check if token already exists
	var exists bool
	err = database.Get(&exists, "SELECT EXISTS(SELECT 1 FROM tokens WHERE token_hash = $1)", tokenHash)
	if err != nil {
		log.Fatal("Failed to check token:", err)
	}

	if !exists {
		// Insert test token
		name := "Development Token"
		_, err = database.CreateToken(tokenHash, &name)
		if err != nil {
			log.Fatal("Failed to create token:", err)
		}

		fmt.Printf("✅ Created development token\n")
		fmt.Printf("🔑 Token: %s\n", tokenValue)
		fmt.Printf("⚠️  Save this token - you'll need it for publishing\n\n")
	} else {
		fmt.Printf("✅ Development token already exists\n")
		fmt.Printf("🔑 Token: %s\n", tokenValue)
	}

	// Show setup instructions
	fmt.Printf("🚀 Setup complete! Next steps:\n")
	fmt.Printf("   1. Start API: go run ./cmd/api\n")
	fmt.Printf("   2. Add registry: ./rfh registry add local http://localhost:8080 %s\n", tokenValue)
	fmt.Printf("   3. Initialize package: ./rfh init\n")
	fmt.Printf("   4. Pack and publish: ./rfh pack && ./rfh publish\n")
}