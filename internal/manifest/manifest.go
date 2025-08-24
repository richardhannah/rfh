package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Manifest represents a rulestack.json file
type Manifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Targets     []string `json:"targets,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Files       []string `json:"files"`
	License     string   `json:"license,omitempty"`
}

var (
	ErrInvalidManifest = errors.New("invalid manifest")
	ErrInvalidName     = errors.New("invalid package name")
	ErrInvalidVersion  = errors.New("invalid version")
)

// nameRegex matches valid package names (with or without scope)
var nameRegex = regexp.MustCompile(`^(@[a-z0-9][a-z0-9\-_]*\/)?[a-z0-9][a-z0-9\-_]*$`)

// versionRegex matches semantic versions
var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9\-]+)?(\+[a-zA-Z0-9\-]+)?$`)

// Load reads and validates a manifest from file
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// Save writes manifest to file
func (m *Manifest) Save(path string) error {
	if err := m.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
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

// GetScope returns the scope part of the package name, or empty string if unscoped
func (m *Manifest) GetScope() string {
	if strings.HasPrefix(m.Name, "@") {
		parts := strings.SplitN(m.Name[1:], "/", 2)
		if len(parts) == 2 {
			return parts[0]
		}
	}
	return ""
}

// GetPackageName returns the package name without scope
func (m *Manifest) GetPackageName() string {
	if strings.HasPrefix(m.Name, "@") {
		parts := strings.SplitN(m.Name[1:], "/", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return m.Name
}

// CreateSample creates a sample manifest for initialization
func CreateSample() *Manifest {
	return &Manifest{
		Name:        "@acme/example-rules",
		Version:     "0.1.0",
		Description: "Example AI ruleset",
		Targets:     []string{"cursor"},
		Tags:        []string{"example", "starter"},
		Files:       []string{"*.md"},
		License:     "MIT",
	}
}