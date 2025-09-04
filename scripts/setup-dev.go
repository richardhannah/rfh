package main

import (
	"fmt"
	"log"

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

	fmt.Printf("✅ Database connection established\n")

	// Show setup instructions
	fmt.Printf("🚀 Setup complete! Next steps:\n")
	fmt.Printf("   1. Start API: go run ./cmd/api\n")
	fmt.Printf("   2. Add registry: ./rfh registry add local http://localhost:8080\n")
	fmt.Printf("   3. Authenticate: ./rfh auth login\n")
	fmt.Printf("   4. Initialize package: ./rfh init\n")
	fmt.Printf("   5. Pack and publish: ./rfh pack && ./rfh publish\n")
	fmt.Printf("\n💡 Authentication now uses JWT tokens via 'rfh auth login'\n")
}
