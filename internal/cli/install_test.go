package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInstall_NoConfigFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "rfh-install-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create a rulestack.json with dependencies
	manifestContent := `{
		"version": "1.0.0",
		"dependencies": {
			"security-rules": "1.0.0"
		}
	}`
	manifestPath := filepath.Join(tempDir, "rulestack.json")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Ensure no config file exists by setting HOME to temp directory
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tempDir)

	// Run install command - should fail with "no registry configured"
	err = runInstall()
	if err == nil {
		t.Fatal("Expected install to fail with no registry configured, but it succeeded")
	}

	expectedError := "no registry configured. Use 'rfh registry add' to add a registry"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestRunInstall_NoCurrentRegistry(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "rfh-install-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create a rulestack.json with dependencies
	manifestContent := `{
		"version": "1.0.0",
		"dependencies": {
			"security-rules": "1.0.0"
		}
	}`
	manifestPath := filepath.Join(tempDir, "rulestack.json")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Set HOME to temp directory and create empty config
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tempDir)

	// Create .rfh directory and empty config file (no current registry)
	rfhDir := filepath.Join(tempDir, ".rfh")
	if err := os.MkdirAll(rfhDir, 0755); err != nil {
		t.Fatalf("Failed to create .rfh dir: %v", err)
	}

	configContent := `# Empty config with no registries`
	configPath := filepath.Join(rfhDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Run install command - should fail with "no registry configured"
	err = runInstall()
	if err == nil {
		t.Fatal("Expected install to fail with no registry configured, but it succeeded")
	}

	expectedError := "no registry configured. Use 'rfh registry add' to add a registry"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestRunInstall_NoDependencies(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "rfh-install-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create a rulestack.json with no dependencies
	manifestContent := `{
		"version": "1.0.0",
		"dependencies": {}
	}`
	manifestPath := filepath.Join(tempDir, "rulestack.json")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Run install command - should succeed and return early (no registry check needed)
	err = runInstall()
	if err != nil {
		t.Fatalf("Expected install to succeed with no dependencies, but got error: %v", err)
	}
}