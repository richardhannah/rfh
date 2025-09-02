package cli

import (
	"fmt"
	"strings"
	"rulestack/internal/config"
)

// getEffectiveToken returns the token to use for API calls
func getEffectiveToken(cfg config.CLIConfig, registry config.Registry) (string, error) {
	// Check registry-specific JWT token
	if registry.JWTToken != "" {
		if verbose {
			fmt.Printf("🔍 Using JWT token from registry config (length: %d chars)\n", len(registry.JWTToken))
		}
		return registry.JWTToken, nil
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

// checkAndWarnRootUser displays a security warning if the current user is logged in as 'root'
func checkAndWarnRootUser(cfg config.CLIConfig, commandName string) {
	// Skip warning for auth-related commands to avoid spam during authentication workflows
	if isAuthCommand(commandName) {
		return
	}

	// Check if user is logged in as root in the current registry
	if cfg.Current != "" {
		if registry, exists := cfg.Registries[cfg.Current]; exists && strings.ToLower(registry.Username) == "root" {
			fmt.Printf("\n⚠️  🚨 SECURITY WARNING 🚨 ⚠️\n")
			fmt.Printf("YOU ARE LOGGED IN AS ROOT USER!\n\n")
			fmt.Printf("This is a high-privilege administrative account that should NOT be used for regular operations.\n\n")
			fmt.Printf("RECOMMENDED ACTIONS:\n")
			fmt.Printf("1. Create a regular user account: rfh auth register\n")
			fmt.Printf("2. Grant admin privileges to your user account\n")
			fmt.Printf("3. Disable or change the root account password\n")
			fmt.Printf("4. Use your regular account for daily operations\n\n")
			fmt.Printf("This warning appears for all commands when logged in as 'root'.\n")
			fmt.Printf("═════════════════════════════════════════════════════════════════\n\n")
		}
	}
}

// isAuthCommand checks if the command is related to authentication (to skip warnings)
func isAuthCommand(commandName string) bool {
	authCommands := []string{
		"auth",
		"auth login",
		"auth logout", 
		"auth register",
		"auth whoami",
	}
	
	for _, cmd := range authCommands {
		if strings.HasPrefix(commandName, cmd) {
			return true
		}
	}
	
	return false
}