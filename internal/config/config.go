package config

import (
	"log"
	"os"
)

type Config struct {
	DBURL       string
	StoragePath string
	APIPort     string
	TokenSalt   string
	JWTSecret   string
}

func Load() Config {
	cfg := Config{
		DBURL:       os.Getenv("DATABASE_URL"),
		StoragePath: getEnv("STORAGE_PATH", "./storage"),
		APIPort:     getEnv("PORT", "8080"),
		TokenSalt:   os.Getenv("TOKEN_SALT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	// Validate required fields
	if cfg.DBURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	if cfg.TokenSalt == "" {
		log.Fatal("TOKEN_SALT environment variable is required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}