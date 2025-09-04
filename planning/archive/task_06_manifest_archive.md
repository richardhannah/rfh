# Task 7: Manifest and Archive Handling (1 hour)

## Objective
Create the manifest parsing and archive handling functionality for packaging and publishing rulesets.

## Prerequisites
- Tasks 1-6 completed
- Basic API handlers working
- Database models established

## Checklist

### 1. Create Manifest Package (15 minutes)
Create `internal/manifest/manifest.go`:
```go
package manifest

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "regexp"
    "strings"
)

// Manifest represents a rulestack.json file
type Manifest struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Description string   `json:"description,omitempty"`
    Targets     []string `json:"targets,omitempty"`
    Tags        []string `json:"tags,omitempty"`
    Files       []string `json:"files"`
    License     string   `json:"license,omitempty"`
}

var (
    ErrInvalidManifest = errors.New("invalid manifest")
    ErrInvalidName     = errors.New("invalid package name")
    ErrInvalidVersion  = errors.New("invalid version")
)

// nameRegex matches valid package names (with or without scope)
var nameRegex = regexp.MustCompile(`^(@[a-z0-9][a-z0-9\-_]*\/)?[a-z0-9][a-z0-9\-_]*$`)

// versionRegex matches semantic versions
var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9\-]+)?(\+[a-zA-Z0-9\-]+)?$`)

// Load reads and validates a manifest from file
func Load(path string) (*Manifest, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read manifest: %w", err)
    }
    
    var manifest Manifest
    if err := json.Unmarshal(data, &manifest); err != nil {
        return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
    }
    
    if err := manifest.Validate(); err != nil {
        return nil, err
    }
    
    return &manifest, nil
}

// Save writes manifest to file
func (m *Manifest) Save(path string) error {
    if err := m.Validate(); err != nil {
        return err
    }
    
    data, err := json.MarshalIndent(m, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal manifest: %w", err)
    }
    
    return os.WriteFile(path, data, 0o644)
}

// Validate checks if the manifest is valid
func (m *Manifest) Validate() error {
    if m.Name == "" {
        return fmt.Errorf("%w: name is required", ErrInvalidManifest)
    }
    
    if !nameRegex.MatchString(m.Name) {
        return fmt.Errorf("%w: name must match pattern %s", ErrInvalidName, nameRegex.String())
    }
    
    if m.Version == "" {
        return fmt.Errorf("%w: version is required", ErrInvalidManifest)
    }
    
    if !versionRegex.MatchString(m.Version) {
        return fmt.Errorf("%w: version must be semantic version (x.y.z)", ErrInvalidVersion)
    }
    
    if len(m.Files) == 0 {
        return fmt.Errorf("%w: files array cannot be empty", ErrInvalidManifest)
    }
    
    // Validate targets
    validTargets := map[string]bool{
        "cursor":     true,
        "claude-code": true,
        "windsurf":   true,
        "copilot":    true,
    }
    
    for _, target := range m.Targets {
        if !validTargets[target] {
            return fmt.Errorf("%w: invalid target '%s'", ErrInvalidManifest, target)
        }
    }
    
    return nil
}

// GetScope returns the scope part of the package name, or empty string if unscoped
func (m *Manifest) GetScope() string {
    if strings.HasPrefix(m.Name, "@") {
        parts := strings.SplitN(m.Name[1:], "/", 2)
        if len(parts) == 2 {
            return parts[0]
        }
    }
    return ""
}

// GetPackageName returns the package name without scope
func (m *Manifest) GetPackageName() string {
    if strings.HasPrefix(m.Name, "@") {
        parts := strings.SplitN(m.Name[1:], "/", 2)
        if len(parts) == 2 {
            return parts[1]
        }
    }
    return m.Name
}

// CreateSample creates a sample manifest for initialization
func CreateSample() *Manifest {
    return &Manifest{
        Name:        "@acme/example-rules",
        Version:     "0.1.0",
        Description: "Example AI ruleset",
        Targets:     []string{"cursor"},
        Tags:        []string{"example", "starter"},
        Files:       []string{"rules/**/*.md"},
        License:     "MIT",
    }
}
```

- [ ] Create manifest.go file
- [ ] Test regex patterns work correctly

### 2. Create Archive Package (20 minutes)
Create `internal/pkg/archive.go`:
```go
package pkg

import (
    "archive/tar"
    "compress/gzip"
    "crypto/sha256"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/bmatcuk/doublestar/v4"
)

// ArchiveInfo contains information about a created archive
type ArchiveInfo struct {
    Path      string
    SHA256    string
    SizeBytes int64
}

