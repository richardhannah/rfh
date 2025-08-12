package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"rulestack/internal/manifest"
	"rulestack/internal/pkg"
)

var (
	outputPath string
)

// packCmd represents the pack command
var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Pack ruleset files into a distributable archive",
	Long: `Creates a tar.gz archive containing all files specified in the manifest.
The archive is ready for publishing to a registry.

The pack command:
1. Reads rulestack.json manifest
2. Collects all files matching the patterns in 'files' array
3. Creates a compressed archive
4. Calculates SHA256 hash for integrity`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPack()
	},
}

func runPack() error {
	// Load manifest
	manifest, err := manifest.Load("rulestack.json")
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Determine output path
	output := outputPath
	if output == "" {
		// Remove @ and / from package name for filename
		safeName := manifest.GetPackageName()
		if scope := manifest.GetScope(); scope != "" {
			safeName = scope + "-" + safeName
		}
		output = fmt.Sprintf("%s-%s.tgz", safeName, manifest.Version)
	}

	if verbose {
		fmt.Printf("📦 Packing %s v%s\n", manifest.Name, manifest.Version)
		fmt.Printf("🎯 Targets: %v\n", manifest.Targets)
		fmt.Printf("🏷️  Tags: %v\n", manifest.Tags)
		fmt.Printf("📄 File patterns: %v\n", manifest.Files)
	}

	// Pack files
	info, err := pkg.Pack(manifest.Files, output)
	if err != nil {
		return fmt.Errorf("failed to pack files: %w", err)
	}

	fmt.Printf("✅ Successfully packed %s\n", manifest.Name)
	fmt.Printf("📦 Archive: %s\n", info.Path)
	fmt.Printf("📏 Size: %d bytes\n", info.SizeBytes)
	fmt.Printf("🔒 SHA256: %s\n", info.SHA256)

	if verbose {
		// List files that were included
		fmt.Printf("\n📋 Files included:\n")
		// This would require modifying Pack to return file list
		// For now, just show the patterns
		for _, pattern := range manifest.Files {
			fmt.Printf("   - %s\n", pattern)
		}
	}

	return nil
}

func init() {
	packCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output archive path")
}