package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"rulestack/internal/client"
	"rulestack/internal/config"
	"rulestack/internal/manifest"
	"rulestack/internal/pkg"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish staged rulesets to the registry",
	Long: `Publish all staged ruleset packages to the configured registry.

This command will:
1. Scan .rulestack/staged/ directory for archives
2. Extract embedded manifest data from each archive
3. Upload each archive to the registry
4. Clean up staged archives after successful upload

Archives must be created with 'rfh pack' command first.
Requires authentication token to be configured in the registry.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPublishStaged()
	},
}

func runPublishStaged() error {
	stagingDir := ".rulestack/staged"

	// Check if staging directory exists
	if _, err := os.Stat(stagingDir); os.IsNotExist(err) {
		return fmt.Errorf("no staged archives found. Use 'rfh pack' to create archives first")
	}

	// Find all .tgz files in staging directory
	archives, err := filepath.Glob(filepath.Join(stagingDir, "*.tgz"))
	if err != nil {
		return fmt.Errorf("failed to scan staging directory: %w", err)
	}

	if len(archives) == 0 {
		return fmt.Errorf("no archives found in staging directory. Use 'rfh pack' to create archives first")
	}

	fmt.Printf("Found %d staged archive(s) to publish:\n", len(archives))
	for _, archivePath := range archives {
		fmt.Printf("  - %s\n", filepath.Base(archivePath))
	}

	// Publish each archive
	successCount := 0
	for _, archivePath := range archives {
		if err := publishSingleArchive(archivePath); err != nil {
			fmt.Printf("❌ Failed to publish %s: %v\n", filepath.Base(archivePath), err)
		} else {
			fmt.Printf("✅ Successfully published %s\n", filepath.Base(archivePath))
			// Remove archive after successful publish
			os.Remove(archivePath)
			successCount++
		}
	}

	if successCount == len(archives) {
		fmt.Printf("\n🎉 All %d archive(s) published successfully!\n", successCount)
		return nil
	} else {
		fmt.Printf("\n⚠️  Published %d out of %d archive(s)\n", successCount, len(archives))
		return fmt.Errorf("failed to publish %d archive(s)", len(archives)-successCount)
	}
}

// publishSingleArchive publishes a single archive file
func publishSingleArchive(archivePath string) error {
	// Extract manifest from archive
	manifestData, err := pkg.ExtractManifest(archivePath)
	if err != nil {
		return fmt.Errorf("failed to extract manifest from archive: %w", err)
	}

	// Parse the manifest
	var packageManifest manifest.PackageManifest
	if err := json.Unmarshal(manifestData, &packageManifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Check if archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return fmt.Errorf("archive not found: %s", archivePath)
	}

	// Get registry configuration
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current registry
	registryName, reg, err := getCurrentRegistry(cfg)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("📦 Publishing %s v%s\n", packageManifest.Name, packageManifest.Version)
		fmt.Printf("🌐 Registry: %s (%s)\n", registryName, reg.URL)
		fmt.Printf("📄 Archive: %s\n", archivePath)
	}

	// Create client using new factory
	c, err := client.GetClient(cfg, verbose)
	if err != nil {
		return err
	}

	// Test registry connection
	ctx, cancel := client.WithTimeout(context.Background())
	defer cancel()
	
	if err := c.Health(ctx); err != nil {
		return fmt.Errorf("registry health check failed: %w", err)
	}

	// Create a temporary manifest file for this specific package (as single object, not array)
	archiveName := strings.TrimSuffix(filepath.Base(archivePath), ".tgz")
	tempManifestPath := fmt.Sprintf(".rulestack/staged/temp-manifest-%s.json", archiveName)
	if err := createSingleManifestFile(&packageManifest, tempManifestPath); err != nil {
		return fmt.Errorf("failed to create temp manifest: %w", err)
	}
	defer os.Remove(tempManifestPath) // Clean up temp file

	// Publish package
	fmt.Printf("🚀 Publishing %s v%s to %s...\n", packageManifest.Name, packageManifest.Version, reg.URL)
	result, err := c.PublishPackage(ctx, tempManifestPath, archivePath)
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	// Show success message
	fmt.Printf("📌 Version: %s\n", result.Version)
	fmt.Printf("🔒 SHA256: %s\n", result.SHA256)

	if verbose {
		fmt.Printf("📋 Response: %+v\n", result)
	}

	return nil
}

// sanitizePackageName removes characters that are invalid in filenames
func sanitizePackageName(name string) string {
	// Replace invalid filename characters with safe alternatives
	safeName := strings.ReplaceAll(name, "@", "")
	safeName = strings.ReplaceAll(safeName, "/", "-")
	safeName = strings.ReplaceAll(safeName, "\\", "-")
	safeName = strings.ReplaceAll(safeName, ":", "-")
	safeName = strings.ReplaceAll(safeName, "*", "-")
	safeName = strings.ReplaceAll(safeName, "?", "-")
	safeName = strings.ReplaceAll(safeName, "\"", "-")
	safeName = strings.ReplaceAll(safeName, "<", "-")
	safeName = strings.ReplaceAll(safeName, ">", "-")
	safeName = strings.ReplaceAll(safeName, "|", "-")

	// Remove any leading/trailing spaces or dashes
	safeName = strings.Trim(safeName, " -")

	// Ensure we have a valid name
	if safeName == "" {
		safeName = "unnamed-package"
	}

	return safeName
}

// createSingleManifestFile creates a temporary manifest file with a single package entry
func createSingleManifestFile(packageManifest *manifest.PackageManifest, filePath string) error {
	// Create the manifest as a single object (not array) for API compatibility
	data, err := json.MarshalIndent(packageManifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return os.WriteFile(filePath, data, 0o644)
}

func init() {
}
