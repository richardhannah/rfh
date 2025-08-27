package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"rulestack/internal/manifest"
	"rulestack/internal/pkg"
)

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
	// Simplified: pack no longer reads existing manifests
	// Just create new package with provided name
	if packageName == "" {
		return fmt.Errorf("--package is required in non-interactive mode")
	}
	
	// Note: --add-to-existing is no longer supported since we don't persist package manifests
	if addToExisting {
		return fmt.Errorf("--add-to-existing is not supported: pack creates new packages only")
	}
	
	// Create new package
	return createNewPackageNonInteractive(fileName, packageName)
}

// createNewPackageNonInteractive creates a new package without prompts
func createNewPackageNonInteractive(fileName string, pkgName string) error {
	return createPackageFromMetadata(fileName, pkgName, "1.0.0")
}