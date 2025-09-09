package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"rulestack/internal/client"
	"rulestack/internal/config"
)

// registryCmd represents the registry command
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage registries",
	Long: `Manage registry configurations for publishing and installing rulesets.

Registries are where rulesets are published and downloaded from. You can
configure multiple registries including public and private ones.`,
}

// registryAddCmd adds a new registry
var registryAddCmd = &cobra.Command{
	Use:   "add <name> <url> [--type remote-http|git]",
	Short: "Add a new registry",
	Long: `Add a new registry configuration.

Registry Types:
  remote-http - Traditional HTTP-based registry (default)
  git        - Git repository-based registry

Examples:
  rfh registry add public https://registry.rulestack.dev
  rfh registry add github https://github.com/org/registry --type git`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		url := args[1]
		registryType, _ := cmd.Flags().GetString("type")

		if registryType == "" {
			registryType = string(config.RegistryTypeHTTP)
		}

		return runRegistryAdd(name, url, config.RegistryType(registryType))
	},
}

// registryListCmd lists configured registries
var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured registries",
	Long:  `List all configured registries showing name, URL, and active status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRegistryList()
	},
}

// registryUseCmd sets the active registry
var registryUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set active registry",
	Long: `Set the active registry for publishing and installing packages.

The active registry is used when no --registry flag is specified.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRegistryUse(args[0])
	},
}

func runRegistryAdd(name, url string, registryType config.RegistryType) error {
	// Validate registry type
	if err := config.ValidateRegistryType(registryType); err != nil {
		return err
	}

	// Validate URL based on type
	if registryType == config.RegistryTypeGit {
		if !strings.HasPrefix(url, "https://github.com/") &&
			!strings.HasPrefix(url, "https://gitlab.com/") &&
			!strings.HasPrefix(url, "git@") {
			fmt.Printf("⚠️  Warning: Git registry URL may not be valid\n")
		}
	}

	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Add registry with type
	cfg.Registries[name] = config.Registry{
		URL:  url,
		Type: registryType,
	}

	// Set as current if it's the first one
	if cfg.Current == "" {
		cfg.Current = name
	}

	// Save config
	if err := config.SaveCLI(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Added registry '%s'\n", name)
	fmt.Printf("🌐 URL: %s\n", url)
	fmt.Printf("📋 Type: %s\n", registryType)

	if cfg.Current == name {
		fmt.Printf("⭐ Set as active registry\n")
	}

	if registryType == config.RegistryTypeHTTP {
		fmt.Printf("💡 Use 'rfh auth login' to authenticate with this registry\n")
	} else if registryType == config.RegistryTypeGit {
		fmt.Printf("💡 Set git_token in config or use GITHUB_TOKEN environment variable for authentication\n")
	}

	return nil
}

func runRegistryList() error {
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Registries) == 0 {
		fmt.Printf("No registries configured.\n")
		fmt.Printf("Add a registry with: rfh registry add <name> <url> [--type remote-http|git]\n")
		return nil
	}

	fmt.Printf("📋 Configured registries:\n\n")
	for name, reg := range cfg.Registries {
		marker := "  "
		if cfg.Current == name {
			marker = "* "
		}

		registryType := reg.GetEffectiveType()

		fmt.Printf("%s%s (%s)\n", marker, name, registryType)
		fmt.Printf("    URL: %s\n", reg.URL)

		// Show appropriate token status based on type
		if registryType == config.RegistryTypeHTTP && reg.JWTToken != "" {
			fmt.Printf("    JWT Token: [configured]\n")
		} else if registryType == config.RegistryTypeGit && reg.GitToken != "" {
			fmt.Printf("    Git Token: [configured]\n")
		}

		fmt.Printf("\n")
	}

	if cfg.Current != "" {
		fmt.Printf("* = active registry\n")
	}

	return nil
}

