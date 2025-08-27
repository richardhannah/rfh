package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
)

// ProjectManifest represents the rulestack.json file in project mode (dependency management)
type ProjectManifest struct {
	Version      string            `json:"version"`
	ProjectRoot  string            `json:"projectRoot"`
	Dependencies map[string]string `json:"dependencies"`
}

// PackageManifest represents a single ruleset package entry
type PackageManifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Targets     []string `json:"targets,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Files       []string `json:"files"`
	License     string   `json:"license,omitempty"`
}

// PackageManifestFile represents the entire rulestack.json file in package mode (array of packages)
type PackageManifestFile []PackageManifest

// Legacy Manifest type for backward compatibility
type Manifest = PackageManifest
type ManifestFile = PackageManifestFile

var (
	ErrInvalidManifest    = errors.New("invalid manifest")
	ErrInvalidName        = errors.New("invalid package name")
	ErrInvalidVersion     = errors.New("invalid version")
	ErrEmptyManifest      = errors.New("manifest file cannot be empty")
	ErrInvalidProjectRoot = errors.New("invalid project root")
)

// nameRegex matches valid package names (with or without scope)
var nameRegex = regexp.MustCompile(`^(@[a-z0-9][a-z0-9\-_]*\/)?[a-z0-9][a-z0-9\-_]*$`)

// versionRegex matches semantic versions
var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9\-]+)?(\+[a-zA-Z0-9\-]+)?$`)

// PROJECT MANIFEST FUNCTIONS

// LoadProjectManifest reads and validates a project manifest from file
func LoadProjectManifest(path string) (*ProjectManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read project manifest: %w", err)
	}

	var manifest ProjectManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse project manifest JSON: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("invalid project manifest: %w", err)
	}

	return &manifest, nil
}

// SaveProjectManifest writes a project manifest to file
func SaveProjectManifest(path string, manifest *ProjectManifest) error {
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("invalid project manifest: %w", err)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project manifest: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// Validate checks if the project manifest is valid
func (pm *ProjectManifest) Validate() error {
	if pm.Version == "" {
		return fmt.Errorf("%w: version is required", ErrInvalidManifest)
	}

	if !versionRegex.MatchString(pm.Version) {
		return fmt.Errorf("%w: version must be semantic version (x.y.z)", ErrInvalidVersion)
	}

	if pm.ProjectRoot == "" {
		return fmt.Errorf("%w: projectRoot is required", ErrInvalidProjectRoot)
	}

	if pm.Dependencies == nil {
		return fmt.Errorf("%w: dependencies field is required (can be empty object)", ErrInvalidManifest)
	}

	return nil
}

// CreateProjectManifest creates a new project manifest with default values
func CreateProjectManifest(projectRoot string) *ProjectManifest {
	return &ProjectManifest{
		Version:      "1.0.0",
		ProjectRoot:  projectRoot,
		Dependencies: make(map[string]string),
	}
}

// PACKAGE MANIFEST FUNCTIONS

// LoadPackageManifests reads and validates all package manifests from file
func LoadPackageManifests(path string) (PackageManifestFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read package manifests: %w", err)
	}

	var manifests PackageManifestFile
	if err := json.Unmarshal(data, &manifests); err != nil {
		// Try to parse as single manifest for backward compatibility
		var singleManifest PackageManifest
		if err := json.Unmarshal(data, &singleManifest); err != nil {
			return nil, fmt.Errorf("failed to parse package manifest JSON: %w", err)
		}
		manifests = PackageManifestFile{singleManifest}
	}

	if len(manifests) == 0 {
		return nil, ErrEmptyManifest
	}

	// Validate all manifests
	for i, manifest := range manifests {
		if err := manifest.Validate(); err != nil {
			return nil, fmt.Errorf("package manifest[%d]: %w", i, err)
		}
	}

	return manifests, nil
}

// LoadFirstPackageManifest reads the first package manifest from file
func LoadFirstPackageManifest(path string) (*PackageManifest, error) {
	manifests, err := LoadPackageManifests(path)
	if err != nil {
		return nil, err
	}
	
	if len(manifests) == 0 {
		return nil, ErrEmptyManifest
	}
	
	return &manifests[0], nil
}

