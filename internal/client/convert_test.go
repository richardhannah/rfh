package client

import (
	"testing"
	"time"
)

func TestPackageToMap(t *testing.T) {
	updatedAt := time.Now()
	pkg := &Package{
		Name:        "test-package",
		Description: "A test package",
		Latest:      "1.0.0",
		Versions:    []string{"1.0.0", "0.9.0"},
		Tags:        []string{"security", "rules"},
		UpdatedAt:   updatedAt,
	}

	m := PackageToMap(pkg)

	if m["name"] != "test-package" {
		t.Errorf("expected name %q, got %q", "test-package", m["name"])
	}
	if m["description"] != "A test package" {
		t.Errorf("expected description %q, got %q", "A test package", m["description"])
	}
	if m["latest"] != "1.0.0" {
		t.Errorf("expected latest %q, got %q", "1.0.0", m["latest"])
	}
	if m["updated_at"] != updatedAt {
		t.Errorf("expected updated_at %v, got %v", updatedAt, m["updated_at"])
	}
}

func TestPackageVersionToMap(t *testing.T) {
	publishedAt := time.Now()
	pv := &PackageVersion{
		Name:        "test-package",
		Version:     "1.0.0",
		Description: "Test version",
		Dependencies: map[string]string{
			"core": "^2.0.0",
		},
		SHA256:      "abc123",
		Size:        1024,
		PublishedAt: publishedAt,
		Metadata: map[string]interface{}{
			"author": "test",
		},
	}

	m := PackageVersionToMap(pv)

	if m["name"] != "test-package" {
		t.Errorf("expected name %q, got %q", "test-package", m["name"])
	}
	if m["version"] != "1.0.0" {
		t.Errorf("expected version %q, got %q", "1.0.0", m["version"])
	}
	if m["sha256"] != "abc123" {
		t.Errorf("expected sha256 %q, got %q", "abc123", m["sha256"])
	}
	if m["size"] != int64(1024) {
		t.Errorf("expected size %d, got %v", 1024, m["size"])
	}
}

func TestMapToPackage(t *testing.T) {
	updatedAt := time.Now()
	m := map[string]interface{}{
		"name":        "test-package",
		"description": "A test package",
		"latest":      "1.0.0",
		"versions":    []interface{}{"1.0.0", "0.9.0"},
		"tags":        []interface{}{"security", "rules"},
		"updated_at":  updatedAt,
	}

	pkg := MapToPackage(m)

	if pkg.Name != "test-package" {
		t.Errorf("expected name %q, got %q", "test-package", pkg.Name)
	}
	if pkg.Description != "A test package" {
		t.Errorf("expected description %q, got %q", "A test package", pkg.Description)
	}
	if pkg.Latest != "1.0.0" {
		t.Errorf("expected latest %q, got %q", "1.0.0", pkg.Latest)
	}
	if len(pkg.Versions) != 2 || pkg.Versions[0] != "1.0.0" {
		t.Errorf("expected versions [1.0.0, 0.9.0], got %v", pkg.Versions)
	}
	if len(pkg.Tags) != 2 || pkg.Tags[0] != "security" {
		t.Errorf("expected tags [security, rules], got %v", pkg.Tags)
	}
	if pkg.UpdatedAt != updatedAt {
		t.Errorf("expected updated_at %v, got %v", updatedAt, pkg.UpdatedAt)
	}
}