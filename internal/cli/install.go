package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"rulestack/internal/client"
	"rulestack/internal/config"
	"rulestack/internal/manifest"
	"rulestack/internal/pkg"
	"rulestack/internal/version"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install .",
	Short: "Install all packages from project manifest",
	Long: `Install packages listed in rulestack.json dependencies.

This command reads the rulestack.json project manifest and ensures all dependencies
are installed with the correct versions.

Operations:
- Installs missing packages
- Updates packages to higher versions specified in manifest
- Skips packages that are already up-to-date
- Reports failures but continues processing other packages

Examples:
  rfh install .`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "." {
			return fmt.Errorf("only '.' is supported (current directory)")
		}
		return runInstall()
	},
}

// InstallResult represents the result of installing a single package
type InstallResult struct {
	Package string
	Version string
	Status  string // "installed", "updated", "skipped", "failed"
	Error   error
	Details string // Additional details about the operation
}

// PackageRequirement represents a package that needs to be processed
type PackageRequirement struct {
	Name             string
	RequiredVersion  string
	InstalledVersion string
	Action           string // "install", "update", "skip"
	PackageDir       string // Path to installed package directory
	Details          string // Additional details about the operation
}

// runInstall implements the install command logic
func runInstall() error {
	if verbose {
		fmt.Printf("📦 Installing packages from project manifest...\n")
	}

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	if verbose {
		fmt.Printf("📁 Project root: %s\n", projectRoot)
	}

	// Load project manifest
	manifestPath := filepath.Join(projectRoot, "rulestack.json")
	projectManifest, err := manifest.LoadProjectManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load project manifest: %w", err)
	}

	if len(projectManifest.Dependencies) == 0 {
		fmt.Printf("ℹ️  No dependencies found in rulestack.json\n")
		return nil
	}

	// Analyze package requirements
	requirements, err := analyzePackageRequirements(projectRoot, projectManifest.Dependencies)
	if err != nil {
		return fmt.Errorf("failed to analyze package requirements: %w", err)
	}

	// Process all packages
	results := processPackages(projectRoot, requirements)

	// Report results
	reportInstallResults(results)

	return nil
}

// analyzePackageRequirements compares manifest dependencies with installed packages
func analyzePackageRequirements(projectRoot string, dependencies map[string]string) ([]PackageRequirement, error) {
	requirements := []PackageRequirement{}
	rulestackDir := filepath.Join(projectRoot, ".rulestack")

	for packageName, requiredVersion := range dependencies {
		req := PackageRequirement{
			Name:            packageName,
			RequiredVersion: requiredVersion,
		}

		// Check if package is already installed
		installedVersion, packageDir, err := findInstalledPackage(rulestackDir, packageName)
		if err != nil {
			req.Action = "install"
			req.Details = "Package not installed"
		} else {
			req.InstalledVersion = installedVersion
			req.PackageDir = packageDir

			// Compare versions
			comparison, err := version.CompareVersions(installedVersion, requiredVersion)
			if err != nil {
				req.Action = "install"
				req.Details = fmt.Sprintf("Version comparison failed: %v", err)
			} else if comparison < 0 {
				req.Action = "update"
				req.Details = fmt.Sprintf("Installed: %s → Required: %s", installedVersion, requiredVersion)
			} else if comparison == 0 {
				req.Action = "skip"
				req.Details = "Already up-to-date"
			} else {
				req.Action = "skip"
				req.Details = fmt.Sprintf("Installed version %s is newer than required %s", installedVersion, requiredVersion)
			}
		}

		requirements = append(requirements, req)
	}

	return requirements, nil
}

// findInstalledPackage finds if a package is installed and returns its version and directory
func findInstalledPackage(rulestackDir, packageName string) (string, string, error) {
	// Look for directories matching pattern: packagename.version
	entries, err := os.ReadDir(rulestackDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("package not installed")
		}
		return "", "", fmt.Errorf("failed to read .rulestack directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse directory name: packagename.version
		dirName := entry.Name()
		parts := strings.Split(dirName, ".")
		if len(parts) < 2 {
			continue
		}

		// Reconstruct package name (everything except the last 3 parts which are version)
		if len(parts) >= 4 {
			// packagename.1.2.3
			candidateName := strings.Join(parts[:len(parts)-3], ".")
			candidateVersion := strings.Join(parts[len(parts)-3:], ".")

			if candidateName == packageName {
				packageDir := filepath.Join(rulestackDir, dirName)
				return candidateVersion, packageDir, nil
			}
		}
	}

	return "", "", fmt.Errorf("package not installed")
}

