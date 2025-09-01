package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rulestack/internal/manifest"
	"rulestack/internal/pkg"
	"rulestack/internal/version"
)

// ExistingPackageInfo holds information about an installed package
type ExistingPackageInfo struct {
	Name          string   // Package name (e.g., "security-rules")
	Version       string   // Current installed version (e.g., "1.0.0")
	Directory     string   // Full path to package directory (e.g., ".rulestack/security-rules.1.0.0")
	ExistingFiles []string // List of rule files in the package (e.g., ["rule1.mdc", "rule2.mdc"])
	ManifestPath  string   // Path to project's rulestack.json
}

// createNewPackage creates a new package with the given file
func createNewPackage(fileName string) error {
	// Prompt for package name
	packageName, err := promptUserInput("Enter new package name")
	if err != nil {
		return err
	}

	if packageName == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	return createPackageFromMetadata(fileName, packageName, "1.0.0")
}

// createPackageFromMetadata creates a package with specified metadata (no manifest files saved)
func createPackageFromMetadata(fileName, packageName, version string) error {
	// Create package manifest in memory only
	packageManifest := &manifest.PackageManifest{
		Name:        packageName,
		Version:     version,
		Description: fmt.Sprintf("Package containing %s", fileName),
		Files:       []string{fileName},
		Targets:     []string{"cursor"}, // Default target
		Tags:        []string{},
		License:     "MIT", // Default license
	}

	// Create package directory
	packageDir := getPackageDirectory(packageName, version)
	if err := ensureDirectoryExists(packageDir); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	// Copy file to package directory
	destFile := filepath.Join(packageDir, fileName)
	if err := copyFile(fileName, destFile); err != nil {
		return fmt.Errorf("failed to copy file to package directory: %w", err)
	}

	// Create archive in staging directory with embedded manifest
	stagingDir := getStagingDirectory()
	if err := ensureDirectoryExists(stagingDir); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	// First, write the manifest to the package directory so it gets included in the archive
	// Use SaveSinglePackageManifest to save as object (not array) for archive embedding
	manifestPath := filepath.Join(packageDir, "rulestack.json")
	if err := manifest.SaveSinglePackageManifest(manifestPath, packageManifest); err != nil {
		return fmt.Errorf("failed to write manifest to package directory: %w", err)
	}

	archivePath := filepath.Join(stagingDir, fmt.Sprintf("%s-%s.tgz", packageName, version))
	info, err := pkg.PackFromDirectory(packageDir, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	fmt.Printf("✅ Created new package: %s v%s\n", packageName, version)
	fmt.Printf("📁 Package directory: %s\n", packageDir)
	fmt.Printf("📦 Archive: %s\n", info.Path)
	fmt.Printf("📏 Size: %d bytes\n", info.SizeBytes)
	fmt.Printf("🔒 SHA256: %s\n", info.SHA256)

	return nil
}

// copyFile copies a file from source to destination
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	
	return os.WriteFile(dst, data, 0o644)
}

// runNonInteractivePack handles non-interactive pack mode with command-line flags
func runNonInteractivePack(fileName string) error {
	if packageName == "" {
		return fmt.Errorf("--package is required in non-interactive mode")
	}
	
	// Check if this package already exists as an installed dependency
	existingPkg, err := checkExistingPackage(packageName)
	if err != nil {
		return fmt.Errorf("failed to check for existing package: %w", err)
	}
	
	if existingPkg != nil {
		// Package exists - create updated version with all files
		fmt.Printf("📦 Found existing package %s@%s with %d files\n", 
			existingPkg.Name, existingPkg.Version, len(existingPkg.ExistingFiles))
		
		if packageVersion == "" {
			// Auto-increment patch version if no version specified
			nextVersion, err := version.IncrementPatchVersion(existingPkg.Version)
			if err != nil {
				return fmt.Errorf("failed to auto-increment version: %w", err)
			}
			packageVersion = nextVersion
			fmt.Printf("🔄 Auto-incrementing version to %s\n", packageVersion)
		}
		
		return createUpdatedPackage(fileName, packageName, packageVersion, existingPkg)
	} else {
		// Package doesn't exist - create new package (existing behavior)
		if packageVersion == "" {
			packageVersion = "1.0.0" // Default version for new packages
		}
		
		fmt.Printf("🆕 Creating new package %s@%s\n", packageName, packageVersion)
		return createNewPackageNonInteractive(fileName, packageName, packageVersion)
	}
}

// createNewPackageNonInteractive creates a new package without prompts
func createNewPackageNonInteractive(fileName string, pkgName string, version string) error {
	return createPackageFromMetadata(fileName, pkgName, version)
}

