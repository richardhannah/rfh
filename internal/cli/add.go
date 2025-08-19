package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"rulestack/internal/client"
	"rulestack/internal/config"
	"rulestack/internal/pkg"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <package@version>",
	Short: "Add (download) a ruleset package",
	Long: `Download and add a ruleset package to the current workspace.

Examples:
  rfh add mypackage@1.0.0
  rfh add @scope/mypackage@2.1.0`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd(args[0])
	},
}

// PackageRef represents a parsed package reference
type PackageRef struct {
	Scope   string
	Name    string
	Version string
}

// ProjectManifest represents the rulestack.json file
type ProjectManifest struct {
	Version      string            `json:"version"`
	ProjectRoot  string            `json:"projectRoot"`
	Dependencies map[string]string `json:"dependencies"`
}

// LockManifest represents the rulestack.lock.json file
type LockManifest struct {
	Version      string                       `json:"version"`
	ProjectRoot  string                       `json:"projectRoot"`
	Packages     map[string]LockPackageEntry  `json:"packages"`
}

type LockPackageEntry struct {
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
}

// runAdd implements the add command logic
func runAdd(packageSpec string) error {
	// Parse package specification
	pkgRef, err := parsePackageRef(packageSpec)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("📦 Adding package: %s@%s\n", pkgRef.FullName(), pkgRef.Version)
	}

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	if verbose {
		fmt.Printf("📁 Project root: %s\n", projectRoot)
	}

	// Check if package already exists
	rulestackDir := filepath.Join(projectRoot, ".rulestack")
	packageDir := filepath.Join(rulestackDir, pkgRef.Name)
	
	if _, err := os.Stat(packageDir); err == nil {
		// Package exists, prompt user
		if !confirmOverwrite(pkgRef.FullName()) {
			fmt.Printf("⏭️  Skipping %s\n", pkgRef.FullName())
			return nil
		}
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

	// Create client
	c := client.NewClient(reg.URL, reg.Token)
	c.SetVerbose(verbose)

	// Get package version info
	if verbose {
		fmt.Printf("🔍 Looking up package version...\n")
	}
	
	versionInfo, err := c.GetPackageVersion(pkgRef.Scope, pkgRef.Name, pkgRef.Version)
	if err != nil {
		return fmt.Errorf("failed to get package version: %w", err)
	}

	// Extract SHA256 from version info
	sha256, ok := versionInfo["sha256"].(string)
	if !ok {
		return fmt.Errorf("package version missing sha256 hash")
	}

	// Create .rulestack directory if it doesn't exist
	if err := os.MkdirAll(rulestackDir, 0755); err != nil {
		return fmt.Errorf("failed to create .rulestack directory: %w", err)
	}

	// Download package
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s.tgz", pkgRef.Name, pkgRef.Version))
	
	if verbose {
		fmt.Printf("📥 Downloading package...\n")
	}
	
	if err := c.DownloadBlob(sha256, tempFile); err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}
	defer os.Remove(tempFile) // Clean up temp file

	// Extract package
	if verbose {
		fmt.Printf("📂 Extracting package...\n")
	}
	
	if err := pkg.Unpack(tempFile, packageDir); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// Update manifests
	if err := updateManifests(projectRoot, pkgRef, sha256); err != nil {
		return fmt.Errorf("failed to update manifests: %w", err)
	}

	fmt.Printf("✅ Successfully added %s@%s\n", pkgRef.FullName(), pkgRef.Version)
	return nil
}

// parsePackageRef parses a package reference like "name@version" or "@scope/name@version"
func parsePackageRef(spec string) (*PackageRef, error) {
	if spec == "" {
		return nil, fmt.Errorf("package specification cannot be empty")
	}

	// Check if version is specified
	if !strings.Contains(spec, "@") || strings.Count(spec, "@") == 0 {
		return nil, fmt.Errorf("version must be specified: use package@version format")
	}

	var scope, name, version string

	if strings.HasPrefix(spec, "@") {
		// Scoped package: @scope/name@version
		parts := strings.Split(spec[1:], "@") // Remove leading @ and split
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid scoped package format: use @scope/name@version")
		}
		
		scopeAndName := parts[0]
		version = parts[1]
		
		if !strings.Contains(scopeAndName, "/") {
			return nil, fmt.Errorf("invalid scoped package format: use @scope/name@version")
		}
		
		scopeParts := strings.SplitN(scopeAndName, "/", 2)
		scope = scopeParts[0]
		name = scopeParts[1]
	} else {
		// Regular package: name@version
		parts := strings.Split(spec, "@")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid package format: use name@version")
		}
		name = parts[0]
		version = parts[1]
	}

	if name == "" {
		return nil, fmt.Errorf("package name cannot be empty")
	}
	
	if version == "" {
		return nil, fmt.Errorf("package version cannot be empty")
	}

	return &PackageRef{
		Scope:   scope,
		Name:    name,
		Version: version,
	}, nil
}

