package version

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string // Pre-release identifier (e.g., "alpha", "beta.1")
	Build string // Build metadata (e.g., "20230101.abcd123")
}

// Parse parses a semantic version string into a Version struct
func Parse(versionStr string) (*Version, error) {
	if versionStr == "" {
		return nil, fmt.Errorf("version cannot be empty")
	}

	// Handle build metadata (+)
	var buildMeta string
	if idx := strings.Index(versionStr, "+"); idx != -1 {
		buildMeta = versionStr[idx+1:]
		versionStr = versionStr[:idx]
	}

	// Handle pre-release (-)
	var preRelease string
	if idx := strings.Index(versionStr, "-"); idx != -1 {
		preRelease = versionStr[idx+1:]
		versionStr = versionStr[:idx]
	}

	// Parse major.minor.patch
	parts := strings.Split(versionStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: expected x.y.z, got %s", versionStr)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil || minor < 0 {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil || patch < 0 {
		return nil, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   preRelease,
		Build: buildMeta,
	}, nil
}

// String returns the string representation of the version
func (v *Version) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)

	if v.Pre != "" {
		result += "-" + v.Pre
	}

	if v.Build != "" {
		result += "+" + v.Build
	}

	return result
}

// Compare compares two versions and returns:
// -1 if v < other
//
//	0 if v == other
//	1 if v > other
func (v *Version) Compare(other *Version) int {
	// Compare major.minor.patch
	if v.Major != other.Major {
		if v.Major > other.Major {
			return 1
		}
		return -1
	}

	if v.Minor != other.Minor {
		if v.Minor > other.Minor {
			return 1
		}
		return -1
	}

	if v.Patch != other.Patch {
		if v.Patch > other.Patch {
			return 1
		}
		return -1
	}

	// Handle pre-release versions
	// Per semver: 1.0.0-alpha < 1.0.0
	if v.Pre == "" && other.Pre != "" {
		return 1 // Normal version > pre-release
	}
	if v.Pre != "" && other.Pre == "" {
		return -1 // Pre-release < normal version
	}
	if v.Pre != "" && other.Pre != "" {
		// Both are pre-releases, compare lexicographically
		if v.Pre > other.Pre {
			return 1
		} else if v.Pre < other.Pre {
			return -1
		}
	}

	// Build metadata is ignored in precedence comparison
	return 0
}

// IsGreaterThan returns true if v > other
func (v *Version) IsGreaterThan(other *Version) bool {
	return v.Compare(other) > 0
}

// IsLessThan returns true if v < other
func (v *Version) IsLessThan(other *Version) bool {
	return v.Compare(other) < 0
}

// IsEqual returns true if v == other (ignoring build metadata)
func (v *Version) IsEqual(other *Version) bool {
	return v.Compare(other) == 0
}

// IncrementPatch returns a new version with patch incremented
func (v *Version) IncrementPatch() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch + 1,
		// Clear pre-release and build when incrementing
		Pre:   "",
		Build: "",
	}
}

// IncrementMinor returns a new version with minor incremented and patch reset
func (v *Version) IncrementMinor() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor + 1,
		Patch: 0,
		Pre:   "",
		Build: "",
	}
}

// IncrementMajor returns a new version with major incremented and minor/patch reset
func (v *Version) IncrementMajor() *Version {
	return &Version{
		Major: v.Major + 1,
		Minor: 0,
		Patch: 0,
		Pre:   "",
		Build: "",
	}
}

// CONVENIENCE FUNCTIONS FOR BACKWARD COMPATIBILITY

// ValidateVersionIncrease ensures newVersion is greater than currentVersion
// This is the function previously in pack_interactive.go
func ValidateVersionIncrease(currentVersion, newVersion string) error {
	current, err := Parse(currentVersion)
	if err != nil {
		return fmt.Errorf("invalid current version %s: %w", currentVersion, err)
	}

	new, err := Parse(newVersion)
	if err != nil {
		return fmt.Errorf("invalid new version %s: %w", newVersion, err)
	}

	if !new.IsGreaterThan(current) {
		return fmt.Errorf("new version %s must be greater than current version %s", newVersion, currentVersion)
	}

	return nil
}

// IncrementPatchVersion increments the patch version (x.y.z -> x.y.z+1)
// This is the function previously in pack_interactive.go
func IncrementPatchVersion(versionStr string) (string, error) {
	v, err := Parse(versionStr)
	if err != nil {
		return "", fmt.Errorf("invalid version format: %w", err)
	}

	return v.IncrementPatch().String(), nil
}

// IncrementMinorVersion increments the minor version (x.y.z -> x.y+1.0)
func IncrementMinorVersion(versionStr string) (string, error) {
	v, err := Parse(versionStr)
	if err != nil {
		return "", fmt.Errorf("invalid version format: %w", err)
	}

	return v.IncrementMinor().String(), nil
}

// IncrementMajorVersion increments the major version (x.y.z -> x+1.0.0)
func IncrementMajorVersion(versionStr string) (string, error) {
	v, err := Parse(versionStr)
	if err != nil {
		return "", fmt.Errorf("invalid version format: %w", err)
	}

	return v.IncrementMajor().String(), nil
}

// CompareVersions compares two version strings and returns:
// -1 if version1 < version2
//
//	0 if version1 == version2
//	1 if version1 > version2
func CompareVersions(version1, version2 string) (int, error) {
	v1, err := Parse(version1)
	if err != nil {
		return 0, fmt.Errorf("invalid version1 %s: %w", version1, err)
	}

	v2, err := Parse(version2)
	if err != nil {
		return 0, fmt.Errorf("invalid version2 %s: %w", version2, err)
	}

	return v1.Compare(v2), nil
}

// IsValidVersion checks if a string is a valid semantic version
func IsValidVersion(versionStr string) bool {
	_, err := Parse(versionStr)
	return err == nil
}

// GetNextVersions returns suggested next versions for all increment types
func GetNextVersions(currentVersion string) (patch, minor, major string, err error) {
	v, err := Parse(currentVersion)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid version: %w", err)
	}

	return v.IncrementPatch().String(),
		v.IncrementMinor().String(),
		v.IncrementMajor().String(),
		nil
}