// checkExistingPackage looks for an installed package by name in the project
func checkExistingPackage(packageName string) (*ExistingPackageInfo, error) {
	// 1. Look for project manifest (rulestack.json in project root)
	manifestPath := "rulestack.json"
	
	// Check if project manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, nil // No project manifest = no installed packages
	}
	
	// 2. Load project manifest
	projectManifest, err := manifest.LoadProjectManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load project manifest: %w", err)
	}
	
	// 3. Check if package exists in dependencies
	if projectManifest.Dependencies == nil {
		return nil, nil // No dependencies
	}
	
	currentVersion, exists := projectManifest.Dependencies[packageName]
	if !exists {
		return nil, nil // Package not installed
	}
	
	// 4. Verify package directory exists
	packageDir := getPackageDirectory(packageName, currentVersion)
	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("package %s@%s is in manifest but directory %s does not exist", 
			packageName, currentVersion, packageDir)
	}
	
	// 5. Discover existing rule files in package directory
	existingFiles, err := findRuleFilesInDirectory(packageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan existing files in %s: %w", packageDir, err)
	}
	
	return &ExistingPackageInfo{
		Name:          packageName,
		Version:       currentVersion,
		Directory:     packageDir,
		ExistingFiles: existingFiles,
		ManifestPath:  manifestPath,
	}, nil
}

// findRuleFilesInDirectory scans a directory for .mdc rule files
func findRuleFilesInDirectory(dirPath string) ([]string, error) {
	var files []string
	
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}
		
		name := entry.Name()
		// Include .mdc files but exclude rulestack.json manifest
		if strings.HasSuffix(strings.ToLower(name), ".mdc") && name != "rulestack.json" {
			files = append(files, name)
		}
	}
	
	return files, nil
}

// createUpdatedPackage creates a new version of an existing package with additional files
func createUpdatedPackage(fileName, packageName, newVersion string, existingPkg *ExistingPackageInfo) error {
	// Pre-flight validations
	
	// 1. Ensure new file exists and is readable
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return fmt.Errorf("input file %s does not exist", fileName)
	}
	
	// 2. Check for file name conflicts
	for _, existingFile := range existingPkg.ExistingFiles {
		if existingFile == fileName {
			return fmt.Errorf("file %s already exists in package %s@%s, use a different filename or increment version to replace", 
				fileName, packageName, existingPkg.Version)
		}
	}
	
	// 3. Validate file is .mdc format
	if !strings.HasSuffix(strings.ToLower(fileName), ".mdc") {
		return fmt.Errorf("file %s must be a .mdc rule file", fileName)
	}
	
	// 4. Validate version increase using version package
	if err := version.ValidateVersionIncrease(existingPkg.Version, newVersion); err != nil {
		return fmt.Errorf("version validation failed: %w", err)
	}
	
	// 5. Create new package directory
	newPackageDir := getPackageDirectory(packageName, newVersion)
	if err := ensureDirectoryExists(newPackageDir); err != nil {
		return fmt.Errorf("failed to create new package directory %s: %w", newPackageDir, err)
	}
	
	// Set up cleanup on failure
	success := false
	defer func() {
		if !success {
			// Clean up partial directory on failure
			os.RemoveAll(newPackageDir)
		}
	}()
	
	// 6. Copy ALL existing files from old package to new package
	for _, existingFile := range existingPkg.ExistingFiles {
		srcPath := filepath.Join(existingPkg.Directory, existingFile)
		destPath := filepath.Join(newPackageDir, existingFile)
		
		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy existing file %s: %w", existingFile, err)
		}
	}
	
	// 7. Copy new rule file to package directory
	newFilePath := filepath.Join(newPackageDir, fileName)
	if err := copyFile(fileName, newFilePath); err != nil {
		return fmt.Errorf("failed to copy new file %s: %w", fileName, err)
	}
	
	// 8. Build complete file list for manifest
	allFiles := make([]string, 0, len(existingPkg.ExistingFiles)+1)
	allFiles = append(allFiles, existingPkg.ExistingFiles...)
	allFiles = append(allFiles, fileName)
	
	// 9. Create updated package manifest
	packageManifest := &manifest.PackageManifest{
		Name:        packageName,
		Version:     newVersion,
		Description: fmt.Sprintf("Updated package containing %d rule files", len(allFiles)),
		Files:       allFiles,
		Targets:     []string{"cursor"}, // Default target
		Tags:        []string{},
		License:     "MIT", // Default license
	}
	
	// 10. Save manifest to new package directory
	manifestPath := filepath.Join(newPackageDir, "rulestack.json")
	if err := manifest.SaveSinglePackageManifest(manifestPath, packageManifest); err != nil {
		return fmt.Errorf("failed to write manifest to new package directory: %w", err)
	}
	
	// 11. Create archive in staging directory
	stagingDir := getStagingDirectory()
	if err := ensureDirectoryExists(stagingDir); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}
	
	archivePath := filepath.Join(stagingDir, fmt.Sprintf("%s-%s.tgz", packageName, newVersion))
	info, err := pkg.PackFromDirectory(newPackageDir, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	
	// 12. Success output
	fmt.Printf("✅ Updated existing package: %s v%s -> v%s\n", packageName, existingPkg.Version, newVersion)
	fmt.Printf("📁 Package directory: %s\n", newPackageDir)
	fmt.Printf("📦 Archive: %s\n", info.Path)
	fmt.Printf("📏 Size: %d bytes\n", info.SizeBytes)
	fmt.Printf("🔒 SHA256: %s\n", info.SHA256)
	fmt.Printf("📋 Files included: %s\n", strings.Join(allFiles, ", "))
	
	// Mark success at the end
	success = true
	return nil
}