// FullName returns the full package name including scope if present
func (p *PackageRef) FullName() string {
	if p.Scope != "" {
		return fmt.Sprintf("@%s/%s", p.Scope, p.Name)
	}
	return p.Name
}

// findProjectRoot finds the project root by looking for rulestack.json
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up the directory tree looking for rulestack.json
	for {
		manifestPath := filepath.Join(dir, "rulestack.json")
		if _, err := os.Stat(manifestPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, no rulestack.json found
			return "", fmt.Errorf("no RuleStack project found. Run 'rfh init' first to initialize a project")
		}
		dir = parent
	}
}

// confirmOverwrite prompts the user to confirm overwriting an existing package
func confirmOverwrite(packageName string) bool {
	fmt.Printf("⚠️  Package %s already exists. Do you want to reinstall it? (y/N): ", packageName)
	
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return response == "y" || response == "yes"
	}
	
	return false
}

// updateManifests updates both rulestack.json and rulestack.lock.json
func updateManifests(projectRoot string, pkgRef *PackageRef, sha256 string) error {
	// Update rulestack.json
	manifestPath := filepath.Join(projectRoot, "rulestack.json")
	manifest, err := loadOrCreateProjectManifest(manifestPath, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project manifest: %w", err)
	}

	manifest.Dependencies[pkgRef.FullName()] = pkgRef.Version

	if err := saveProjectManifest(manifestPath, manifest); err != nil {
		return fmt.Errorf("failed to save project manifest: %w", err)
	}

	// Update rulestack.lock.json
	lockPath := filepath.Join(projectRoot, "rulestack.lock.json")
	lockManifest, err := loadOrCreateLockManifest(lockPath, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load lock manifest: %w", err)
	}

	lockManifest.Packages[pkgRef.FullName()] = LockPackageEntry{
		Version: pkgRef.Version,
		SHA256:  sha256,
	}

	if err := saveLockManifest(lockPath, lockManifest); err != nil {
		return fmt.Errorf("failed to save lock manifest: %w", err)
	}

	return nil
}

// loadOrCreateProjectManifest loads or creates a new project manifest
func loadOrCreateProjectManifest(path, projectRoot string) (*ProjectManifest, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create new manifest
		return &ProjectManifest{
			Version:      "1.0.0",
			ProjectRoot:  projectRoot,
			Dependencies: make(map[string]string),
		}, nil
	}

	// Load existing manifest
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest ProjectManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}

	// Ensure dependencies map exists
	if manifest.Dependencies == nil {
		manifest.Dependencies = make(map[string]string)
	}

	return &manifest, nil
}

// saveProjectManifest saves the project manifest
func saveProjectManifest(path string, manifest *ProjectManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// loadOrCreateLockManifest loads or creates a new lock manifest
func loadOrCreateLockManifest(path, projectRoot string) (*LockManifest, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create new lock manifest
		return &LockManifest{
			Version:     "1.0.0",
			ProjectRoot: projectRoot,
			Packages:    make(map[string]LockPackageEntry),
		}, nil
	}

	// Load existing lock manifest
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var lockManifest LockManifest
	if err := json.Unmarshal(data, &lockManifest); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}

	// Ensure packages map exists
	if lockManifest.Packages == nil {
		lockManifest.Packages = make(map[string]LockPackageEntry)
	}

	return &lockManifest, nil
}

// saveLockManifest saves the lock manifest
func saveLockManifest(path string, lockManifest *LockManifest) error {
	data, err := json.MarshalIndent(lockManifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}