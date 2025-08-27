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

	// Create new manifest entry
	newManifest := &manifest.Manifest{
		Name:        packageName,
		Version:     "1.0.0",
		Description: fmt.Sprintf("Package containing %s", fileName),
		Files:       []string{fileName},
		Targets:     []string{"cursor"}, // Default target
		Tags:        []string{},
		License:     "MIT", // Default license
	}

	// Load existing manifests or create empty array
	manifestFile := "rulestack.json"
	var manifests manifest.ManifestFile
	
	if existingManifests, err := manifest.LoadAll(manifestFile); err == nil {
		manifests = existingManifests
	}

	// Add new manifest to array
	manifests = append(manifests, *newManifest)

	// Save updated manifests
	if err := manifests.Save(manifestFile); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	// Create package directory
	packageDir := getPackageDirectory(packageName, "1.0.0")
	if err := ensureDirectoryExists(packageDir); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	// Copy file to package directory
	destFile := filepath.Join(packageDir, fileName)
	if err := copyFile(fileName, destFile); err != nil {
		return fmt.Errorf("failed to copy file to package directory: %w", err)
	}

	// Create archive in staging directory
	stagingDir := getStagingDirectory()
	if err := ensureDirectoryExists(stagingDir); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	archivePath := filepath.Join(stagingDir, fmt.Sprintf("%s-1.0.0.tgz", packageName))
	info, err := pkg.PackFromDirectory(packageDir, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	fmt.Printf("✅ Created new package: %s v1.0.0\n", packageName)
	fmt.Printf("📁 Package directory: %s\n", packageDir)
	fmt.Printf("📦 Archive: %s\n", info.Path)
	fmt.Printf("📏 Size: %d bytes\n", info.SizeBytes)
	fmt.Printf("🔒 SHA256: %s\n", info.SHA256)

	return nil
}

