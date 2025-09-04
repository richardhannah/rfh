package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type RegistryType string

const (
	RegistryTypeHTTP RegistryType = "remote-http"
	RegistryTypeGit  RegistryType = "git"
)

type Registry struct {
	URL      string       `toml:"url"`
	Type     RegistryType `toml:"type"`                   // New field
	Username string       `toml:"username,omitempty"`     // Username for this registry
	JWTToken string       `toml:"jwt_token,omitempty"`    // JWT token for this registry
	GitToken string       `toml:"git_token,omitempty"`    // New field for git auth
}

type CLIConfig struct {
	Current    string              `toml:"current"`
	Registries map[string]Registry `toml:"registries"`
}

// ConfigDir returns the CLI config directory path
// It first checks the RFH_CONFIG environment variable for a custom config location.
// If not set, it falls back to the default ~/.rfh directory.
func ConfigDir() (string, error) {
	// Check for RFH_CONFIG environment variable first
	if rfhConfig := os.Getenv("RFH_CONFIG"); rfhConfig != "" {
		return rfhConfig, nil
	}
	
	// Fall back to default ~/.rfh location
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

	// Migrate existing registries to have explicit type
	for name, reg := range config.Registries {
		if reg.Type == "" {
			reg.Type = RegistryTypeHTTP
			config.Registries[name] = reg
		}
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

// ValidateRegistryType checks if a registry type is valid
func ValidateRegistryType(t RegistryType) error {
	switch t {
	case RegistryTypeHTTP, RegistryTypeGit:
		return nil
	default:
		return fmt.Errorf("unsupported registry type: %s", t)
	}
}

// GetEffectiveType returns the effective type for a registry
func (r Registry) GetEffectiveType() RegistryType {
	if r.Type == "" {
		return RegistryTypeHTTP
	}
	return r.Type
}
