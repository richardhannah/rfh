package db

import (
	"testing"
)

func TestPackageFullPackageName(t *testing.T) {
	tests := []struct {
		name     string
		pkg      Package
		expected string
	}{
		{
			name: "package name",
			pkg: Package{
				Name: "test-rules",
			},
			expected: "test-rules",
		},
		{
			name: "package with complex name",
			pkg: Package{
				Name: "acme-security-rules",
			},
			expected: "acme-security-rules",
		},
		{
			name: "package with hyphens",
			pkg: Package{
				Name: "my-awesome-package",
			},
			expected: "my-awesome-package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pkg.FullPackageName()
			if result != tt.expected {
				t.Errorf("FullPackageName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestHashToken removed - legacy token functionality no longer supported