// SavePackageManifests writes all package manifests to file
func SavePackageManifests(path string, manifests PackageManifestFile) error {
	if len(manifests) == 0 {
		return ErrEmptyManifest
	}
	
	// Validate all manifests
	for i, manifest := range manifests {
		if err := manifest.Validate(); err != nil {
			return fmt.Errorf("package manifest[%d]: %w", i, err)
		}
	}

	data, err := json.MarshalIndent(manifests, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package manifests: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// SavePackageManifest writes a single package manifest to file as an array
func SavePackageManifest(path string, manifest *PackageManifest) error {
	manifests := PackageManifestFile{*manifest}
	return SavePackageManifests(path, manifests)
}

// SaveSinglePackageManifest writes a single package manifest as an object (not array)
func SaveSinglePackageManifest(path string, manifest *PackageManifest) error {
	// Validate the manifest
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("package manifest validation failed: %w", err)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package manifest: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// Validate checks if the package manifest is valid
func (pm *PackageManifest) Validate() error {
	if pm.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidManifest)
	}

	if !nameRegex.MatchString(pm.Name) {
		return fmt.Errorf("%w: name must match pattern %s", ErrInvalidName, nameRegex.String())
	}

	if pm.Version == "" {
		return fmt.Errorf("%w: version is required", ErrInvalidManifest)
	}

	if !versionRegex.MatchString(pm.Version) {
		return fmt.Errorf("%w: version must be semantic version (x.y.z)", ErrInvalidVersion)
	}

	if len(pm.Files) == 0 {
		return fmt.Errorf("%w: files array cannot be empty", ErrInvalidManifest)
	}

	// Validate targets
	validTargets := map[string]bool{
		"cursor":      true,
		"claude-code": true,
		"windsurf":    true,
		"copilot":     true,
	}

	for _, target := range pm.Targets {
		if !validTargets[target] {
			return fmt.Errorf("%w: invalid target '%s'", ErrInvalidManifest, target)
		}
	}

	return nil
}

// GetPackageName returns the package name (no scope support)
func (pm *PackageManifest) GetPackageName() string {
	return pm.Name
}

// CreateSamplePackageManifest creates a sample package manifest for initialization
func CreateSamplePackageManifest() *PackageManifest {
	return &PackageManifest{
		Name:        "example-rules",
		Version:     "0.1.0",
		Description: "Example AI ruleset",
		Targets:     []string{"cursor"},
		Tags:        []string{"example", "starter"},
		Files:       []string{"*.md"},
		License:     "MIT",
	}
}

// CreateSamplePackageManifestFile creates a sample package manifest file with one entry
func CreateSamplePackageManifestFile() PackageManifestFile {
	return PackageManifestFile{*CreateSamplePackageManifest()}
}

// LEGACY FUNCTIONS FOR BACKWARD COMPATIBILITY

// Load reads and validates a manifest from file (legacy function)
func Load(path string) (*Manifest, error) {
	return LoadFirstPackageManifest(path)
}

// LoadAll reads and validates all manifests from file (legacy function)
func LoadAll(path string) (ManifestFile, error) {
	return LoadPackageManifests(path)
}

// Save writes manifest to file as an array with single entry (legacy method)
func (m *Manifest) Save(path string) error {
	return SavePackageManifest(path, m)
}

// Save writes all manifests to file (legacy method)
func (mf ManifestFile) Save(path string) error {
	return SavePackageManifests(path, mf)
}

// CreateSample creates a sample manifest for initialization (legacy function)
func CreateSample() *Manifest {
	return CreateSamplePackageManifest()
}

// CreateSampleFile creates a sample manifest file with one entry (legacy function)
func CreateSampleFile() ManifestFile {
	return CreateSamplePackageManifestFile()
}

// UTILITY FUNCTIONS

// IsProjectManifest checks if a rulestack.json file contains a project manifest
func IsProjectManifest(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	// Try to parse as project manifest
	var pm ProjectManifest
	if err := json.Unmarshal(data, &pm); err != nil {
		return false
	}

	// Check if it has project manifest fields and lacks package manifest fields
	return pm.ProjectRoot != "" && pm.Dependencies != nil
}

// IsPackageManifest checks if a rulestack.json file contains package manifests
func IsPackageManifest(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	// Try to parse as package manifest array
	var pmf PackageManifestFile
	if err := json.Unmarshal(data, &pmf); err != nil {
		// Try single package manifest
		var pm PackageManifest
		if err := json.Unmarshal(data, &pm); err != nil {
			return false
		}
		return pm.Name != "" // Package manifests must have names
	}

	return len(pmf) > 0 && pmf[0].Name != ""
}