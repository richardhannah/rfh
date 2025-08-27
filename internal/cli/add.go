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
	"rulestack/internal/manifest"
	"rulestack/internal/pkg"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <package@version>",
	Short: "Add (download) a ruleset package",
	Long: `Download and add a ruleset package to the current workspace.

Examples:
  rfh add mypackage@1.0.0`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd(args[0])
	},
}

// PackageRef represents a parsed package reference
type PackageRef struct {
	Name    string
	Version string
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
	packageDir := filepath.Join(rulestackDir, fmt.Sprintf("%s.%s", pkgRef.Name, pkgRef.Version))
	
	if _, err := os.Stat(packageDir); err == nil {
		// Package exists, prompt user
		if !confirmOverwrite(pkgRef.FullName()) {
			fmt.Printf("⏭️  Skipping %s\n", pkgRef.FullName())
			return nil
		}
	}

	// Get registry configuration (use default config only)
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Use current registry (no overrides)
	registryName := cfg.Current
	if registryName == "" {
		return fmt.Errorf("no registry configured. Use 'rfh registry add' to add a registry")
	}

	reg, exists := cfg.Registries[registryName]
	if !exists {
		return fmt.Errorf("registry '%s' not found. Use 'rfh registry list' to see available registries", registryName)
	}

	// Use default token (no overrides)
	authToken := getDefaultToken(reg)

	// Create client
	c := client.NewClient(reg.URL, authToken)
	c.SetVerbose(verbose)

	// Get package version info
	if verbose {
		fmt.Printf("🔍 Looking up package version...\n")
	}
	
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

	// Update CLAUDE.md with new package rules
	if err := updateClaudeFile(projectRoot, pkgRef); err != nil {
		// Don't fail the entire operation if CLAUDE.md update fails
		if verbose {
			fmt.Printf("⚠️ Warning: Failed to update CLAUDE.md: %v\n", err)
		}
	} else if verbose {
		fmt.Printf("📝 Updated CLAUDE.md with new package rules\n")
	}

	fmt.Printf("✅ Successfully added %s@%s\n", pkgRef.FullName(), pkgRef.Version)
	return nil
}

// parsePackageRef parses a package reference like "name@version"
func parsePackageRef(spec string) (*PackageRef, error) {
	if spec == "" {
		return nil, fmt.Errorf("package specification cannot be empty")
	}

	// Reject scoped package format (we don't support scopes anymore)
	if strings.HasPrefix(spec, "@") {
		return nil, fmt.Errorf("scoped packages are not supported: use simple name@version format (not @scope/name@version)")
	}

	// Check if version is specified
	if !strings.Contains(spec, "@") {
		return nil, fmt.Errorf("version must be specified: use package@version format")
	}

	// Parse name@version
	parts := strings.Split(spec, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid package format: use name@version")
	}
	
	name := parts[0]
	version := parts[1]

	if name == "" {
		return nil, fmt.Errorf("package name cannot be empty")
	}
	
	if version == "" {
		return nil, fmt.Errorf("package version cannot be empty")
	}

	return &PackageRef{
		Name:    name,
		Version: version,
	}, nil
}

