package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Registry struct {
	URL      string `toml:"url"`
	Username string `toml:"username,omitempty"` // Username for this registry
	JWTToken string `toml:"jwt_token,omitempty"` // JWT token for this registry
}

type User struct {
	Username string `toml:"username"`
	Token    string `toml:"token"` // JWT token
}

type CLIConfig struct {
	Current    string              `toml:"current"`
	Registries map[string]Registry `toml:"registries"`
	User       *User               `toml:"user,omitempty"` // DEPRECATED: Legacy global user config
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

	// Migrate global user config to per-registry config if needed
	config = migrateToPerRegistryAuth(config)

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

// migrateToPerRegistryAuth migrates global user config to per-registry auth
func migrateToPerRegistryAuth(config CLIConfig) CLIConfig {
	// If we have a global user config and a current registry, migrate it
	if config.User != nil && config.Current != "" {
		if registry, exists := config.Registries[config.Current]; exists {
			// Only migrate if the registry doesn't already have per-registry auth
			if registry.Username == "" && registry.JWTToken == "" {
				registry.Username = config.User.Username
				registry.JWTToken = config.User.Token
				config.Registries[config.Current] = registry
				
				// Keep the global user config for backward compatibility
				// It will be used as fallback by getEffectiveToken
			}
		}
	}
	
	return config
}