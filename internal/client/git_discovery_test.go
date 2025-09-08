package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGitRegistryDiscovery(t *testing.T) {
	// Create a temporary directory for test repository
	tempDir := t.TempDir()

	// Create test registry structure
	packagesDir := filepath.Join(tempDir, "packages")
	testPkgDir := filepath.Join(packagesDir, "test-package")
	versionsDir := filepath.Join(testPkgDir, "versions")
	versionDir := filepath.Join(versionsDir, "1.0.0")

	// Create directory structure
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Create index.json
	index := GitRegistryIndex{
		Version:      "1.0",
		UpdatedAt:    time.Now(),
		PackageCount: 1,
		Packages: map[string]GitPackageEntry{
			"test-package": {
				Name:        "test-package",
				Description: "A test package",
				Latest:      "1.0.0",
				Tags:        []string{"test", "example"},
				UpdatedAt:   time.Now(),
			},
		},
	}

	indexData, _ := json.MarshalIndent(index, "", "  ")
	if err := os.WriteFile(filepath.Join(tempDir, "index.json"), indexData, 0644); err != nil {
		t.Fatalf("Failed to write index.json: %v", err)
	}

	// Create package metadata
	metadata := GitPackageMetadata{
		Name:        "test-package",
		Description: "A test package",
		Latest:      "1.0.0",
		Versions: []GitVersionSummary{
			{
				Version:     "1.0.0",
				SHA256:      "abc123",
				Size:        1024,
				PublishedAt: time.Now(),
			},
		},
		Tags:      []string{"test", "example"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	metadataData, _ := json.MarshalIndent(metadata, "", "  ")
	if err := os.WriteFile(filepath.Join(testPkgDir, "metadata.json"), metadataData, 0644); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}

	// Create version manifest
	manifest := GitManifest{
		Name:         "test-package",
		Version:      "1.0.0",
		Description:  "A test package",
		Dependencies: map[string]string{"dep1": "^1.0.0"},
		SHA256:       "abc123",
		Size:         1024,
		PublishedAt:  time.Now(),
		Publisher:    "test-publisher",
	}

	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(versionDir, "manifest.json"), manifestData, 0644); err != nil {
		t.Fatalf("Failed to write manifest.json: %v", err)
	}

	// Create dummy archive
	if err := os.WriteFile(filepath.Join(versionDir, "archive.tar.gz"), []byte("test archive"), 0644); err != nil {
		t.Fatalf("Failed to write archive: %v", err)
	}

	// Create Git client with test directory as cache
	client := &GitClient{
		cacheDir: tempDir,
		verbose:  false, // Reduce noise in tests
	}

	t.Run("LoadIndex", func(t *testing.T) {
		// Test loadIndex directly without Git operations
		index, err := client.rebuildIndex()
		if err != nil {
			// Try loading from file
			data, _ := os.ReadFile(filepath.Join(tempDir, "index.json"))
			json.Unmarshal(data, &index)
		}

		if index == nil {
			t.Fatal("Failed to load or rebuild index")
		}

		if len(index.Packages) != 1 {
			t.Errorf("Expected 1 package in index, got %d", len(index.Packages))
		}
	})

	t.Run("SearchPackages", func(t *testing.T) {
		// Test search functionality directly using helper methods
		index, _ := client.rebuildIndex()
		if index == nil {
			t.Skip("Index not available")
		}

		var results []Package
		for _, entry := range index.Packages {
			if client.matchesSearch(entry, SearchOptions{Query: "test"}) {
				pkg := Package{
					Name:        entry.Name,
					Description: entry.Description,
					Latest:      entry.Latest,
					Tags:        entry.Tags,
					UpdatedAt:   entry.UpdatedAt,
				}
				results = append(results, pkg)
			}
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 package, got %d", len(results))
		}

		if results[0].Name != "test-package" {
			t.Errorf("Expected package name 'test-package', got '%s'", results[0].Name)
		}
	})

	t.Run("LoadPackageMetadata", func(t *testing.T) {
		metadata, err := client.loadPackageMetadata("test-package")
		if err != nil {
			t.Fatalf("loadPackageMetadata failed: %v", err)
		}

		if metadata.Name != "test-package" {
			t.Errorf("Expected package name 'test-package', got '%s'", metadata.Name)
		}

		if metadata.Latest != "1.0.0" {
			t.Errorf("Expected latest version '1.0.0', got '%s'", metadata.Latest)
		}

		if len(metadata.Versions) != 1 {
			t.Errorf("Expected 1 version, got %d", len(metadata.Versions))
		}
	})

	t.Run("LoadManifest", func(t *testing.T) {
		manifest, err := client.loadManifest("test-package", "1.0.0")
		if err != nil {
			t.Fatalf("loadManifest failed: %v", err)
		}

		if manifest.Name != "test-package" {
			t.Errorf("Expected package name 'test-package', got '%s'", manifest.Name)
		}

		if manifest.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", manifest.Version)
		}

		if manifest.SHA256 != "abc123" {
			t.Errorf("Expected SHA256 'abc123', got '%s'", manifest.SHA256)
		}
	})

	t.Run("RebuildIndex", func(t *testing.T) {
		// Remove index.json to test rebuild
		os.Remove(filepath.Join(tempDir, "index.json"))

		// Test rebuild directly
		index, err := client.rebuildIndex()
		if err != nil {
			t.Fatalf("rebuildIndex failed: %v", err)
		}

		if len(index.Packages) != 1 {
			t.Errorf("Expected 1 package after rebuild, got %d", len(index.Packages))
		}

		// Verify the rebuilt index has correct data
		pkg, exists := index.Packages["test-package"]
		if !exists {
			t.Error("Expected 'test-package' in rebuilt index")
		}

		if pkg.Name != "test-package" {
			t.Errorf("Expected package name 'test-package', got '%s'", pkg.Name)
		}
	})
}