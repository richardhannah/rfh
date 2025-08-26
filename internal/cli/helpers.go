package cli

import (
	"fmt"
	"rulestack/internal/config"
)

// getEffectiveToken returns the token to use for API calls
// Priority: 1) --token flag, 2) registry JWT token, 3) global JWT token (deprecated)
func getEffectiveToken(cfg config.CLIConfig, registry config.Registry) (string, error) {
	// 1. Check command line flag (highest priority)
	if token != "" {
		if verbose {
			fmt.Printf("🔍 Using token from --token flag (length: %d chars)\n", len(token))
		}
		return token, nil
	}

	// 2. Check registry-specific JWT token
	if registry.JWTToken != "" {
		if verbose {
			fmt.Printf("🔍 Using JWT token from registry config (length: %d chars)\n", len(registry.JWTToken))
		}
		return registry.JWTToken, nil
	}

	// 3. Check global JWT token from user authentication (deprecated, for backward compatibility)
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

	// Allow registry override from flag
	if registry != "" {
		// If registry flag is provided, we need to find it or use it as URL
		if reg, exists := cfg.Registries[registry]; exists {
			return registry, reg, nil
		}
		// Treat registry flag as URL if not found in config
		return "override", config.Registry{URL: registry}, nil
	}

	if registryName == "" {
		return "", config.Registry{}, fmt.Errorf("no active registry configured. Use 'rfh registry add' to add one")
	}

	reg, exists := cfg.Registries[registryName]
	if !exists {
		return "", config.Registry{}, fmt.Errorf("active registry '%s' not found", registryName)
	}

	return registryName, reg, nil
}