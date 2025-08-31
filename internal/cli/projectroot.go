package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"rulestack/internal/manifest"
)

// projectrootCmd represents the projectroot command (temporary diagnostic tool)
var projectrootCmd = &cobra.Command{
	Use:   "projectroot",
	Short: "Diagnostic tool to show project root discovery information",
	Long: `Shows detailed information about how RFH discovers and uses project roots.

This is a temporary diagnostic command to help troubleshoot path resolution issues.

The command displays:
1. Current working directory (where the command was invoked)
2. Location of the closest rulestack.json file found by walking up the directory tree
3. The registered project root from the rulestack.json file (if it exists)

This helps identify discrepancies between where commands are run and where RFH thinks
the project root should be.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectRootDiagnostic()
	},
}

func runProjectRootDiagnostic() error {
	// 1. Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	fmt.Printf("📁 Current Working Directory: %s\n", cwd)

	// 2. Find closest rulestack.json using existing logic
	projectRoot, rulestackPath, err := findProjectRootWithPath()
	if err != nil {
		fmt.Printf("❌ No rulestack.json found in directory tree\n")
		fmt.Printf("   Error: %v\n", err)
		return nil // Don't error out, this is diagnostic
	}

	fmt.Printf("📄 Closest rulestack.json: %s\n", rulestackPath)

	// 3. Read the rulestack.json and extract project root if it exists
	registeredRoot, err := getRegisteredProjectRoot(rulestackPath)
	if err != nil {
		fmt.Printf("⚠️  Failed to read project root from rulestack.json: %v\n", err)
	} else if registeredRoot == "" {
		fmt.Printf("📋 Registered Project Root: (not specified - this is a package manifest)\n")
	} else {
		fmt.Printf("🎯 Registered Project Root: %s\n", registeredRoot)
	}

	// 4. Show comparison
	fmt.Printf("\n--- Analysis ---\n")
	if projectRoot != cwd {
		fmt.Printf("⚠️  Working directory differs from discovered project root\n")
		fmt.Printf("   This may cause path resolution issues\n")
	} else {
		fmt.Printf("✅ Working directory matches discovered project root\n")
	}

	if registeredRoot != "" && registeredRoot != projectRoot {
		fmt.Printf("⚠️  Registered project root differs from discovered location\n")
		fmt.Printf("   This indicates a project manifest with explicit projectRoot\n")
	}

	return nil
}

// findProjectRootWithPath is like findProjectRoot but also returns the path to the rulestack.json file
func findProjectRootWithPath() (string, string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up the directory tree looking for rulestack.json
	for {
		manifestPath := filepath.Join(dir, "rulestack.json")
		if _, err := os.Stat(manifestPath); err == nil {
			return dir, manifestPath, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	return "", "", fmt.Errorf("no rulestack.json found in directory tree")
}

// getRegisteredProjectRoot reads the rulestack.json file and extracts the projectRoot field if it exists
func getRegisteredProjectRoot(rulestackPath string) (string, error) {
	// Try to parse as project manifest first (has projectRoot field)
	projectManifest, err := manifest.LoadProjectManifest(rulestackPath)
	if err == nil && projectManifest.ProjectRoot != "" {
		return projectManifest.ProjectRoot, nil
	}

	// If that fails, it might be a package manifest (array format) - no projectRoot field
	// We can use the manifest type detection functions
	if manifest.IsPackageManifest(rulestackPath) {
		return "", nil // Package manifests don't have projectRoot
	}

	// If we can't determine the type or there was an error, return the error
	return "", err
}

func init() {
	rootCmd.AddCommand(projectrootCmd)
}