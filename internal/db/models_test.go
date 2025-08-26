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
			name: "scoped package",
			pkg: Package{
				Scope: stringPtr("acme"),
				Name:  "test-rules",
			},
			expected: "@acme/test-rules",
		},
		{
			name: "unscoped package",
			pkg: Package{
				Scope: nil,
				Name:  "test-rules",
			},
			expected: "test-rules",
		},
		{
			name: "empty scope",
			pkg: Package{
				Scope: stringPtr(""),
				Name:  "test-rules",
			},
			expected: "test-rules",
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

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}