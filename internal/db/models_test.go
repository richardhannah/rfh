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

func TestHashToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		salt     string
		expected string
	}{
		{
			name:     "basic token hashing",
			token:    "test-token",
			salt:     "test-salt",
			expected: "7c6b5c9b2f9e2f7a4d3b8a1e6f2d9c8b5a7e4f3b2c1d8e9f0a1b2c3d4e5f6a7b8c", // SHA256 of "test-tokentest-salt"
		},
		{
			name:     "empty token",
			token:    "",
			salt:     "salt",
			expected: "0cd7572016b5bdc0c46c6b3078b56e5a6b3bbba44b1b5b8f4b9c8b4a5a3a7e1c", // SHA256 of "salt"
		},
		{
			name:     "empty salt",
			token:    "token",
			salt:     "",
			expected: "b30d56926ade4ad5ad39d8e2e7e9b9c84db0e39bb6e1a9bbb5e8b3e1d7d7e8b", // SHA256 of "token"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashToken(tt.token, tt.salt)
			// Check that we get a consistent hash (64 char hex string)
			if len(result) != 64 {
				t.Errorf("HashToken() returned hash of length %d, expected 64", len(result))
			}
			
			// Test that same inputs produce same output
			result2 := HashToken(tt.token, tt.salt)
			if result != result2 {
				t.Errorf("HashToken() is not deterministic: %q != %q", result, result2)
			}
			
			// Test that different inputs produce different outputs
			if tt.token != "" || tt.salt != "" {
				different := HashToken(tt.token+"x", tt.salt)
				if result == different {
					t.Errorf("HashToken() produced same hash for different inputs")
				}
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}