func runRegistryUse(name string) error {
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if registry exists
	if _, exists := cfg.Registries[name]; !exists {
		return fmt.Errorf("registry '%s' not found. Use 'rfh registry list' to see available registries", name)
	}

	// Set as current
	cfg.Current = name

	// Save config
	if err := config.SaveCLI(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Set '%s' as active registry\n", name)
	fmt.Printf("🌐 URL: %s\n", cfg.Registries[name].URL)

	return nil
}

// registryInitCmd initializes an empty Git registry
var registryInitCmd = &cobra.Command{
	Use:   "init --token <github-token>",
	Short: "Initialize empty Git registry with default structure",
	Long: `Initialize the active Git registry with the default package structure and store authentication token.

This command will:
1. Store the GitHub token for the active registry
2. Create initial repository structure (packages/, index.json, README.md)
3. Make an initial commit to the remote repository

The command operates on the currently active registry. Use 'rfh registry use <name>' to change the active registry.

Example:
  rfh registry add my-rules https://github.com/org/rules --type git
  rfh registry init --token ghp_xxxxxxxxxxxx`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		if token == "" {
			return fmt.Errorf("--token flag is required")
		}
		return runRegistryInit(token)
	},
}

// registryRemoveCmd removes a registry
var registryRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a registry",
	Long: `Remove a registry configuration.

Examples:
  rfh registry remove old-registry
  rfh registry remove test`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRegistryRemove(args[0])
	},
}

func runRegistryRemove(name string) error {
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if registry exists
	if _, exists := cfg.Registries[name]; !exists {
		return fmt.Errorf("registry '%s' not found. Use 'rfh registry list' to see available registries", name)
	}

	// Store URL for display
	url := cfg.Registries[name].URL

	// Remove the registry
	delete(cfg.Registries, name)

	// If this was the current registry, clear the current setting
	if cfg.Current == name {
		cfg.Current = ""
		fmt.Printf("⚠️  Removed active registry. Use 'rfh registry use' to set a new active registry.\n")
	}

	// Save config
	if err := config.SaveCLI(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Removed registry '%s'\n", name)
	fmt.Printf("🌐 URL was: %s\n", url)

	return nil
}

func runRegistryInit(token string) error {
	// 1. Load config
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Get current active registry
	registryName, registry, err := getCurrentRegistry(cfg)
	if err != nil {
		return err
	}

	// 3. Validate registry is Git type
	if registry.GetEffectiveType() != config.RegistryTypeGit {
		return fmt.Errorf("active registry '%s' is not a Git registry (type: %s). Only Git registries can be initialized", registryName, registry.GetEffectiveType())
	}

	fmt.Printf("🔧 Initializing Git registry '%s'...\n", registryName)
	fmt.Printf("🌐 URL: %s\n", registry.URL)

	// 4. Store token in config and save immediately
	registry.GitToken = token
	cfg.Registries[registryName] = registry
	
	if err := config.SaveCLI(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("🔑 Token stored in config\n")

	// 5. Initialize repository structure
	err = initializeGitRegistryStructure(registryName, &registry)
	if err != nil {
		// Token is already saved, so give appropriate feedback
		fmt.Printf("⚠️  Repository structure initialization failed: %v\n", err)
		fmt.Printf("💡 Token has been saved. You can retry initialization later.\n")
		return fmt.Errorf("failed to initialize registry structure: %w", err)
	}

	// 6. Success feedback
	fmt.Printf("✅ Registry '%s' initialized successfully\n", registryName)
	fmt.Printf("📁 Default structure created in remote repository\n")
	fmt.Printf("💡 Registry is now ready for publishing packages\n")

	return nil
}

func initializeGitRegistryStructure(registryName string, registry *config.Registry) error {
	// Create temporary GitClient with the token
	c, err := client.NewGitClient(registry.URL, registry.GitToken, verbose)
	if err != nil {
		return fmt.Errorf("failed to create Git client: %w", err)
	}

	ctx, cancel := client.WithTimeout(context.Background())
	defer cancel()

	fmt.Printf("🚀 Setting up repository structure...\n")
	return c.InitializeRegistry(ctx)
}

func init() {
	registryAddCmd.Flags().String("type", "remote-http", "Registry type (remote-http or git)")
	registryInitCmd.Flags().String("token", "", "GitHub personal access token (required)")

	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryUseCmd)
	registryCmd.AddCommand(registryInitCmd)
	registryCmd.AddCommand(registryRemoveCmd)
}