// processPackages processes all package requirements and returns results
func processPackages(projectRoot string, requirements []PackageRequirement) []InstallResult {
	results := []InstallResult{}

	for _, req := range requirements {
		result := InstallResult{
			Package: req.Name,
			Version: req.RequiredVersion,
		}

		switch req.Action {
		case "skip":
			result.Status = "skipped"
			result.Details = req.Details
		case "install", "update":
			err := installSinglePackage(projectRoot, req.Name, req.RequiredVersion)
			if err != nil {
				result.Status = "failed"
				result.Error = err
				result.Details = err.Error()
			} else {
				if req.Action == "install" {
					result.Status = "installed"
					result.Details = "Successfully installed"
				} else {
					result.Status = "updated"
					result.Details = fmt.Sprintf("Updated from %s", req.InstalledVersion)
				}
			}
		}

		results = append(results, result)
	}

	return results
}

// installSinglePackage installs a single package (extracted from add command logic)
func installSinglePackage(projectRoot, packageName, packageVersion string) error {
	// Create package reference
	pkgRef := &PackageRef{
		Name:    packageName,
		Version: packageVersion,
	}

	if verbose {
		fmt.Printf("📦 Installing %s@%s...\n", pkgRef.FullName(), pkgRef.Version)
	}

	// Get registry configuration
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	registryName := cfg.Current
	if registryName == "" {
		return fmt.Errorf("no registry configured. Use 'rfh registry add' to add a registry")
	}

	reg, exists := cfg.Registries[registryName]
	if !exists {
		return fmt.Errorf("registry '%s' not found. Use 'rfh registry list' to see available registries", registryName)
	}

	authToken := getDefaultToken(reg)

	// Create client
	c := client.NewClient(reg.URL, authToken)
	c.SetVerbose(verbose)

	// Get package version info
	versionInfo, err := c.GetPackageVersion(pkgRef.Name, pkgRef.Version)
	if err != nil {
		return fmt.Errorf("failed to get package version: %w", err)
	}

	// Extract SHA256 from version info
	sha256, ok := versionInfo["sha256"].(string)
	if !ok {
		return fmt.Errorf("package version missing sha256 hash")
	}

	// Create .rulestack directory if it doesn't exist
	rulestackDir := filepath.Join(projectRoot, ".rulestack")
	if err := os.MkdirAll(rulestackDir, 0755); err != nil {
		return fmt.Errorf("failed to create .rulestack directory: %w", err)
	}

	// Download package
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s.tgz", pkgRef.Name, pkgRef.Version))

	if err := c.DownloadBlob(sha256, tempFile); err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}
	defer os.Remove(tempFile) // Clean up temp file

	// Extract package
	packageDir := filepath.Join(rulestackDir, fmt.Sprintf("%s.%s", pkgRef.Name, pkgRef.Version))
	if err := pkg.Unpack(tempFile, packageDir); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// Update manifests
	if err := updateManifests(projectRoot, pkgRef, sha256); err != nil {
		return fmt.Errorf("failed to update manifests: %w", err)
	}

	// Update CLAUDE.md with new package rules
	if err := updateClaudeFile(projectRoot, pkgRef); err != nil {
		// Don't fail the entire operation if CLAUDE.md update fails
		if verbose {
			fmt.Printf("⚠️ Warning: Failed to update CLAUDE.md: %v\n", err)
		}
	}

	return nil
}

// reportInstallResults prints a comprehensive report of installation results
func reportInstallResults(results []InstallResult) {
	fmt.Printf("\n📦 Installation Summary:\n")

	installed := 0
	updated := 0
	skipped := 0
	failed := 0

	for _, result := range results {
		switch result.Status {
		case "installed":
			fmt.Printf("✅ %s@%s → installed successfully\n", result.Package, result.Version)
			installed++
		case "updated":
			fmt.Printf("✅ %s@%s → %s\n", result.Package, result.Version, result.Details)
			updated++
		case "skipped":
			fmt.Printf("⏭️ %s@%s → %s\n", result.Package, result.Version, result.Details)
			skipped++
		case "failed":
			fmt.Printf("❌ %s@%s → failed (%s)\n", result.Package, result.Version, result.Details)
			failed++
		}
	}

	fmt.Printf("\nSummary: %d installed, %d updated, %d skipped, %d failed\n", installed, updated, skipped, failed)

	if failed > 0 {
		fmt.Printf("⚠️  Some packages failed to install. Check network connectivity and registry access.\n")
	}
}
