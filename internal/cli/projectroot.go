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

	// 3. Determine manifest type
	manifestType := "unknown"
	if manifest.IsProjectManifest(rulestackPath) {
		manifestType = "project manifest (dependency management)"
	} else if manifest.IsPackageManifest(rulestackPath) {
		manifestType = "package manifest (array of packages)"
	}
	fmt.Printf("📋 Manifest Type: %s\n", manifestType)

	// 4. Show comparison
	fmt.Printf("\n--- Analysis ---\n")
	if projectRoot != cwd {
		fmt.Printf("⚠️  Working directory differs from discovered project root\n")
		fmt.Printf("   This may cause path resolution issues\n")
	} else {
		fmt.Printf("✅ Working directory matches discovered project root\n")
	}

	fmt.Printf("ℹ️  Project root is determined by walking up directory tree to find rulestack.json\n")
	fmt.Printf("   The projectRoot field has been removed as it was not functionally used\n")

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


func init() {
	rootCmd.AddCommand(projectrootCmd)
}