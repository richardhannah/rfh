package client

import (
	"rulestack/internal/config"
	"testing"
)

func TestGetClient(t *testing.T) {
	t.Run("no active registry configured", func(t *testing.T) {
		cfg := config.CLIConfig{
			Current:    "",
			Registries: make(map[string]config.Registry),
		}

		_, err := GetClient(cfg, false)
		if err == nil {
			t.Error("expected error for no active registry")
		}
		if err.Error() != "no active registry configured" {
			t.Errorf("expected 'no active registry configured', got %q", err.Error())
		}
	})

	t.Run("active registry not found", func(t *testing.T) {
		cfg := config.CLIConfig{
			Current:    "missing",
			Registries: make(map[string]config.Registry),
		}

		_, err := GetClient(cfg, false)
		if err == nil {
			t.Error("expected error for missing registry")
		}
		expected := "active registry 'missing' not found in configuration"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestGetClientForRegistry(t *testing.T) {
	t.Run("registry not found", func(t *testing.T) {
		cfg := config.CLIConfig{
			Current:    "",
			Registries: make(map[string]config.Registry),
		}

		_, err := GetClientForRegistry(cfg, "nonexistent", false)
		if err == nil {
			t.Error("expected error for nonexistent registry")
		}
		expected := "registry 'nonexistent' not found"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestNewRegistryClient(t *testing.T) {
	t.Run("git registry returns not implemented", func(t *testing.T) {
		registry := config.Registry{
			URL:  "https://github.com/org/repo",
			Type: config.RegistryTypeGit,
		}

		_, err := NewRegistryClient(registry, false)
		if err == nil {
			t.Error("expected error for git registry (not yet implemented)")
		}
		if err.Error() != "Git registry client not yet implemented - will be added in Phase 5" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid registry type", func(t *testing.T) {
		registry := config.Registry{
			URL:  "https://example.com",
			Type: "invalid",
		}

		_, err := NewRegistryClient(registry, false)
		if err == nil {
			t.Error("expected error for invalid registry type")
		}
		expected := "unsupported registry type: invalid"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})
}