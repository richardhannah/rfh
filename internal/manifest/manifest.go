package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
)

// Manifest represents a single ruleset entry in rulestack.json
type Manifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Targets     []string `json:"targets,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Files       []string `json:"files"`
	License     string   `json:"license,omitempty"`
}

// ManifestFile represents the entire rulestack.json file (array of manifests)
type ManifestFile []Manifest

var (
	ErrInvalidManifest = errors.New("invalid manifest")
	ErrInvalidName     = errors.New("invalid package name")
	ErrInvalidVersion  = errors.New("invalid version")
	ErrEmptyManifest   = errors.New("manifest file cannot be empty")
)

// nameRegex matches valid package names (with or without scope)
var nameRegex = regexp.MustCompile(`^(@[a-z0-9][a-z0-9\-_]*\/)?[a-z0-9][a-z0-9\-_]*$`)

// versionRegex matches semantic versions
var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9\-]+)?(\+[a-zA-Z0-9\-]+)?$`)

// Load reads and validates a manifest from file
// For backward compatibility, it returns the first manifest in the array
func Load(path string) (*Manifest, error) {
	manifests, err := LoadAll(path)
	if err != nil {
		return nil, err
	}
	
	if len(manifests) == 0 {
		return nil, ErrEmptyManifest
	}
	
	return &manifests[0], nil
}

// LoadAll reads and validates all manifests from file
func LoadAll(path string) (ManifestFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifests ManifestFile
	if err := json.Unmarshal(data, &manifests); err != nil {
		// Try to parse as single manifest for backward compatibility
		var singleManifest Manifest
		if err := json.Unmarshal(data, &singleManifest); err != nil {
			return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
		}
		manifests = ManifestFile{singleManifest}
	}

	if len(manifests) == 0 {
		return nil, ErrEmptyManifest
	}

	// Validate all manifests
	for i, manifest := range manifests {
		if err := manifest.Validate(); err != nil {
			return nil, fmt.Errorf("manifest[%d]: %w", i, err)
		}
	}

	return manifests, nil
}

// Save writes manifest to file as an array with single entry
func (m *Manifest) Save(path string) error {
	manifests := ManifestFile{*m}
	return manifests.Save(path)
}

// Save writes all manifests to file
func (mf ManifestFile) Save(path string) error {
	if len(mf) == 0 {
		return ErrEmptyManifest
	}
	
	// Validate all manifests
	for i, manifest := range mf {
		if err := manifest.Validate(); err != nil {
			return fmt.Errorf("manifest[%d]: %w", i, err)
		}
	}

	data, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// Validate checks if the manifest is valid
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidManifest)
	}

	if !nameRegex.MatchString(m.Name) {
		return fmt.Errorf("%w: name must match pattern %s", ErrInvalidName, nameRegex.String())
	}

	if m.Version == "" {
		return fmt.Errorf("%w: version is required", ErrInvalidManifest)
	}

	if !versionRegex.MatchString(m.Version) {
		return fmt.Errorf("%w: version must be semantic version (x.y.z)", ErrInvalidVersion)
	}

	if len(m.Files) == 0 {
		return fmt.Errorf("%w: files array cannot be empty", ErrInvalidManifest)
	}

	// Validate targets
	validTargets := map[string]bool{
		"cursor":      true,
		"claude-code": true,
		"windsurf":    true,
		"copilot":     true,
	}

	for _, target := range m.Targets {
		if !validTargets[target] {
			return fmt.Errorf("%w: invalid target '%s'", ErrInvalidManifest, target)
		}
	}

	return nil
}

// GetPackageName returns the package name (no scope support)
func (m *Manifest) GetPackageName() string {
	return m.Name
}

// CreateSample creates a sample manifest for initialization
func CreateSample() *Manifest {
	return &Manifest{
		Name:        "example-rules",
		Version:     "0.1.0",
		Description: "Example AI ruleset",
		Targets:     []string{"cursor"},
		Tags:        []string{"example", "starter"},
		Files:       []string{"*.md"},
		License:     "MIT",
	}
}

// CreateSampleFile creates a sample manifest file with one entry
func CreateSampleFile() ManifestFile {
	return ManifestFile{*CreateSample()}
}