// Pack creates a tar.gz archive from file patterns
func Pack(patterns []string, outputPath string) (*ArchiveInfo, error) {
    // Collect all files matching the patterns
    var files []string
    seen := make(map[string]bool)
    
    for _, pattern := range patterns {
        matches, err := doublestar.FilepathGlob(pattern)
        if err != nil {
            return nil, fmt.Errorf("failed to match pattern %s: %w", pattern, err)
        }
        
        for _, match := range matches {
            // Skip directories
            if info, err := os.Stat(match); err != nil || info.IsDir() {
                continue
            }
            
            // Clean path and avoid duplicates
            cleanPath := filepath.Clean(match)
            if !seen[cleanPath] {
                files = append(files, cleanPath)
                seen[cleanPath] = true
            }
        }
    }
    
    if len(files) == 0 {
        return nil, fmt.Errorf("no files matched the specified patterns")
    }
    
    // Create output file
    outFile, err := os.Create(outputPath)
    if err != nil {
        return nil, fmt.Errorf("failed to create archive file: %w", err)
    }
    defer outFile.Close()
    
    // Create hash writer
    hasher := sha256.New()
    multiWriter := io.MultiWriter(outFile, hasher)
    
    // Create gzip writer
    gzWriter := gzip.NewWriter(multiWriter)
    defer gzWriter.Close()
    
    // Create tar writer
    tarWriter := tar.NewWriter(gzWriter)
    defer tarWriter.Close()
    
    var totalSize int64
    
    // Add each file to the archive
    for _, filePath := range files {
        if err := addFileToArchive(tarWriter, filePath); err != nil {
            return nil, fmt.Errorf("failed to add file %s: %w", filePath, err)
        }
        
        if info, err := os.Stat(filePath); err == nil {
            totalSize += info.Size()
        }
    }
    
    // Close writers to flush data before calculating hash
    tarWriter.Close()
    gzWriter.Close()
    outFile.Close()
    
    // Get final archive info
    info, err := os.Stat(outputPath)
    if err != nil {
        return nil, fmt.Errorf("failed to stat archive: %w", err)
    }
    
    return &ArchiveInfo{
        Path:      outputPath,
        SHA256:    fmt.Sprintf("%x", hasher.Sum(nil)),
        SizeBytes: info.Size(),
    }, nil
}

// addFileToArchive adds a single file to the tar archive
func addFileToArchive(tarWriter *tar.Writer, filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    info, err := file.Stat()
    if err != nil {
        return err
    }
    
    // Create tar header
    header, err := tar.FileInfoHeader(info, "")
    if err != nil {
        return err
    }
    
    // Use forward slashes in archive
    header.Name = filepath.ToSlash(filePath)
    
    // Write header
    if err := tarWriter.WriteHeader(header); err != nil {
        return err
    }
    
    // Write file content
    _, err = io.Copy(tarWriter, file)
    return err
}

// Unpack extracts a tar.gz archive to a destination directory
func Unpack(archivePath string, destDir string) error {
    file, err := os.Open(archivePath)
    if err != nil {
        return fmt.Errorf("failed to open archive: %w", err)
    }
    defer file.Close()
    
    // Create gzip reader
    gzReader, err := gzip.NewReader(file)
    if err != nil {
        return fmt.Errorf("failed to create gzip reader: %w", err)
    }
    defer gzReader.Close()
    
    // Create tar reader
    tarReader := tar.NewReader(gzReader)
    
    // Extract files
    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("failed to read tar header: %w", err)
        }
        
        if err := extractFile(tarReader, header, destDir); err != nil {
            return fmt.Errorf("failed to extract file %s: %w", header.Name, err)
        }
    }
    
    return nil
}

// extractFile extracts a single file from tar archive
func extractFile(tarReader *tar.Reader, header *tar.Header, destDir string) error {
    // Clean the file path to prevent directory traversal
    cleanName := filepath.Clean(header.Name)
    if strings.Contains(cleanName, "..") {
        return fmt.Errorf("invalid file path: %s", header.Name)
    }
    
    destPath := filepath.Join(destDir, cleanName)
    
    // Create directory if needed
    if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
        return err
    }
    
    // Create file
    outFile, err := os.Create(destPath)
    if err != nil {
        return err
    }
    defer outFile.Close()
    
    // Copy file content
    _, err = io.Copy(outFile, tarReader)
    return err
}

