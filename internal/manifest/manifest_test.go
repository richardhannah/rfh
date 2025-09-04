package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestManifestValidation(t *testing.T) {
	tests := []struct {
		name      string
		manifest  Manifest
		expectErr bool
		errType   error
	}{
		{
			name: "valid scoped package",
			manifest: Manifest{
				Name:    "@acme/test-rules",
				Version: "1.0.0",
				Files:   []string{"rules/*.md"},
			},
			expectErr: false,
		},
		{
			name: "valid unscoped package",
			manifest: Manifest{
				Name:    "test-rules",
				Version: "1.0.0",
				Files:   []string{"rules/*.md"},
			},
			expectErr: false,
		},
		{
			name: "valid semver with prerelease",
			manifest: Manifest{
				Name:    "test-rules",
				Version: "1.0.0-alpha1",
				Files:   []string{"rules/*.md"},
			},
			expectErr: false,
		},
		{
			name: "valid semver with build metadata",
			manifest: Manifest{
				Name:    "test-rules",
				Version: "1.0.0+build1",
				Files:   []string{"rules/*.md"},
			},
			expectErr: false,
		},
		{
			name: "missing name",
			manifest: Manifest{
				Version: "1.0.0",
				Files:   []string{"rules/*.md"},
			},
			expectErr: true,
			errType:   ErrInvalidManifest,
		},
		{
			name: "invalid name format",
			manifest: Manifest{
				Name:    "Invalid Name",
				Version: "1.0.0",
				Files:   []string{"rules/*.md"},
			},
			expectErr: true,
			errType:   ErrInvalidName,
		},
		{
			name: "missing version",
			manifest: Manifest{
				Name:  "test-rules",
				Files: []string{"rules/*.md"},
			},
			expectErr: true,
			errType:   ErrInvalidManifest,
		},
		{
			name: "invalid version format",
			manifest: Manifest{
				Name:    "test-rules",
				Version: "1.0",
				Files:   []string{"rules/*.md"},
			},
			expectErr: true,
			errType:   ErrInvalidVersion,
		},
		{
			name: "empty files array",
			manifest: Manifest{
				Name:    "test-rules",
				Version: "1.0.0",
				Files:   []string{},
			},
			expectErr: true,
			errType:   ErrInvalidManifest,
		},
		{
			name: "invalid target",
			manifest: Manifest{
				Name:    "test-rules",
				Version: "1.0.0",
				Files:   []string{"rules/*.md"},
				Targets: []string{"invalid-target"},
			},
			expectErr: true,
			errType:   ErrInvalidManifest,
		},
		{
			name: "valid targets",
			manifest: Manifest{
				Name:    "test-rules",
				Version: "1.0.0",
				Files:   []string{"rules/*.md"},
				Targets: []string{"cursor", "claude-code", "windsurf", "copilot"},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errType != nil && !isErrorType(err, tt.errType) {
					t.Errorf("expected error type %v, got %v", tt.errType, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetPackageName(t *testing.T) {
	tests := []struct {
		name     string
		manifest Manifest
		expected string
	}{
		{
			name:     "scoped package name",
			manifest: Manifest{Name: "@acme/test-rules"},
			expected: "@acme/test-rules",
		},
		{
			name:     "unscoped package name",
			manifest: Manifest{Name: "test-rules"},
			expected: "test-rules",
		},
		{
			name:     "simple package name",
			manifest: Manifest{Name: "@acme"},
			expected: "@acme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.manifest.GetPackageName()
			if result != tt.expected {
				t.Errorf("GetPackageName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLoadManifest(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "manifest_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("loads valid manifest", func(t *testing.T) {
		validManifest := Manifest{
			Name:        "@test/rules",
			Version:     "1.0.0",
			Description: "Test rules",
			Files:       []string{"rules/*.md"},
			Targets:     []string{"cursor"},
		}

		manifestPath := filepath.Join(tempDir, "valid.json")
		data, _ := json.MarshalIndent(validManifest, "", "  ")
		os.WriteFile(manifestPath, data, 0644)

		manifest, err := Load(manifestPath)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if manifest.Name != validManifest.Name {
			t.Errorf("Name = %q, want %q", manifest.Name, validManifest.Name)
		}
	})

	t.Run("fails on invalid JSON", func(t *testing.T) {
		invalidPath := filepath.Join(tempDir, "invalid.json")
		os.WriteFile(invalidPath, []byte("invalid json"), 0644)

		_, err := Load(invalidPath)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("fails on missing file", func(t *testing.T) {
		_, err := Load("nonexistent.json")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})
}

func TestSaveManifest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "manifest_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("saves valid manifest", func(t *testing.T) {
		manifest := Manifest{
			Name:    "@test/rules",
			Version: "1.0.0",
			Files:   []string{"rules/*.md"},
		}

		manifestPath := filepath.Join(tempDir, "save_test.json")
		err := SaveSinglePackageManifest(manifestPath, &manifest)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Verify file was created and is valid JSON
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Errorf("failed to read saved file: %v", err)
		}

		var loaded Manifest
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Errorf("saved file is not valid JSON: %v", err)
		}

		if loaded.Name != manifest.Name {
			t.Errorf("loaded name = %q, want %q", loaded.Name, manifest.Name)
		}
	})

	t.Run("fails to save invalid manifest", func(t *testing.T) {
		manifest := Manifest{
			Name:    "", // invalid - empty name
			Version: "1.0.0",
			Files:   []string{"rules/*.md"},
		}

		manifestPath := filepath.Join(tempDir, "invalid_save.json")
		err := manifest.Save(manifestPath)
		if err == nil {
			t.Error("expected error for invalid manifest")
		}
	})
}

func TestCreateSample(t *testing.T) {
	sample := CreateSample()

	if err := sample.Validate(); err != nil {
		t.Errorf("sample manifest is invalid: %v", err)
	}

	if sample.Name != "example-rules" {
		t.Errorf("sample name = %q, want %q", sample.Name, "example-rules")
	}

	if sample.Version != "0.1.0" {
		t.Errorf("sample version = %q, want %q", sample.Version, "0.1.0")
	}
}

// Helper function to check if error is of specific type
func isErrorType(err, target error) bool {
	return err.Error() == target.Error() ||
		(target == ErrInvalidManifest && err.Error() != "" && err.Error() != ErrInvalidName.Error() && err.Error() != ErrInvalidVersion.Error()) ||
		(target == ErrInvalidName && err.Error() != "" && err.Error() != ErrInvalidManifest.Error() && err.Error() != ErrInvalidVersion.Error()) ||
		(target == ErrInvalidVersion && err.Error() != "" && err.Error() != ErrInvalidManifest.Error() && err.Error() != ErrInvalidName.Error())
}
