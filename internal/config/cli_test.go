package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Errorf("ConfigDir() returned error: %v", err)
	}

	if dir == "" {
		t.Error("ConfigDir() returned empty string")
	}

	// Should end with .rfh
	if filepath.Base(dir) != ".rfh" {
		t.Errorf("ConfigDir() = %q, expected to end with .rfh", dir)
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Errorf("ConfigPath() returned error: %v", err)
	}

	if path == "" {
		t.Error("ConfigPath() returned empty string")
	}

	// Should end with config.toml
	if filepath.Base(path) != "config.toml" {
		t.Errorf("ConfigPath() = %q, expected to end with config.toml", path)
	}
}

func TestLoadCLI(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cli_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock ConfigPath by temporarily changing HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	t.Run("loads empty config when file doesn't exist", func(t *testing.T) {
		config, err := LoadCLI()
		if err != nil {
			t.Errorf("LoadCLI() returned error: %v", err)
		}

		if config.Current != "" {
			t.Errorf("expected empty current, got %q", config.Current)
		}

		if config.Registries == nil {
			t.Error("expected initialized registries map")
		}

		if len(config.Registries) != 0 {
			t.Errorf("expected empty registries, got %d", len(config.Registries))
		}
	})

	t.Run("loads valid config file", func(t *testing.T) {
		// Create config directory and file
		configDir := filepath.Join(tempDir, ".rfh")
		configPath := filepath.Join(configDir, "config.toml")
		
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("failed to create config dir: %v", err)
		}

		configContent := `current = "local"

[registries.local]
url = "http://localhost:8080"
jwt_token = "test-jwt-token"

[registries.public]
url = "https://registry.example.com"
`
		err = os.WriteFile(configPath, []byte(configContent), 0600)
		if err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		config, err := LoadCLI()
		if err != nil {
			t.Errorf("LoadCLI() returned error: %v", err)
		}

		if config.Current != "local" {
			t.Errorf("expected current 'local', got %q", config.Current)
		}

		if len(config.Registries) != 2 {
			t.Errorf("expected 2 registries, got %d", len(config.Registries))
		}

		localReg, exists := config.Registries["local"]
		if !exists {
			t.Error("expected 'local' registry to exist")
		} else {
			if localReg.URL != "http://localhost:8080" {
				t.Errorf("expected local URL 'http://localhost:8080', got %q", localReg.URL)
			}
			if localReg.JWTToken != "test-jwt-token" {
				t.Errorf("expected local JWT token 'test-jwt-token', got %q", localReg.JWTToken)
			}
		}

		publicReg, exists := config.Registries["public"]
		if !exists {
			t.Error("expected 'public' registry to exist")
		} else {
			if publicReg.URL != "https://registry.example.com" {
				t.Errorf("expected public URL 'https://registry.example.com', got %q", publicReg.URL)
			}
			if publicReg.JWTToken != "" {
				t.Errorf("expected empty public JWT token, got %q", publicReg.JWTToken)
			}
		}
	})

	t.Run("handles invalid TOML", func(t *testing.T) {
		// Create config directory and invalid file
		configDir := filepath.Join(tempDir, ".rfh")
		configPath := filepath.Join(configDir, "config.toml")
		
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("failed to create config dir: %v", err)
		}

		invalidContent := `invalid toml content [[[`
		err = os.WriteFile(configPath, []byte(invalidContent), 0600)
		if err != nil {
			t.Fatalf("failed to write invalid config file: %v", err)
		}

		_, err = LoadCLI()
		if err == nil {
			t.Error("expected error for invalid TOML")
		}
	})
}

func TestSaveCLI(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cli_save_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock ConfigPath by temporarily changing HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	t.Run("saves config successfully", func(t *testing.T) {
		config := CLIConfig{
			Current: "test",
			Registries: map[string]Registry{
				"test": {
					URL:      "https://test.example.com",
					JWTToken: "secret-jwt-token",
				},
				"public": {
					URL: "https://public.example.com",
				},
			},
		}

		err := SaveCLI(config)
		if err != nil {
			t.Errorf("SaveCLI() returned error: %v", err)
		}

		// Verify file was created
		configPath, _ := ConfigPath()
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Load and verify content
		loadedConfig, err := LoadCLI()
		if err != nil {
			t.Errorf("failed to load saved config: %v", err)
		}

		if loadedConfig.Current != config.Current {
			t.Errorf("current mismatch: expected %q, got %q", config.Current, loadedConfig.Current)
		}

		if len(loadedConfig.Registries) != len(config.Registries) {
			t.Errorf("registries count mismatch: expected %d, got %d", len(config.Registries), len(loadedConfig.Registries))
		}

		for name, expectedReg := range config.Registries {
			loadedReg, exists := loadedConfig.Registries[name]
			if !exists {
				t.Errorf("registry %q not found in loaded config", name)
				continue
			}

			if loadedReg.URL != expectedReg.URL {
				t.Errorf("registry %q URL mismatch: expected %q, got %q", name, expectedReg.URL, loadedReg.URL)
			}

			if loadedReg.JWTToken != expectedReg.JWTToken {
				t.Errorf("registry %q JWT token mismatch: expected %q, got %q", name, expectedReg.JWTToken, loadedReg.JWTToken)
			}
		}
	})

	t.Run("creates directory if it doesn't exist", func(t *testing.T) {
		// Remove the .rfh directory if it exists
		configDir := filepath.Join(tempDir, ".rfh")
		os.RemoveAll(configDir)

		config := CLIConfig{
			Current:    "test",
			Registries: map[string]Registry{},
		}

		err := SaveCLI(config)
		if err != nil {
			t.Errorf("SaveCLI() returned error: %v", err)
		}

		// Verify directory was created
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			t.Error("config directory was not created")
		}
	})
}