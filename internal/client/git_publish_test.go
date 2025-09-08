package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGitPublishing(t *testing.T) {
	// Create a temporary directory for test repository
	tempDir := t.TempDir()

	// Create Git client with test directory as cache
	client := &GitClient{
		cacheDir: tempDir,
		verbose:  false, // Reduce noise in tests
	}

	t.Run("DetectFork", func(t *testing.T) {
		// Set test username
		os.Setenv("GITHUB_USERNAME", "test-user")
		defer os.Unsetenv("GITHUB_USERNAME")

		// Test valid GitHub URL
		fork, err := client.detectFork("https://github.com/owner/repo.git")
		if err != nil {
			t.Fatalf("detectFork failed: %v", err)
		}

		if fork.Username != "test-user" {
			t.Errorf("Expected username 'test-user', got '%s'", fork.Username)
		}

		if fork.RepoName != "repo" {
			t.Errorf("Expected repo name 'repo', got '%s'", fork.RepoName)
		}

		if fork.ForkURL != "https://github.com/test-user/repo.git" {
			t.Errorf("Expected fork URL 'https://github.com/test-user/repo.git', got '%s'", fork.ForkURL)
		}

		// Test non-GitHub URL
		_, err = client.detectFork("https://gitlab.com/owner/repo.git")
		if err == nil {
			t.Error("Expected error for non-GitHub URL")
		}
	})

	t.Run("CalculateFileInfo", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "test.txt")
		testContent := "test content"
		if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		hash, size, err := client.calculateFileInfo(testFile)
		if err != nil {
			t.Fatalf("calculateFileInfo failed: %v", err)
		}

		if size != int64(len(testContent)) {
			t.Errorf("Expected size %d, got %d", len(testContent), size)
		}

		if hash == "" {
			t.Error("Expected non-empty hash")
		}

		// Verify hash is consistent
		hash2, _, err := client.calculateFileInfo(testFile)
		if err != nil {
			t.Fatalf("Second calculateFileInfo failed: %v", err)
		}

		if hash != hash2 {
			t.Errorf("Hash inconsistent: %s != %s", hash, hash2)
		}
	})

	t.Run("UpdatePackageMetadata", func(t *testing.T) {
		// Create package directory
		packageDir := filepath.Join(tempDir, "test-package")
		if err := os.MkdirAll(packageDir, 0755); err != nil {
			t.Fatalf("Failed to create package directory: %v", err)
		}

		// Create test manifest
		manifest := &GitManifest{
			Name:         "test-package",
			Version:      "1.0.0",
			Description:  "A test package",
			SHA256:       "abcdef123456",
			Size:         1024,
			PublishedAt:  time.Now(),
			Publisher:    "test-user",
		}

		// Update metadata (first time - creates new)
		err := client.updatePackageMetadata(packageDir, manifest)
		if err != nil {
			t.Fatalf("updatePackageMetadata failed: %v", err)
		}

		// Verify metadata file was created
		metadataPath := filepath.Join(packageDir, "metadata.json")
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			t.Fatalf("Failed to read metadata file: %v", err)
		}

		var metadata GitPackageMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			t.Fatalf("Failed to parse metadata: %v", err)
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

		// Update with new version
		manifest.Version = "1.1.0"
		err = client.updatePackageMetadata(packageDir, manifest)
		if err != nil {
			t.Fatalf("Second updatePackageMetadata failed: %v", err)
		}

		// Verify metadata was updated
		data, _ = os.ReadFile(metadataPath)
		json.Unmarshal(data, &metadata)

		if metadata.Latest != "1.1.0" {
			t.Errorf("Expected latest version '1.1.0', got '%s'", metadata.Latest)
		}

		if len(metadata.Versions) != 2 {
			t.Errorf("Expected 2 versions, got %d", len(metadata.Versions))
		}
	})

	t.Run("GetAuthor", func(t *testing.T) {
		// Test with environment variables
		os.Setenv("GIT_AUTHOR_NAME", "Test User")
		os.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
		defer func() {
			os.Unsetenv("GIT_AUTHOR_NAME")
			os.Unsetenv("GIT_AUTHOR_EMAIL")
		}()

		author := client.getAuthor()

		if author.Name != "Test User" {
			t.Errorf("Expected author name 'Test User', got '%s'", author.Name)
		}

		if author.Email != "test@example.com" {
			t.Errorf("Expected author email 'test@example.com', got '%s'", author.Email)
		}

		// Test without environment variables (defaults)
		os.Unsetenv("GIT_AUTHOR_NAME")
		os.Unsetenv("GIT_AUTHOR_EMAIL")

		author = client.getAuthor()

		if author.Name != "RuleStack Publisher" {
			t.Errorf("Expected default author name 'RuleStack Publisher', got '%s'", author.Name)
		}

		if author.Email != "publisher@rulestack.dev" {
			t.Errorf("Expected default author email 'publisher@rulestack.dev', got '%s'", author.Email)
		}
	})

	t.Run("GetUsername", func(t *testing.T) {
		// Test with GITHUB_USERNAME
		os.Setenv("GITHUB_USERNAME", "github-user")
		defer os.Unsetenv("GITHUB_USERNAME")

		username := client.getUsername()
		if username != "github-user" {
			t.Errorf("Expected username 'github-user', got '%s'", username)
		}

		// Test with GIT_USER when GITHUB_USERNAME is not set
		os.Unsetenv("GITHUB_USERNAME")
		os.Setenv("GIT_USER", "git-user")
		defer os.Unsetenv("GIT_USER")

		username = client.getUsername()
		if username != "git-user" {
			t.Errorf("Expected username 'git-user', got '%s'", username)
		}

		// Test with no environment variables
		os.Unsetenv("GIT_USER")
		username = client.getUsername()
		// Should return empty string or git config value
		// We don't test git config here as it's system dependent
	})
}