// addToExistingPackage adds a file to an existing package with version increment
func addToExistingPackage(fileName string, manifests manifest.ManifestFile, packageIndex int) error {
	selectedManifest := manifests[packageIndex]
	
	fmt.Printf("Adding %s to package: %s (v%s)\n", fileName, selectedManifest.Name, selectedManifest.Version)

	// Prompt for new version
	newVersion, err := promptNewVersion(selectedManifest.Version)
	if err != nil {
		return err
	}

	// Update manifest
	updatedManifest := selectedManifest
	updatedManifest.Version = newVersion
	
	// Add file to files list if not already present
	fileExists := false
	for _, existingFile := range updatedManifest.Files {
		if existingFile == fileName {
			fileExists = true
			break
		}
	}
	
	if !fileExists {
		updatedManifest.Files = append(updatedManifest.Files, fileName)
	}

	// Update the manifest in the array
	manifests[packageIndex] = updatedManifest

	// Save updated manifests
	if err := manifests.Save("rulestack.json"); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	// Handle directory renaming (from old version to new version)
	oldPackageDir := getPackageDirectory(selectedManifest.Name, selectedManifest.Version)
	newPackageDir := getPackageDirectory(selectedManifest.Name, newVersion)

	// Check if old directory exists and rename it
	if _, err := os.Stat(oldPackageDir); err == nil {
		if err := os.Rename(oldPackageDir, newPackageDir); err != nil {
			return fmt.Errorf("failed to rename package directory: %w", err)
		}
	} else {
		// Old directory doesn't exist, create new one
		if err := ensureDirectoryExists(newPackageDir); err != nil {
			return fmt.Errorf("failed to create package directory: %w", err)
		}
	}

	// Copy new file to package directory
	destFile := filepath.Join(newPackageDir, fileName)
	if err := copyFile(fileName, destFile); err != nil {
		return fmt.Errorf("failed to copy file to package directory: %w", err)
	}

	// Create archive in staging directory
	stagingDir := getStagingDirectory()
	if err := ensureDirectoryExists(stagingDir); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	// Remove old version archive if it exists
	oldArchivePath := filepath.Join(stagingDir, fmt.Sprintf("%s-%s.tgz", selectedManifest.Name, selectedManifest.Version))
	if _, err := os.Stat(oldArchivePath); err == nil {
		os.Remove(oldArchivePath)
		fmt.Printf("🗑️  Removed old archive: %s-%s.tgz\n", selectedManifest.Name, selectedManifest.Version)
	}

	archivePath := filepath.Join(stagingDir, fmt.Sprintf("%s-%s.tgz", selectedManifest.Name, newVersion))
	info, err := pkg.PackFromDirectory(newPackageDir, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	fmt.Printf("✅ Updated package: %s v%s -> v%s\n", selectedManifest.Name, selectedManifest.Version, newVersion)
	fmt.Printf("📁 Package directory: %s\n", newPackageDir)
	fmt.Printf("📦 Archive: %s\n", info.Path)
	fmt.Printf("📏 Size: %d bytes\n", info.SizeBytes)
	fmt.Printf("🔒 SHA256: %s\n", info.SHA256)

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	
	return os.WriteFile(dst, data, 0o644)
}

// runNonInteractivePack handles non-interactive pack mode with command-line flags
func runNonInteractivePack(fileName string) error {
	// Check if rulestack.json exists
	manifestFile := "rulestack.json"
	manifests, err := manifest.LoadAll(manifestFile)
	if err != nil {
		if os.IsNotExist(err) && !addToExisting {
			// No rulestack.json exists, create new package
			return createNewPackageNonInteractive(fileName, packageName)
		} else if os.IsNotExist(err) && addToExisting {
			return fmt.Errorf("cannot add to existing package: no rulestack.json found")
		}
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	if addToExisting {
		// Add to existing package
		if newVersion == "" {
			return fmt.Errorf("--version is required when using --add-to-existing")
		}
		
		// Find the package by name
		packageIndex := -1
		for i, m := range manifests {
			if m.Name == packageName {
				packageIndex = i
				break
			}
		}
		
		if packageIndex == -1 {
			return fmt.Errorf("package '%s' not found in manifest", packageName)
		}
		
		// Validate version increment
		currentVersion := manifests[packageIndex].Version
		if err := validateVersionIncrease(currentVersion, newVersion); err != nil {
			return fmt.Errorf("version validation failed: %w", err)
		}
		
		return addToExistingPackageNonInteractive(fileName, manifests, packageIndex, newVersion)
	} else {
		// Create new package
		return createNewPackageNonInteractive(fileName, packageName)
	}
}

// createNewPackageNonInteractive creates a new package without prompts
func createNewPackageNonInteractive(fileName string, pkgName string) error {
	// Create new manifest entry
	newManifest := &manifest.Manifest{
		Name:        pkgName,
		Version:     "1.0.0",
		Description: fmt.Sprintf("Package containing %s", fileName),
		Files:       []string{fileName},
		Targets:     []string{"cursor"}, // Default target
		Tags:        []string{},
		License:     "MIT", // Default license
	}

	// Load existing manifests or create empty array
	manifestFile := "rulestack.json"
	var manifests manifest.ManifestFile
	
	if existingManifests, err := manifest.LoadAll(manifestFile); err == nil {
		manifests = existingManifests
	}

	// Add new manifest to array
	manifests = append(manifests, *newManifest)

	// Save updated manifests
	if err := manifests.Save(manifestFile); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	// Create package directory
	packageDir := getPackageDirectory(pkgName, "1.0.0")
	if err := ensureDirectoryExists(packageDir); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	// Copy file to package directory
	destFile := filepath.Join(packageDir, fileName)
	if err := copyFile(fileName, destFile); err != nil {
		return fmt.Errorf("failed to copy file to package directory: %w", err)
	}

	// Create archive in staging directory
	stagingDir := getStagingDirectory()
	if err := ensureDirectoryExists(stagingDir); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	archivePath := filepath.Join(stagingDir, fmt.Sprintf("%s-1.0.0.tgz", pkgName))
	info, err := pkg.PackFromDirectory(packageDir, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	fmt.Printf("✅ Created new package: %s v1.0.0\n", pkgName)
	fmt.Printf("📁 Package directory: %s\n", packageDir)
	fmt.Printf("📦 Archive: %s\n", info.Path)
	fmt.Printf("📏 Size: %d bytes\n", info.SizeBytes)
	fmt.Printf("🔒 SHA256: %s\n", info.SHA256)

	return nil
}

// addToExistingPackageNonInteractive adds a file to an existing package without prompts
func addToExistingPackageNonInteractive(fileName string, manifests manifest.ManifestFile, packageIndex int, version string) error {
	selectedManifest := manifests[packageIndex]
	
	fmt.Printf("Adding %s to package: %s (v%s -> v%s)\n", fileName, selectedManifest.Name, selectedManifest.Version, version)

	// Update manifest
	updatedManifest := selectedManifest
	updatedManifest.Version = version
	
	// Add file to files list if not already present
	fileExists := false
	for _, existingFile := range updatedManifest.Files {
		if existingFile == fileName {
			fileExists = true
			break
		}
	}
	
	if !fileExists {
		updatedManifest.Files = append(updatedManifest.Files, fileName)
	}

	// Update the manifest in the array
	manifests[packageIndex] = updatedManifest

	// Save updated manifests
	if err := manifests.Save("rulestack.json"); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	// Handle directory renaming (from old version to new version)
	oldPackageDir := getPackageDirectory(selectedManifest.Name, selectedManifest.Version)
	newPackageDir := getPackageDirectory(selectedManifest.Name, version)

	// Check if old directory exists and rename it
	if _, err := os.Stat(oldPackageDir); err == nil {
		if err := os.Rename(oldPackageDir, newPackageDir); err != nil {
			return fmt.Errorf("failed to rename package directory: %w", err)
		}
	} else {
		// Old directory doesn't exist, create new one
		if err := ensureDirectoryExists(newPackageDir); err != nil {
			return fmt.Errorf("failed to create package directory: %w", err)
		}
	}

	// Copy new file to package directory
	destFile := filepath.Join(newPackageDir, fileName)
	if err := copyFile(fileName, destFile); err != nil {
		return fmt.Errorf("failed to copy file to package directory: %w", err)
	}

	// Create archive in staging directory
	stagingDir := getStagingDirectory()
	if err := ensureDirectoryExists(stagingDir); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	// Remove old version archive if it exists
	oldArchivePath := filepath.Join(stagingDir, fmt.Sprintf("%s-%s.tgz", selectedManifest.Name, selectedManifest.Version))
	if _, err := os.Stat(oldArchivePath); err == nil {
		os.Remove(oldArchivePath)
		fmt.Printf("🗑️  Removed old archive: %s-%s.tgz\n", selectedManifest.Name, selectedManifest.Version)
	}

	archivePath := filepath.Join(stagingDir, fmt.Sprintf("%s-%s.tgz", selectedManifest.Name, version))
	info, err := pkg.PackFromDirectory(newPackageDir, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	fmt.Printf("✅ Updated package: %s v%s -> v%s\n", selectedManifest.Name, selectedManifest.Version, version)
	fmt.Printf("📁 Package directory: %s\n", newPackageDir)
	fmt.Printf("📦 Archive: %s\n", info.Path)
	fmt.Printf("📏 Size: %d bytes\n", info.SizeBytes)
	fmt.Printf("🔒 SHA256: %s\n", info.SHA256)

	return nil
}