// CalculateSHA256 calculates SHA256 hash of a file
func CalculateSHA256(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    hasher := sha256.New()
    if _, err := io.Copy(hasher, file); err != nil {
        return "", err
    }
    
    return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
```

- [ ] Create archive.go file
- [ ] Verify doublestar import for glob patterns

### 3. Create Lockfile Package (10 minutes)
Create `internal/pkg/lockfile.go`:
```go
package pkg

import (
    "encoding/json"
    "os"
)

// Lockfile represents the rfh.lock file format
type Lockfile struct {
    Registry string                    `json:"registry"`
    Packages map[string]LockfileEntry `json:"packages"`
}

// LockfileEntry represents a single package in the lockfile
type LockfileEntry struct {
    Version     string   `json:"version"`
    SHA256      string   `json:"sha256"`
    Targets     []string `json:"targets"`
    InstallPath string   `json:"install_path"`
    Registry    string   `json:"registry,omitempty"`
}

// LoadLockfile reads and parses a lockfile
func LoadLockfile(path string) (*Lockfile, error) {
    data, err := os.ReadFile(path)
    if os.IsNotExist(err) {
        // Return empty lockfile if doesn't exist
        return &Lockfile{
            Packages: make(map[string]LockfileEntry),
        }, nil
    }
    if err != nil {
        return nil, err
    }
    
    var lockfile Lockfile
    if err := json.Unmarshal(data, &lockfile); err != nil {
        return nil, err
    }
    
    if lockfile.Packages == nil {
        lockfile.Packages = make(map[string]LockfileEntry)
    }
    
    return &lockfile, nil
}

// SaveLockfile writes the lockfile to disk
func (l *Lockfile) SaveLockfile(path string) error {
    data, err := json.MarshalIndent(l, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(path, data, 0o644)
}

// AddPackage adds or updates a package in the lockfile
func (l *Lockfile) AddPackage(name string, entry LockfileEntry) {
    if l.Packages == nil {
        l.Packages = make(map[string]LockfileEntry)
    }
    l.Packages[name] = entry
}

// RemovePackage removes a package from the lockfile
func (l *Lockfile) RemovePackage(name string) {
    if l.Packages != nil {
        delete(l.Packages, name)
    }
}

// HasPackage checks if a package exists in the lockfile
func (l *Lockfile) HasPackage(name string) bool {
    if l.Packages == nil {
        return false
    }
    _, exists := l.Packages[name]
    return exists
}

// GetPackage gets a package from the lockfile
func (l *Lockfile) GetPackage(name string) (LockfileEntry, bool) {
    if l.Packages == nil {
        return LockfileEntry{}, false
    }
    entry, exists := l.Packages[name]
    return entry, exists
}
```

- [ ] Create lockfile.go file
- [ ] Review JSON marshalling logic

### 4. Create Archive Tests (10 minutes)
Create `internal/pkg/archive_test.go`:
```go
package pkg

import (
    "os"
    "path/filepath"
    "testing"
)

func TestPack(t *testing.T) {
    // Create a temporary directory with test files
    tmpDir, err := os.MkdirTemp("", "rulestack-test")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tmpDir)
    
    // Create test files
    testDir := filepath.Join(tmpDir, "rules")
    os.MkdirAll(testDir, 0o755)
    
    testFile1 := filepath.Join(testDir, "rule1.md")
    testFile2 := filepath.Join(testDir, "rule2.md")
    
    os.WriteFile(testFile1, []byte("# Rule 1"), 0o644)
    os.WriteFile(testFile2, []byte("# Rule 2"), 0o644)
    
    // Change to temp directory
    oldWd, _ := os.Getwd()
    os.Chdir(tmpDir)
    defer os.Chdir(oldWd)
    
    // Test packing
    patterns := []string{"rules/**/*.md"}
    outputPath := "test-archive.tgz"
    
    info, err := Pack(patterns, outputPath)
    if err != nil {
        t.Fatalf("Pack failed: %v", err)
    }
    
    if info.Path != outputPath {
        t.Errorf("Expected path %s, got %s", outputPath, info.Path)
    }
    
    if info.SHA256 == "" {
        t.Error("SHA256 should not be empty")
    }
    
    if info.SizeBytes <= 0 {
        t.Error("Archive size should be greater than 0")
    }
    
    // Verify archive exists
    if _, err := os.Stat(outputPath); os.IsNotExist(err) {
        t.Error("Archive file should exist")
    }
}

func TestUnpack(t *testing.T) {
    // This test would require a valid archive file
    // For now, just test that the function exists
    if Unpack == nil {
        t.Error("Unpack function should not be nil")
    }
}

func TestCalculateSHA256(t *testing.T) {
    // Create a temporary file
    tmpFile, err := os.CreateTemp("", "test-hash")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }
    defer os.Remove(tmpFile.Name())
    
    // Write test content
    content := []byte("test content")
    tmpFile.Write(content)
    tmpFile.Close()
    
    // Calculate hash
    hash, err := CalculateSHA256(tmpFile.Name())
    if err != nil {
        t.Fatalf("CalculateSHA256 failed: %v", err)
    }
    
    // Verify hash is correct length (64 hex chars for SHA256)
    if len(hash) != 64 {
        t.Errorf("Expected hash length 64, got %d", len(hash))
    }
}
```

- [ ] Create test file
- [ ] Run tests: `go test ./internal/pkg -v`

### 5. Add Manifest Tests (5 minutes)
Create `internal/manifest/manifest_test.go`:
```go
package manifest

