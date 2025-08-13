package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns environment value when set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "env_value",
			expected:     "env_value",
		},
		{
			name:         "returns default when env not set",
			key:          "UNSET_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "returns empty string as env value",
			key:          "EMPTY_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			os.Unsetenv(tt.key)
			
			// Set environment variable if specified
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Save original env vars
	originalDBURL := os.Getenv("DATABASE_URL")
	originalTokenSalt := os.Getenv("TOKEN_SALT")
	originalStoragePath := os.Getenv("STORAGE_PATH")
	originalPort := os.Getenv("PORT")

	// Clean up after test
	defer func() {
		setOrUnset("DATABASE_URL", originalDBURL)
		setOrUnset("TOKEN_SALT", originalTokenSalt)
		setOrUnset("STORAGE_PATH", originalStoragePath)
		setOrUnset("PORT", originalPort)
	}()

	t.Run("loads config with all env vars set", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://test")
		os.Setenv("TOKEN_SALT", "test_salt")
		os.Setenv("STORAGE_PATH", "/tmp/storage")
		os.Setenv("PORT", "9000")

		cfg := Load()

		if cfg.DBURL != "postgres://test" {
			t.Errorf("DBURL = %q, want %q", cfg.DBURL, "postgres://test")
		}
		if cfg.TokenSalt != "test_salt" {
			t.Errorf("TokenSalt = %q, want %q", cfg.TokenSalt, "test_salt")
		}
		if cfg.StoragePath != "/tmp/storage" {
			t.Errorf("StoragePath = %q, want %q", cfg.StoragePath, "/tmp/storage")
		}
		if cfg.APIPort != "9000" {
			t.Errorf("APIPort = %q, want %q", cfg.APIPort, "9000")
		}
	})

	t.Run("uses defaults for optional vars", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://test")
		os.Setenv("TOKEN_SALT", "test_salt")
		os.Unsetenv("STORAGE_PATH")
		os.Unsetenv("PORT")

		cfg := Load()

		if cfg.StoragePath != "./storage" {
			t.Errorf("StoragePath = %q, want %q", cfg.StoragePath, "./storage")
		}
		if cfg.APIPort != "8080" {
			t.Errorf("APIPort = %q, want %q", cfg.APIPort, "8080")
		}
	})
}

// Helper function to set or unset environment variable
func setOrUnset(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}