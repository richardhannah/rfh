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
		// Create HTTP client using existing Client struct
		httpClient := NewClient(registry.URL, registry.JWTToken)
		httpClient.SetVerbose(verbose)

		// Wrap existing HTTP client to implement RegistryClient interface
		// This will be implemented in Phase 3
		return NewHTTPRegistryClient(httpClient), nil

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

// NewHTTPRegistryClient wraps the existing HTTP client to implement RegistryClient interface
// This will be implemented in Phase 3: HTTP Client Refactoring
func NewHTTPRegistryClient(httpClient *Client) RegistryClient {
	// This is a placeholder - will be implemented in Phase 3
	panic("NewHTTPRegistryClient not yet implemented - will be added in Phase 3")
}

// NewGitRegistryClient creates a new Git-based registry client
// This will be implemented in Phase 5: Git Client Implementation
func NewGitRegistryClient(repoURL, gitToken string, verbose bool) (RegistryClient, error) {
	// This is a placeholder - will be implemented in Phase 5
	return nil, fmt.Errorf("Git registry client not yet implemented - will be added in Phase 5")
}