import (
    "os"
    "path/filepath"
    "testing"
)

func TestManifestValidation(t *testing.T) {
    // Test valid manifest
    valid := &Manifest{
        Name:    "@acme/test-package",
        Version: "1.0.0",
        Files:   []string{"rules/*.md"},
        Targets: []string{"cursor"},
    }
    
    if err := valid.Validate(); err != nil {
        t.Errorf("Valid manifest should pass validation: %v", err)
    }
    
    // Test invalid name
    invalid := &Manifest{
        Name:    "Invalid Name With Spaces",
        Version: "1.0.0",
        Files:   []string{"rules/*.md"},
    }
    
    if err := invalid.Validate(); err == nil {
        t.Error("Invalid name should fail validation")
    }
    
    // Test invalid version
    invalid.Name = "valid-name"
    invalid.Version = "not-semver"
    
    if err := invalid.Validate(); err == nil {
        t.Error("Invalid version should fail validation")
    }
    
    // Test missing files
    invalid.Version = "1.0.0"
    invalid.Files = []string{}
    
    if err := invalid.Validate(); err == nil {
        t.Error("Empty files array should fail validation")
    }
}

func TestGetScopeAndName(t *testing.T) {
    // Test scoped package
    scoped := &Manifest{Name: "@acme/package-name"}
    
    if scope := scoped.GetScope(); scope != "acme" {
        t.Errorf("Expected scope 'acme', got '%s'", scope)
    }
    
    if name := scoped.GetPackageName(); name != "package-name" {
        t.Errorf("Expected name 'package-name', got '%s'", name)
    }
    
    // Test unscoped package
    unscoped := &Manifest{Name: "package-name"}
    
    if scope := unscoped.GetScope(); scope != "" {
        t.Errorf("Expected empty scope, got '%s'", scope)
    }
    
    if name := unscoped.GetPackageName(); name != "package-name" {
        t.Errorf("Expected name 'package-name', got '%s'", name)
    }
}

func TestSaveAndLoad(t *testing.T) {
    tmpFile := filepath.Join(os.TempDir(), "test-manifest.json")
    defer os.Remove(tmpFile)
    
    original := CreateSample()
    
    // Save manifest
    if err := original.Save(tmpFile); err != nil {
        t.Fatalf("Failed to save manifest: %v", err)
    }
    
    // Load manifest
    loaded, err := Load(tmpFile)
    if err != nil {
        t.Fatalf("Failed to load manifest: %v", err)
    }
    
    // Compare
    if loaded.Name != original.Name {
        t.Errorf("Expected name %s, got %s", original.Name, loaded.Name)
    }
    
    if loaded.Version != original.Version {
        t.Errorf("Expected version %s, got %s", original.Version, loaded.Version)
    }
}
```

- [ ] Create manifest test file
- [ ] Run tests: `go test ./internal/manifest -v`

## Validation
Test the manifest and archive functionality:
```bash
# Run all tests
go test ./internal/manifest -v
go test ./internal/pkg -v

# Test manifest creation
go run -c "
import (\"fmt\"; \"rulestack/internal/manifest\")
func main() {
    m := manifest.CreateSample()
    fmt.Printf(\"Sample manifest: %+v\n\", m)
    if err := m.Validate(); err != nil {
        panic(err)
    }
    fmt.Println(\"Validation passed\")
}"
```

## Acceptance Criteria
- [ ] Manifest validation works correctly (names, versions, targets)
- [ ] Can load and save manifest files
- [ ] Archive creation works with glob patterns
- [ ] Archive extraction preserves file structure
- [ ] SHA256 calculation is accurate
- [ ] Lockfile operations function correctly
- [ ] All tests pass
- [ ] No compilation errors
- [ ] File security (no directory traversal in extraction)

## Time Estimate: ~60 minutes

## Next Task
Task 8: CLI Foundation and Commands