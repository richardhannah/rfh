package cli

import (
	"fmt"
	"rulestack/internal/config"
)

// getEffectiveToken returns the token to use for API calls
// Priority: 1) registry JWT token, 2) global JWT token (deprecated)
func getEffectiveToken(cfg config.CLIConfig, registry config.Registry) (string, error) {
	// 1. Check registry-specific JWT token
	if registry.JWTToken != "" {
		if verbose {
			fmt.Printf("🔍 Using JWT token from registry config (length: %d chars)\n", len(registry.JWTToken))
		}
		return registry.JWTToken, nil
	}

	// 2. Check global JWT token from user authentication (deprecated, for backward compatibility)
	if cfg.User != nil && cfg.User.Token != "" {
		if verbose {
			fmt.Printf("🔍 Using global JWT token (DEPRECATED) (length: %d chars)\n", len(cfg.User.Token))
		}
		return cfg.User.Token, nil
	}

	return "", fmt.Errorf("no authentication token available. Use 'rfh auth login' to authenticate or configure a registry JWT token")
}

// getCurrentRegistry returns the current active registry
func getCurrentRegistry(cfg config.CLIConfig) (string, config.Registry, error) {
	registryName := cfg.Current

	if registryName == "" {
		return "", config.Registry{}, fmt.Errorf("no active registry configured. Use 'rfh registry add' to add one")
	}

	reg, exists := cfg.Registries[registryName]
	if !exists {
		return "", config.Registry{}, fmt.Errorf("active registry '%s' not found", registryName)
	}

	return registryName, reg, nil
}

// getDefaultToken returns the default token for a registry (no command line overrides)
func getDefaultToken(registry config.Registry) string {
	// Use registry-specific JWT token
	if registry.JWTToken != "" {
		if verbose {
			fmt.Printf("🔍 Using JWT token from registry config (length: %d chars)\n", len(registry.JWTToken))
		}
		return registry.JWTToken
	}

	// No token available - return empty string (will cause auth error)
	return ""
}