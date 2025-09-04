package client

import (
	"fmt"
	"rulestack/internal/config"
)

// NewRegistryClient creates the appropriate client based on registry type
func NewRegistryClient(registry config.Registry, verbose bool) (RegistryClient, error) {
	registryType := registry.GetEffectiveType()

	switch registryType {
	case config.RegistryTypeHTTP:
		// Create new HTTP client that implements RegistryClient interface
		return NewHTTPClient(registry.URL, registry.JWTToken, verbose), nil

	case config.RegistryTypeGit:
		// Git client will be implemented in later phases
		return NewGitRegistryClient(registry.URL, registry.GitToken, verbose)

	default:
		return nil, fmt.Errorf("unsupported registry type: %s", registryType)
	}
}

// GetClient creates a client for the current active registry
func GetClient(cfg config.CLIConfig, verbose bool) (RegistryClient, error) {
	if cfg.Current == "" {
		return nil, fmt.Errorf("no active registry configured")
	}

	registry, exists := cfg.Registries[cfg.Current]
	if !exists {
		return nil, fmt.Errorf("active registry '%s' not found in configuration", cfg.Current)
	}

	return NewRegistryClient(registry, verbose)
}

// GetClientForRegistry creates a client for a specific named registry
func GetClientForRegistry(cfg config.CLIConfig, registryName string, verbose bool) (RegistryClient, error) {
	registry, exists := cfg.Registries[registryName]
	if !exists {
		return nil, fmt.Errorf("registry '%s' not found", registryName)
	}

	return NewRegistryClient(registry, verbose)
}

// Placeholder functions for clients that will be implemented in later phases


// NewGitRegistryClient creates a new Git-based registry client
func NewGitRegistryClient(repoURL, gitToken string, verbose bool) (RegistryClient, error) {
	return NewGitClient(repoURL, gitToken, verbose)
}