// FullName returns the package name
func (p *PackageRef) FullName() string {
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
	projectManifest, err := loadOrCreateProjectManifest(manifestPath, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project manifest: %w", err)
	}

	projectManifest.Dependencies[pkgRef.FullName()] = pkgRef.Version

	if err := manifest.SaveProjectManifest(manifestPath, projectManifest); err != nil {
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
func loadOrCreateProjectManifest(path, projectRoot string) (*manifest.ProjectManifest, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create new manifest using centralized function
		return manifest.CreateProjectManifest(projectRoot), nil
	}

	// Load existing manifest using centralized function
	projectManifest, err := manifest.LoadProjectManifest(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load project manifest: %w", err)
	}

	return projectManifest, nil
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

// updateClaudeFile adds the newly installed package to CLAUDE.md
func updateClaudeFile(projectRoot string, pkgRef *PackageRef) error {
	claudePath := filepath.Join(projectRoot, "CLAUDE.md")
	templatePath := filepath.Join(projectRoot, "CLAUDE.TEMPLATE.md")
	
	// If CLAUDE.md doesn't exist, copy from template
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		if _, err := os.Stat(templatePath); err == nil {
			// Copy template to CLAUDE.md
			templateData, err := os.ReadFile(templatePath)
			if err != nil {
				return fmt.Errorf("failed to read CLAUDE template: %w", err)
			}
			if err := os.WriteFile(claudePath, templateData, 0644); err != nil {
				return fmt.Errorf("failed to create CLAUDE.md from template: %w", err)
			}
		} else {
			// Create basic CLAUDE.md if no template exists
			basicContent := `# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Active Rules (Rulestack core)
- @.rulestack/core.v1.0.0/core_rules.md
`
			if err := os.WriteFile(claudePath, []byte(basicContent), 0644); err != nil {
				return fmt.Errorf("failed to create basic CLAUDE.md: %w", err)
			}
		}
	}
	
	// Read current CLAUDE.md content
	content, err := os.ReadFile(claudePath)
	if err != nil {
		return fmt.Errorf("failed to read CLAUDE.md: %w", err)
	}
	
	lines := strings.Split(string(content), "\n")
	
	// Find actual rule files in the package directory
	packageDir := filepath.Join(projectRoot, ".rulestack", fmt.Sprintf("%s.%s", pkgRef.Name, pkgRef.Version))
	ruleFiles, err := findRuleFiles(packageDir)
	if err != nil {
		return fmt.Errorf("failed to find rule files in package: %w", err)
	}
	
	if len(ruleFiles) == 0 {
		// No rule files found, skip CLAUDE.md update
		return nil
	}
	
	// Generate rule lines for all found rule files
	var newRuleLines []string
	for _, ruleFile := range ruleFiles {
		// Make path relative to .rulestack directory
		relPath := filepath.Join(fmt.Sprintf("%s.%s", pkgRef.Name, pkgRef.Version), ruleFile)
		newRuleLines = append(newRuleLines, fmt.Sprintf("- @.rulestack/%s", strings.ReplaceAll(relPath, "\\", "/")))
	}
	
	// Check if any of these rules are already present
	existingRules := make(map[string]bool)
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "- @.rulestack/") {
			existingRules[strings.TrimSpace(line)] = true
		}
	}
	
	// Filter out already existing rules
	var rulesToAdd []string
	for _, ruleeLine := range newRuleLines {
		if !existingRules[ruleeLine] {
			rulesToAdd = append(rulesToAdd, ruleeLine)
		}
	}
	
	if len(rulesToAdd) == 0 {
		// All rules already exist, no need to add anything
		return nil
	}
	
	// Find where to insert the new rule
	var updatedLines []string
	inserted := false
	
	for i, line := range lines {
		// Look for the Active Rules section header
		if strings.Contains(line, "### Active Rules (Rulestack core)") || 
		   strings.Contains(line, "## Active Rules (Rulestack core)") {
			updatedLines = append(updatedLines, line)
			
			// Add all existing rules after the header
			j := i + 1
			for j < len(lines) {
				if strings.HasPrefix(strings.TrimSpace(lines[j]), "- @.rulestack/") {
					updatedLines = append(updatedLines, lines[j])
					j++
				} else if strings.TrimSpace(lines[j]) == "" {
					// Empty line, might be more rules after it
					updatedLines = append(updatedLines, lines[j])
					j++
				} else {
					// Found end of rules section, insert new rules here
					for _, ruleToAdd := range rulesToAdd {
						updatedLines = append(updatedLines, ruleToAdd)
					}
					inserted = true
					break
				}
			}
			if !inserted {
				// End of file reached, add the rules
				for _, ruleToAdd := range rulesToAdd {
					updatedLines = append(updatedLines, ruleToAdd)
				}
				inserted = true
			}
			
			// Add remaining lines after the rules section
			for j < len(lines) {
				updatedLines = append(updatedLines, lines[j])
				j++
			}
			break
		} else {
			updatedLines = append(updatedLines, line)
		}
	}
	
	// If we couldn't find the section, append to the end
	if !inserted {
		updatedLines = append(updatedLines, "", "## Active Rules (Rulestack core)")
		for _, ruleToAdd := range rulesToAdd {
			updatedLines = append(updatedLines, ruleToAdd)
		}
	}
	
	// Write updated content back to file
	updatedContent := strings.Join(updatedLines, "\n")
	if err := os.WriteFile(claudePath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to update CLAUDE.md: %w", err)
	}
	
	return nil
}

// findRuleFiles finds all .md files in the package directory that are likely rule files
func findRuleFiles(packageDir string) ([]string, error) {
	var ruleFiles []string
	
	err := filepath.Walk(packageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Only consider .md and .mdc files
		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(info.Name()), ".md") || strings.HasSuffix(strings.ToLower(info.Name()), ".mdc")) {
			// Get relative path from package directory
			relPath, err := filepath.Rel(packageDir, path)
			if err != nil {
				return err
			}
			ruleFiles = append(ruleFiles, relPath)
		}
		
		return nil
	})
	
	return ruleFiles, err
}