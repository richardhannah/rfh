package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"rulestack/internal/client"
	"rulestack/internal/config"
	"rulestack/internal/manifest"
)

var (
	archivePath string
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a ruleset to the registry",
	Long: `Publish a ruleset package to the configured registry.

This command will:
1. Read the rulestack.json manifest
2. Use the specified archive (or create one if not specified)
3. Upload both files to the registry
4. Validate the upload was successful

Requires authentication token to be configured in the registry.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPublish()
	},
}

func runPublish() error {
	// Load manifest
	manifest, err := manifest.Load("rulestack.json")
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Determine archive path
	archive := archivePath
	if archive == "" {
		// Generate default archive name
		safeName := manifest.GetPackageName()
		if scope := manifest.GetScope(); scope != "" {
			safeName = scope + "-" + safeName
		}
		archive = fmt.Sprintf("%s-%s.tgz", safeName, manifest.Version)
	}

	// Check if archive exists
	if _, err := os.Stat(archive); os.IsNotExist(err) {
		return fmt.Errorf("archive not found: %s. Run 'rfh pack' first or specify --archive", archive)
	}

	// Get registry configuration
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine which registry to use
	registryName := cfg.Current
	if registry != "" {
		registryName = registry
	}

	if registryName == "" {
		return fmt.Errorf("no registry configured. Use 'rfh registry add' to add a registry")
	}

	reg, exists := cfg.Registries[registryName]
	if !exists {
		return fmt.Errorf("registry '%s' not found. Use 'rfh registry list' to see available registries", registryName)
	}

	// Use token from flag or config
	authToken := reg.Token
	if token != "" {
		authToken = token
	}

	if authToken == "" {
		return fmt.Errorf("no authentication token configured for registry '%s'", registryName)
	}

	if verbose {
		fmt.Printf("📦 Publishing %s v%s\n", manifest.Name, manifest.Version)
		fmt.Printf("🌐 Registry: %s (%s)\n", registryName, reg.URL)
		fmt.Printf("📄 Archive: %s\n", archive)
	}

	// Create client
	c := client.NewClient(reg.URL, authToken)
	c.SetVerbose(verbose)

	// Test registry connection
	if err := c.Health(); err != nil {
		return fmt.Errorf("registry health check failed: %w", err)
	}

	// Publish package
	fmt.Printf("🚀 Publishing to %s...\n", reg.URL)
	result, err := c.PublishPackage("rulestack.json", archive)
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	// Show success message
	fmt.Printf("✅ Successfully published %s\n", manifest.Name)
	if version, ok := result["version"].(string); ok {
		fmt.Printf("📌 Version: %s\n", version)
	}
	if sha, ok := result["sha256"].(string); ok {
		fmt.Printf("🔒 SHA256: %s\n", sha)
	}

	if verbose {
		fmt.Printf("📋 Response: %+v\n", result)
	}

	return nil
}

func init() {
	publishCmd.Flags().StringVarP(&archivePath, "archive", "a", "", "path to archive file (defaults to <name>-<version>.tgz)")
}