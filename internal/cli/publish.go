package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"rulestack/internal/client"
	"rulestack/internal/config"
	"rulestack/internal/manifest"
)


// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish staged rulesets to the registry",
	Long: `Publish all staged ruleset packages to the configured registry.

This command will:
1. Scan .rulestack/staged/ directory for archives
2. Read associated manifest data from rulestack.json
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

	// Load manifests to get package information
	manifests, err := manifest.LoadAll("rulestack.json")
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	fmt.Printf("Found %d staged archive(s) to publish:\n", len(archives))
	for _, archivePath := range archives {
		fmt.Printf("  - %s\n", filepath.Base(archivePath))
	}

	// Publish each archive
	successCount := 0
	for _, archivePath := range archives {
		if err := publishSingleArchive(archivePath, manifests); err != nil {
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
func publishSingleArchive(archivePath string, manifests manifest.ManifestFile) error {
	// Extract package name and version from archive filename
	archiveName := filepath.Base(archivePath)
	archiveName = strings.TrimSuffix(archiveName, ".tgz")
	
	// Find matching manifest entry
	var packageManifest *manifest.Manifest
	for _, m := range manifests {
		expectedName := fmt.Sprintf("%s-%s", m.Name, m.Version)
		if expectedName == archiveName {
			packageManifest = &m
			break
		}
	}
	
	if packageManifest == nil {
		return fmt.Errorf("no manifest found for archive: %s", archiveName)
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

	// Get effective token (flag, registry token, or JWT token)
	authToken, err := getEffectiveToken(cfg, reg)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("📦 Publishing %s v%s\n", packageManifest.Name, packageManifest.Version)
		fmt.Printf("🌐 Registry: %s (%s)\n", registryName, reg.URL)
		fmt.Printf("📄 Archive: %s\n", archivePath)
	}

	// Create client
	c := client.NewClient(reg.URL, authToken)
	c.SetVerbose(verbose)

	// Test registry connection
	if err := c.Health(); err != nil {
		return fmt.Errorf("registry health check failed: %w", err)
	}

	// Create a temporary manifest file for this specific package (as single object, not array)
	tempManifestPath := fmt.Sprintf(".rulestack/staged/temp-manifest-%s.json", archiveName)
	if err := createSingleManifestFile(packageManifest, tempManifestPath); err != nil {
		return fmt.Errorf("failed to create temp manifest: %w", err)
	}
	defer os.Remove(tempManifestPath) // Clean up temp file

	// Publish package
	fmt.Printf("🚀 Publishing %s v%s to %s...\n", packageManifest.Name, packageManifest.Version, reg.URL)
	result, err := c.PublishPackage(tempManifestPath, archivePath)
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	// Show success message
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
func createSingleManifestFile(packageManifest *manifest.Manifest, filePath string) error {
	// Create the manifest as a single object (not array) for API compatibility
	data, err := json.MarshalIndent(packageManifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return os.WriteFile(filePath, data, 0o644)
}

func init() {
}