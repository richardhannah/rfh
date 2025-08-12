# Task 4: Configuration Module (1 hour)

## Objective
Create the configuration loading system that reads environment variables and manages app settings.

## Prerequisites
- Tasks 1-3 completed (including Docker setup)
- `.env` file configured with database and other settings
- Docker development environment running

## Checklist

### 1. Create Config Package (15 minutes)
Create `internal/config/config.go`:
```go
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
}

func Load() Config {
    cfg := Config{
        DBURL:       os.Getenv("DATABASE_URL"),
        StoragePath: getEnv("STORAGE_PATH", "./storage"),
        APIPort:     getEnv("PORT", "8080"),
        TokenSalt:   os.Getenv("TOKEN_SALT"),
    }
    
    // Validate required fields
    if cfg.DBURL == "" {
        log.Fatal("DATABASE_URL environment variable is required")
    }
    if cfg.TokenSalt == "" {
        log.Fatal("TOKEN_SALT environment variable is required")
    }
    
    return cfg
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

- [ ] Create the file with above content
- [ ] Review code for syntax errors
- [ ] Add any additional config fields if needed

### 2. Create Environment Loader (10 minutes)
Create `internal/config/env.go` to load `.env` files:
```go
package config

import (
    "bufio"
    "os"
    "strings"
)

// LoadEnvFile loads environment variables from a .env file
func LoadEnvFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err // File might not exist, which is okay
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        
        // Skip empty lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        
        // Split on first = sign
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }
        
        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        
        // Only set if not already set in environment
        if os.Getenv(key) == "" {
            os.Setenv(key, value)
        }
    }
    
    return scanner.Err()
}
```

- [ ] Create the env.go file
- [ ] Test that it can read your `.env` file

### 3. Add CLI Config Support (20 minutes)
Create `internal/config/cli.go` for CLI-specific configuration:
```go
package config

import (
    "os"
    "path/filepath"
    
    "github.com/pelletier/go-toml/v2"
)

type Registry struct {
    URL   string `toml:"url"`
    Token string `toml:"token,omitempty"`
}

type CLIConfig struct {
    Current    string              `toml:"current"`
    Registries map[string]Registry `toml:"registries"`
}

// ConfigDir returns the CLI config directory path
func ConfigDir() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".rfh"), nil
}

// ConfigPath returns the full path to config.toml
func ConfigPath() (string, error) {
    dir, err := ConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dir, "config.toml"), nil
}

// LoadCLI loads CLI configuration from ~/.rfh/config.toml
func LoadCLI() (CLIConfig, error) {
    configPath, err := ConfigPath()
    if err != nil {
        return CLIConfig{}, err
    }
    
    data, err := os.ReadFile(configPath)
    if os.IsNotExist(err) {
        // Return empty config if file doesn't exist
        return CLIConfig{
            Registries: make(map[string]Registry),
        }, nil
    }
    if err != nil {
        return CLIConfig{}, err
    }
    
    var config CLIConfig
    if err := toml.Unmarshal(data, &config); err != nil {
        return CLIConfig{}, err
    }
    
    if config.Registries == nil {
        config.Registries = make(map[string]Registry)
    }
    
    return config, nil
}

// SaveCLI saves CLI configuration to ~/.rfh/config.toml
func SaveCLI(config CLIConfig) error {
    configPath, err := ConfigPath()
    if err != nil {
        return err
    }
    
    // Ensure config directory exists
    if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
        return err
    }
    
    data, err := toml.Marshal(config)
    if err != nil {
        return err
    }
    
    return os.WriteFile(configPath, data, 0o600)
}
```

- [ ] Create cli.go file
- [ ] Verify TOML import works

### 4. Create Test File (10 minutes)
Create `internal/config/config_test.go`:
```go
package config

import (
    "os"
    "testing"
)

func TestLoad(t *testing.T) {
    // Set required environment variables
    os.Setenv("DATABASE_URL", "postgres://test")
    os.Setenv("TOKEN_SALT", "test-salt")
    
    defer func() {
        os.Unsetenv("DATABASE_URL") 
        os.Unsetenv("TOKEN_SALT")
    }()
    
    cfg := Load()
    
    if cfg.DBURL != "postgres://test" {
        t.Errorf("Expected DBURL to be 'postgres://test', got %s", cfg.DBURL)
    }
    
    if cfg.TokenSalt != "test-salt" {
        t.Errorf("Expected TokenSalt to be 'test-salt', got %s", cfg.TokenSalt)
    }
    
    if cfg.APIPort != "8080" {
        t.Errorf("Expected default APIPort to be '8080', got %s", cfg.APIPort)
    }
}

func TestGetEnv(t *testing.T) {
    // Test with existing env var
    os.Setenv("TEST_VAR", "test-value")
    defer os.Unsetenv("TEST_VAR")
    
    result := getEnv("TEST_VAR", "default")
    if result != "test-value" {
        t.Errorf("Expected 'test-value', got %s", result)
    }
    
    // Test with default value
    result = getEnv("NON_EXISTENT", "default")
    if result != "default" {
        t.Errorf("Expected 'default', got %s", result)
    }
}
```

- [ ] Create test file
- [ ] Run tests: `go test ./internal/config`
- [ ] Ensure all tests pass

### 5. Integration Test (5 minutes)
Create a simple test program to verify configuration loading:
```go
// test/config_test/main.go
package main

import (
    "fmt"
    "rulestack/internal/config"
)

func main() {
    // Try to load .env file
    if err := config.LoadEnvFile(".env"); err != nil {
        fmt.Printf("Warning: Could not load .env file: %v\n", err)
    }
    
    // Load main config
    cfg := config.Load()
    fmt.Printf("Config loaded successfully:\n")
    fmt.Printf("  Database URL: %s\n", cfg.DBURL)
    fmt.Printf("  Storage Path: %s\n", cfg.StoragePath)
    fmt.Printf("  API Port: %s\n", cfg.APIPort)
    fmt.Printf("  Token Salt: [REDACTED]\n")
    
    // Test CLI config
    cliCfg, err := config.LoadCLI()
    if err != nil {
        fmt.Printf("CLI config error: %v\n", err)
    } else {
        fmt.Printf("CLI config loaded, registries: %d\n", len(cliCfg.Registries))
    }
}
```

- [ ] Create test program (optional)
- [ ] Run test to verify config loading works

## Validation
Test configuration loading:
```bash
# Run tests
go test ./internal/config -v

# Test with your actual .env
go run -c "
package main
import (\"fmt\"; \"rulestack/internal/config\")
func main() {
    config.LoadEnvFile(\".env\")
    cfg := config.Load() 
    fmt.Println(\"Config loaded:\", cfg.APIPort)
}"
```

## Acceptance Criteria
- [ ] Config package loads environment variables correctly
- [ ] Required validation works (DATABASE_URL, TOKEN_SALT)
- [ ] Default values are applied properly
- [ ] CLI config (TOML) loading and saving works
- [ ] All tests pass
- [ ] Can load configuration from .env file
- [ ] Error handling for missing required config

## Time Estimate: ~60 minutes

## Next Task
Task 5: Database Connection and Models