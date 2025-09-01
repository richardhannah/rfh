package version

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    *Version
		wantErr bool
	}{
		{
			name:    "basic version",
			version: "1.2.3",
			want:    &Version{Major: 1, Minor: 2, Patch: 3},
			wantErr: false,
		},
		{
			name:    "version with pre-release",
			version: "1.2.3-alpha",
			want:    &Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
			wantErr: false,
		},
		{
			name:    "version with build",
			version: "1.2.3+build.1",
			want:    &Version{Major: 1, Minor: 2, Patch: 3, Build: "build.1"},
			wantErr: false,
		},
		{
			name:    "version with pre-release and build",
			version: "1.2.3-beta.2+build.123",
			want:    &Version{Major: 1, Minor: 2, Patch: 3, Pre: "beta.2", Build: "build.123"},
			wantErr: false,
		},
		{
			name:    "zero version",
			version: "0.0.0",
			want:    &Version{Major: 0, Minor: 0, Patch: 0},
			wantErr: false,
		},
		{
			name:    "large version numbers",
			version: "999.999.999",
			want:    &Version{Major: 999, Minor: 999, Patch: 999},
			wantErr: false,
		},
		{
			name:    "empty string",
			version: "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - two parts",
			version: "1.2",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - four parts",
			version: "1.2.3.4",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid major version",
			version: "a.2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid minor version",
			version: "1.b.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid patch version",
			version: "1.2.c",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "negative major version",
			version: "-1.2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "negative minor version",
			version: "1.-2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "negative patch version",
			version: "1.2.-3",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Major != tt.want.Major || got.Minor != tt.want.Minor || 
				   got.Patch != tt.want.Patch || got.Pre != tt.want.Pre || got.Build != tt.want.Build {
					t.Errorf("Parse() = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name    string
		version *Version
		want    string
	}{
		{
			name:    "basic version",
			version: &Version{Major: 1, Minor: 2, Patch: 3},
			want:    "1.2.3",
		},
		{
			name:    "version with pre-release",
			version: &Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
			want:    "1.2.3-alpha",
		},
		{
			name:    "version with build",
			version: &Version{Major: 1, Minor: 2, Patch: 3, Build: "build.1"},
			want:    "1.2.3+build.1",
		},
		{
			name:    "version with pre-release and build",
			version: &Version{Major: 1, Minor: 2, Patch: 3, Pre: "beta.2", Build: "build.123"},
			want:    "1.2.3-beta.2+build.123",
		},
		{
			name:    "zero version",
			version: &Version{Major: 0, Minor: 0, Patch: 0},
			want:    "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		want     int
	}{
		// Basic comparisons
		{name: "1.0.0 vs 1.0.0", version1: "1.0.0", version2: "1.0.0", want: 0},
		{name: "1.0.1 vs 1.0.0", version1: "1.0.1", version2: "1.0.0", want: 1},
		{name: "1.0.0 vs 1.0.1", version1: "1.0.0", version2: "1.0.1", want: -1},
		{name: "1.1.0 vs 1.0.0", version1: "1.1.0", version2: "1.0.0", want: 1},
		{name: "1.0.0 vs 1.1.0", version1: "1.0.0", version2: "1.1.0", want: -1},
		{name: "2.0.0 vs 1.0.0", version1: "2.0.0", version2: "1.0.0", want: 1},
		{name: "1.0.0 vs 2.0.0", version1: "1.0.0", version2: "2.0.0", want: -1},
		
		// Pre-release comparisons
		{name: "1.0.0-alpha vs 1.0.0", version1: "1.0.0-alpha", version2: "1.0.0", want: -1},
		{name: "1.0.0 vs 1.0.0-alpha", version1: "1.0.0", version2: "1.0.0-alpha", want: 1},
		{name: "1.0.0-alpha vs 1.0.0-beta", version1: "1.0.0-alpha", version2: "1.0.0-beta", want: -1},
		{name: "1.0.0-beta vs 1.0.0-alpha", version1: "1.0.0-beta", version2: "1.0.0-alpha", want: 1},
		{name: "1.0.0-alpha vs 1.0.0-alpha", version1: "1.0.0-alpha", version2: "1.0.0-alpha", want: 0},
		
		// Build metadata should be ignored
		{name: "1.0.0+build1 vs 1.0.0+build2", version1: "1.0.0+build1", version2: "1.0.0+build2", want: 0},
		{name: "1.0.0+build vs 1.0.0", version1: "1.0.0+build", version2: "1.0.0", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := Parse(tt.version1)
			if err != nil {
				t.Fatalf("Parse(%s) failed: %v", tt.version1, err)
			}
			v2, err := Parse(tt.version2)
			if err != nil {
				t.Fatalf("Parse(%s) failed: %v", tt.version2, err)
			}
			
			if got := v1.Compare(v2); got != tt.want {
				t.Errorf("Version.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_Increment(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		wantPatch string
		wantMinor string
		wantMajor string
	}{
		{
			name:      "basic version",
			version:   "1.2.3",
			wantPatch: "1.2.4",
			wantMinor: "1.3.0",
			wantMajor: "2.0.0",
		},
		{
			name:      "zero version",
			version:   "0.0.0",
			wantPatch: "0.0.1",
			wantMinor: "0.1.0",
			wantMajor: "1.0.0",
		},
		{
			name:      "version with pre-release",
			version:   "1.2.3-alpha",
			wantPatch: "1.2.4",
			wantMinor: "1.3.0",
			wantMajor: "2.0.0",
		},
		{
			name:      "version with build",
			version:   "1.2.3+build.1",
			wantPatch: "1.2.4",
			wantMinor: "1.3.0",
			wantMajor: "2.0.0",
		},
		{
			name:      "version with pre-release and build",
			version:   "1.2.3-beta+build.1",
			wantPatch: "1.2.4",
			wantMinor: "1.3.0",
			wantMajor: "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Parse(tt.version)
			if err != nil {
				t.Fatalf("Parse(%s) failed: %v", tt.version, err)
			}

			if got := v.IncrementPatch().String(); got != tt.wantPatch {
				t.Errorf("IncrementPatch() = %v, want %v", got, tt.wantPatch)
			}

			if got := v.IncrementMinor().String(); got != tt.wantMinor {
				t.Errorf("IncrementMinor() = %v, want %v", got, tt.wantMinor)
			}

			if got := v.IncrementMajor().String(); got != tt.wantMajor {
				t.Errorf("IncrementMajor() = %v, want %v", got, tt.wantMajor)
			}
		})
	}
}

func TestValidateVersionIncrease(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		newVersion     string
		wantErr        bool
	}{
		{name: "patch increment", currentVersion: "1.0.0", newVersion: "1.0.1", wantErr: false},
		{name: "minor increment", currentVersion: "1.0.0", newVersion: "1.1.0", wantErr: false},
		{name: "major increment", currentVersion: "1.0.0", newVersion: "2.0.0", wantErr: false},
		{name: "same version", currentVersion: "1.0.0", newVersion: "1.0.0", wantErr: true},
		{name: "patch decrease", currentVersion: "1.0.1", newVersion: "1.0.0", wantErr: true},
		{name: "minor decrease", currentVersion: "1.1.0", newVersion: "1.0.0", wantErr: true},
		{name: "major decrease", currentVersion: "2.0.0", newVersion: "1.0.0", wantErr: true},
		{name: "pre-release to release", currentVersion: "1.0.0-alpha", newVersion: "1.0.0", wantErr: false},
		{name: "release to pre-release", currentVersion: "1.0.0", newVersion: "1.0.0-alpha", wantErr: true},
		{name: "invalid current version", currentVersion: "invalid", newVersion: "1.0.0", wantErr: true},
		{name: "invalid new version", currentVersion: "1.0.0", newVersion: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersionIncrease(tt.currentVersion, tt.newVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersionIncrease() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIncrementPatchVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
		wantErr bool
	}{
		{name: "basic version", version: "1.2.3", want: "1.2.4", wantErr: false},
		{name: "zero patch", version: "1.2.0", want: "1.2.1", wantErr: false},
		{name: "with pre-release", version: "1.2.3-alpha", want: "1.2.4", wantErr: false},
		{name: "with build", version: "1.2.3+build", want: "1.2.4", wantErr: false},
		{name: "invalid version", version: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IncrementPatchVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncrementPatchVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IncrementPatchVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIncrementMinorVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
		wantErr bool
	}{
		{name: "basic version", version: "1.2.3", want: "1.3.0", wantErr: false},
		{name: "zero minor", version: "1.0.3", want: "1.1.0", wantErr: false},
		{name: "with pre-release", version: "1.2.3-alpha", want: "1.3.0", wantErr: false},
		{name: "with build", version: "1.2.3+build", want: "1.3.0", wantErr: false},
		{name: "invalid version", version: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IncrementMinorVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncrementMinorVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IncrementMinorVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIncrementMajorVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
		wantErr bool
	}{
		{name: "basic version", version: "1.2.3", want: "2.0.0", wantErr: false},
		{name: "zero major", version: "0.2.3", want: "1.0.0", wantErr: false},
		{name: "with pre-release", version: "1.2.3-alpha", want: "2.0.0", wantErr: false},
		{name: "with build", version: "1.2.3+build", want: "2.0.0", wantErr: false},
		{name: "invalid version", version: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IncrementMajorVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncrementMajorVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IncrementMajorVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		want     int
		wantErr  bool
	}{
		{name: "equal versions", version1: "1.0.0", version2: "1.0.0", want: 0, wantErr: false},
		{name: "first greater", version1: "1.0.1", version2: "1.0.0", want: 1, wantErr: false},
		{name: "first smaller", version1: "1.0.0", version2: "1.0.1", want: -1, wantErr: false},
		{name: "invalid first version", version1: "invalid", version2: "1.0.0", want: 0, wantErr: true},
		{name: "invalid second version", version1: "1.0.0", version2: "invalid", want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareVersions(tt.version1, tt.version2)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{name: "valid basic version", version: "1.0.0", want: true},
		{name: "valid with pre-release", version: "1.0.0-alpha", want: true},
		{name: "valid with build", version: "1.0.0+build", want: true},
		{name: "valid with pre-release and build", version: "1.0.0-alpha+build", want: true},
		{name: "invalid empty", version: "", want: false},
		{name: "invalid format", version: "1.0", want: false},
		{name: "invalid characters", version: "a.b.c", want: false},
		{name: "invalid negative", version: "-1.0.0", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidVersion(tt.version); got != tt.want {
				t.Errorf("IsValidVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNextVersions(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantPatch string
		wantMinor string
		wantMajor string
		wantErr   bool
	}{
		{
			name:      "basic version",
			version:   "1.2.3",
			wantPatch: "1.2.4",
			wantMinor: "1.3.0",
			wantMajor: "2.0.0",
			wantErr:   false,
		},
		{
			name:      "version with pre-release",
			version:   "1.2.3-alpha",
			wantPatch: "1.2.4",
			wantMinor: "1.3.0",
			wantMajor: "2.0.0",
			wantErr:   false,
		},
		{
			name:    "invalid version",
			version: "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPatch, gotMinor, gotMajor, err := GetNextVersions(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNextVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotPatch != tt.wantPatch {
					t.Errorf("GetNextVersions() patch = %v, want %v", gotPatch, tt.wantPatch)
				}
				if gotMinor != tt.wantMinor {
					t.Errorf("GetNextVersions() minor = %v, want %v", gotMinor, tt.wantMinor)
				}
				if gotMajor != tt.wantMajor {
					t.Errorf("GetNextVersions() major = %v, want %v", gotMajor, tt.wantMajor)
				}
			}
		})
	}
}