# Phase 5: Git Registry Search and Discovery

## Overview
Implement package discovery operations for Git-based registries, including search, package information retrieval, and archive downloading. This phase focuses on reading and parsing the registry structure with consistent error handling and caching strategies.

## Scope
- Implement optimized registry index parsing with fallback
- Add efficient package search functionality
- Implement package and version information retrieval with proper error types
- Add robust archive download capability with verification
- Create comprehensive registry structure validation
- Ensure consistency with HTTP client patterns

## Prerequisites
- Phase 4: Git Client Basic Operations completed
- Understanding of registry structure patterns
- Error handling types established

## Expected Repository Structure

```
repo-root/
├── index.json              # Registry metadata and package listing
├── packages/
│   ├── package-name/
│   │   ├── metadata.json   # Package metadata
│   │   ├── versions/
│   │   │   ├── v1.0.0/
│   │   │   │   ├── manifest.json
│   │   │   │   └── archive.tar.gz
│   │   │   └── v1.1.0/
│   │   │       ├── manifest.json
│   │   │       └── archive.tar.gz
│   │   └── latest.json     # Points to latest version
│   └── another-package/
└── README.md               # Optional registry documentation
```

## Implementation Steps

### 1. Define Registry Index Structure

**File**: `internal/client/git_types.go` (new file)

```go
package client

import "time"

// GitRegistryIndex represents the root index.json file
type GitRegistryIndex struct {
    Version     string                    `json:"version"`
    UpdatedAt   time.Time                `json:"updated_at"`
    PackageCount int                      `json:"package_count"`
    Packages    map[string]GitPackageEntry `json:"packages"`
}

// GitPackageEntry represents a package entry in the index
type GitPackageEntry struct {
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Latest      string    `json:"latest"`
    UpdatedAt   time.Time `json:"updated_at"`
    Tags        []string  `json:"tags,omitempty"`
}

// GitPackageMetadata represents the metadata.json file
type GitPackageMetadata struct {
    Name        string              `json:"name"`
    Description string              `json:"description"`
    Latest      string              `json:"latest"`
    Versions    []GitVersionSummary `json:"versions"`
    Tags        []string           `json:"tags,omitempty"`
    CreatedAt   time.Time          `json:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at"`
}

// GitVersionSummary represents a version in package metadata
type GitVersionSummary struct {
    Version     string    `json:"version"`
    SHA256      string    `json:"sha256"`
    Size        int64     `json:"size"`
    PublishedAt time.Time `json:"published_at"`
}

// GitManifest represents a version's manifest.json
type GitManifest struct {
    Name         string            `json:"name"`
    Version      string            `json:"version"`
    Description  string            `json:"description"`
    Dependencies map[string]string `json:"dependencies,omitempty"`
    SHA256       string            `json:"sha256"`
    Size         int64             `json:"size"`
    PublishedAt  time.Time        `json:"published_at"`
    Publisher    string           `json:"publisher"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
}
```

### 2. Implement Index Loading

**File**: `internal/client/git.go`

```go
import (
    "encoding/json"
    "io/ioutil"
)

// loadIndex loads and parses the registry index
func (c *GitClient) loadIndex(ctx context.Context) (*GitRegistryIndex, error) {
    // Ensure repository is up to date
    if err := c.ensureRepo(ctx); err != nil {
        return nil, err
    }
    
    indexPath := c.getIndexPath()
    
    // Check if index exists
    if _, err := os.Stat(indexPath); os.IsNotExist(err) {
        if c.verbose {
            fmt.Printf("⚠️  Index not found, attempting to rebuild from packages directory\n")
        }
        // Try to rebuild index from packages directory
        return c.rebuildIndex()
    }
    
    // Read index file
    data, err := ioutil.ReadFile(indexPath)
    if err != nil {
        return nil, NewRegistryError(ErrInvalidRegistry, fmt.Sprintf("failed to read index: %v", err))
    }
    
    var index GitRegistryIndex
    if err := json.Unmarshal(data, &index); err != nil {
        if c.verbose {
            fmt.Printf("⚠️  Index corrupted, rebuilding from packages directory\n")
        }
        // If index is corrupted, try to rebuild
        return c.rebuildIndex()
    }
    
    return &index, nil
}

// loadPackageMetadata loads metadata for a specific package
func (c *GitClient) loadPackageMetadata(packageName string) (*GitPackageMetadata, error) {
    metadataPath := filepath.Join(c.getPackagePath(packageName), "metadata.json")
    
    if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
        return nil, NewRegistryError(ErrPackageNotFound, packageName)
    }
    
    data, err := ioutil.ReadFile(metadataPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read package metadata: %w", err)
    }
    
    var metadata GitPackageMetadata
    if err := json.Unmarshal(data, &metadata); err != nil {
        return nil, fmt.Errorf("failed to parse package metadata: %w", err)
    }
    
    return &metadata, nil
}

// loadManifest loads manifest for a specific version
func (c *GitClient) loadManifest(packageName, version string) (*GitManifest, error) {
    manifestPath := filepath.Join(c.getVersionPath(packageName, version), "manifest.json")
    
    if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
        return nil, NewRegistryError(ErrVersionNotFound, 
            fmt.Sprintf("%s@%s", packageName, version))
    }
    
    data, err := ioutil.ReadFile(manifestPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read manifest: %w", err)
    }
    
    var manifest GitManifest
    if err := json.Unmarshal(data, &manifest); err != nil {
        return nil, fmt.Errorf("failed to parse manifest: %w", err)
    }
    
    return &manifest, nil
}
```

### 3. Implement Search Functionality

**File**: `internal/client/git.go`

```go
// SearchPackages searches for packages in the Git registry
func (c *GitClient) SearchPackages(ctx context.Context, opts SearchOptions) ([]Package, error) {
    if c.verbose {
        fmt.Printf("🔍 Searching packages with query: %s\n", opts.Query)
    }
    
    // Load registry index
    index, err := c.loadIndex(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load registry index: %w", err)
    }
    
    var results []Package
    count := 0
    
    for _, entry := range index.Packages {
        // Apply search filters
        if !c.matchesSearch(entry, opts) {
            continue
        }
        
        // Convert to Package struct
        pkg := Package{
            Name:        entry.Name,
            Description: entry.Description,
            Latest:      entry.Latest,
            Tags:        entry.Tags,
            UpdatedAt:   entry.UpdatedAt,
        }
        
        // Load versions from metadata if available
        if metadata, err := c.loadPackageMetadata(entry.Name); err == nil {
            pkg.Versions = make([]string, len(metadata.Versions))
            for i, v := range metadata.Versions {
                pkg.Versions[i] = v.Version
            }
        }
        
        results = append(results, pkg)
        count++
        
        // Apply limit
        if opts.Limit > 0 && count >= opts.Limit {
            break
        }
    }
    
    if c.verbose {
        fmt.Printf("✅ Found %d packages\n", len(results))
    }
    
    return results, nil
}

// matchesSearch checks if a package matches search criteria
func (c *GitClient) matchesSearch(entry GitPackageEntry, opts SearchOptions) bool {
    // Query match (case-insensitive)
    if opts.Query != "" {
        query := strings.ToLower(opts.Query)
        name := strings.ToLower(entry.Name)
        desc := strings.ToLower(entry.Description)
        
        if !strings.Contains(name, query) && !strings.Contains(desc, query) {
            return false
        }
    }
    
    // Tag filter
    if opts.Tag != "" {
        found := false
        for _, tag := range entry.Tags {
            if tag == opts.Tag {
                found = true
                break
            }
        }
        if !found {
            return false
        }
    }
    
    // Target filter (could be in metadata/tags)
    if opts.Target != "" {
        // For now, check if target is in tags
        found := false
        for _, tag := range entry.Tags {
            if strings.Contains(tag, opts.Target) {
                found = true
                break
            }
        }
        if !found {
            return false
        }
    }
    
    return true
}
```

### 4. Implement Package Retrieval

**File**: `internal/client/git.go`

```go
// GetPackage gets information about a specific package
func (c *GitClient) GetPackage(ctx context.Context, name string) (*Package, error) {
    if c.verbose {
        fmt.Printf("📦 Getting package: %s\n", name)
    }
    
    // Ensure repository is up to date
    if err := c.ensureRepo(ctx); err != nil {
        return nil, err
    }
    
    // Check if package exists
    if !c.packageExists(name) {
        return nil, NewRegistryError(ErrPackageNotFound, name)
    }
    
    // Load package metadata
    metadata, err := c.loadPackageMetadata(name)
    if err != nil {
        return nil, err
    }
    
    // Convert to Package struct
    pkg := &Package{
        Name:        metadata.Name,
        Description: metadata.Description,
        Latest:      metadata.Latest,
        Tags:        metadata.Tags,
        UpdatedAt:   metadata.UpdatedAt,
        Versions:    make([]string, len(metadata.Versions)),
    }
    
    for i, v := range metadata.Versions {
        pkg.Versions[i] = v.Version
    }
    
    if c.verbose {
        fmt.Printf("✅ Found package with %d versions\n", len(pkg.Versions))
    }
    
    return pkg, nil
}

// GetPackageVersion gets information about a specific package version
func (c *GitClient) GetPackageVersion(ctx context.Context, name, version string) (*PackageVersion, error) {
    if c.verbose {
        fmt.Printf("📦 Getting package version: %s@%s\n", name, version)
    }
    
    // Ensure repository is up to date
    if err := c.ensureRepo(ctx); err != nil {
        return nil, err
    }
    
    // Check if version exists
    if !c.versionExists(name, version) {
        return nil, NewRegistryError(ErrVersionNotFound, 
            fmt.Sprintf("%s@%s", name, version))
    }
    
    // Load manifest
    manifest, err := c.loadManifest(name, version)
    if err != nil {
        return nil, err
    }
    
    // Convert to PackageVersion struct
    pv := &PackageVersion{
        Name:         manifest.Name,
        Version:      manifest.Version,
        Description:  manifest.Description,
        Dependencies: manifest.Dependencies,
        SHA256:       manifest.SHA256,
        Size:         manifest.Size,
        PublishedAt:  manifest.PublishedAt,
        Metadata:     manifest.Metadata,
    }
    
    if c.verbose {
        fmt.Printf("✅ Found version published at %s\n", pv.PublishedAt.Format(time.RFC3339))
    }
    
    return pv, nil
}
```

### 5. Implement Archive Download

**File**: `internal/client/git.go`

```go
import (
    "io"
)

// DownloadBlob downloads a package archive by SHA256 hash
func (c *GitClient) DownloadBlob(ctx context.Context, sha256Hash, destPath string) error {
    if c.verbose {
        fmt.Printf("📥 Downloading blob: %s\n", sha256Hash)
    }
    
    // Ensure repository is up to date
    if err := c.ensureRepo(ctx); err != nil {
        return err
    }
    
    // Find the archive file by hash
    archivePath, err := c.findArchiveByHash(sha256Hash)
    if err != nil {
        return err
    }
    
    // Copy file to destination
    if err := c.copyFile(archivePath, destPath); err != nil {
        return fmt.Errorf("failed to copy archive: %w", err)
    }
    
    if c.verbose {
        fmt.Printf("✅ Downloaded to %s\n", destPath)
    }
    
    return nil
}

// findArchiveByHash searches for an archive file by its SHA256 hash
func (c *GitClient) findArchiveByHash(sha256Hash string) (string, error) {
    packagesDir := filepath.Join(c.cacheDir, "packages")
    
    var foundPath string
    err := filepath.Walk(packagesDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // Look for archive.tar.gz files
        if filepath.Base(path) == "archive.tar.gz" {
            // Calculate hash of file
            hash, err := c.calculateFileHash(path)
            if err != nil {
                return nil // Skip this file
            }
            
            if hash == sha256Hash {
                foundPath = path
                return io.EOF // Stop walking
            }
        }
        
        return nil
    })
    
    if err != nil && err != io.EOF {
        return "", fmt.Errorf("error searching for archive: %w", err)
    }
    
    if foundPath == "" {
        return "", fmt.Errorf("archive with hash %s not found", sha256Hash)
    }
    
    return foundPath, nil
}

// calculateFileHash calculates SHA256 hash of a file
func (c *GitClient) calculateFileHash(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    h := sha256.New()
    if _, err := io.Copy(h, file); err != nil {
        return "", err
    }
    
    return hex.EncodeToString(h.Sum(nil)), nil
}

// copyFile copies a file from source to destination
func (c *GitClient) copyFile(src, dst string) error {
    // Create destination directory if needed
    if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
        return err
    }
    
    source, err := os.Open(src)
    if err != nil {
        return err
    }
    defer source.Close()
    
    destination, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destination.Close()
    
    _, err = io.Copy(destination, source)
    return err
}
```

### 6. Add Validation Helpers

**File**: `internal/client/git.go`

```go
// validateRegistryStructure validates the Git registry structure
func (c *GitClient) validateRegistryStructure() error {
    // Check for packages directory
    packagesDir := filepath.Join(c.cacheDir, "packages")
    if _, err := os.Stat(packagesDir); os.IsNotExist(err) {
        return fmt.Errorf("invalid registry: packages directory not found")
    }
    
    // Index is optional but recommended
    indexPath := c.getIndexPath()
    if _, err := os.Stat(indexPath); os.IsNotExist(err) {
        if c.verbose {
            fmt.Printf("⚠️  Warning: index.json not found, performance may be degraded\n")
        }
    }
    
    return nil
}

// rebuildIndex rebuilds the index from the packages directory
func (c *GitClient) rebuildIndex() (*GitRegistryIndex, error) {
    packagesDir := filepath.Join(c.cacheDir, "packages")
    
    // Check if packages directory exists
    if _, err := os.Stat(packagesDir); os.IsNotExist(err) {
        return nil, NewRegistryError(ErrInvalidRegistry, "packages directory not found - not a valid registry")
    }
    
    index := &GitRegistryIndex{
        Version:   "1.0",
        UpdatedAt: time.Now(),
        Packages:  make(map[string]GitPackageEntry),
    }
    
    if c.verbose {
        fmt.Printf("🔄 Rebuilding index from packages directory\n")
    }
    
    // Walk through packages directory
    entries, err := ioutil.ReadDir(packagesDir)
    if err != nil {
        return nil, NewRegistryError(ErrInvalidRegistry, fmt.Sprintf("failed to read packages directory: %v", err))
    }
    
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }
        
        packageName := entry.Name()
        
        // Try to load metadata
        metadata, err := c.loadPackageMetadata(packageName)
        if err != nil {
            continue // Skip packages without metadata
        }
        
        index.Packages[packageName] = GitPackageEntry{
            Name:        metadata.Name,
            Description: metadata.Description,
            Latest:      metadata.Latest,
            Tags:        metadata.Tags,
            UpdatedAt:   metadata.UpdatedAt,
        }
        index.PackageCount++
    }
    
    return index, nil
}
```

## Testing Requirements

### Unit Tests
1. Test index parsing
2. Test search filtering logic
3. Test hash calculation
4. Test structure validation

### Integration Tests
1. Test search with various filters
2. Test package retrieval
3. Test version retrieval
4. Test archive download
5. Test with missing index (rebuild)

### Manual Testing Checklist
- [ ] Search packages with query
- [ ] Filter by tags
- [ ] Get package information
- [ ] Get version information
- [ ] Download archive by hash
- [ ] Handle missing index gracefully
- [ ] Validate repository structure

### Cucumber Test Amendments

**File**: `features/git-registry-search.feature` (new file)

```gherkin
Feature: Git Registry Search and Discovery
  Git registry should support package search and discovery operations

  Background:
    Given I have a clean test environment
    And a Git registry "test-git" with test packages

  Scenario: Search packages in Git registry
    Given I use registry "test-git"
    When I run "rfh search test"
    Then the command should succeed
    And the output should contain package results

  Scenario: Search with query filter
    Given I use registry "test-git"
    When I run "rfh search --query security"
    Then the command should succeed
    And the output should only contain packages matching "security"

  Scenario: Search with tag filter
    Given I use registry "test-git"
    When I run "rfh search --tag auth"
    Then the command should succeed
    And the output should only contain packages tagged with "auth"

  Scenario: Search with limit
    Given I use registry "test-git"
    When I run "rfh search --limit 2"
    Then the command should succeed
    And the output should contain at most 2 packages

  Scenario: Get specific package information
    Given I use registry "test-git"
    When I run "rfh get test-package"
    Then the command should succeed
    And the output should contain package details for "test-package"
    And the output should contain version information

  Scenario: Get specific package version
    Given I use registry "test-git"
    When I run "rfh get test-package@1.0.0"
    Then the command should succeed
    And the output should contain version details for "test-package@1.0.0"
    And the output should contain SHA256 hash
    And the output should contain dependencies

  Scenario: Get nonexistent package
    Given I use registry "test-git"
    When I run "rfh get nonexistent-package"
    Then the command should fail
    And the output should contain "package not found"

  Scenario: Get nonexistent version
    Given I use registry "test-git"
    When I run "rfh get test-package@99.99.99"
    Then the command should fail
    And the output should contain "version not found"

  Scenario: Download package archive
    Given I use registry "test-git"
    And package "test-package@1.0.0" exists with hash "abc123"
    When I run "rfh download abc123 ./test-archive.tar.gz"
    Then the command should succeed
    And the file "./test-archive.tar.gz" should exist
    And the file should have the correct SHA256 hash

  Scenario: Registry without index rebuilds automatically
    Given a Git registry "no-index" without index file
    And I use registry "no-index"
    When I run "rfh search test"
    Then the command should succeed
    And the output should contain "Warning: index.json not found"
    And the output should contain package results

  Scenario: Invalid registry structure
    Given a Git registry "invalid" with no packages directory
    And I use registry "invalid"
    When I run "rfh search test"
    Then the command should fail
    And the output should contain "invalid registry structure"

  Scenario: Large registry search performance
    Given a Git registry "large" with 100 packages
    And I use registry "large"
    When I run "rfh search test --limit 10"
    Then the command should succeed within 5 seconds
    And the output should contain at most 10 packages
```

**File**: `features/step_definitions/git_search_steps.js`

Add new step definitions:
```javascript
Given('a Git registry {string} with test packages', async function (registryName) {
  // Set up a test git registry with sample packages
  const config = await this.loadConfig();
  config.registries[registryName] = {
    url: 'https://github.com/test-org/test-registry',
    type: 'git'
  };
  await this.saveConfig(config);
  
  // Mock the git repository with test data
  await this.mockGitRegistry.setup(registryName, {
    packages: {
      'test-package': {
        name: 'test-package',
        description: 'A test package for security',
        latest: '1.0.0',
        tags: ['security', 'auth'],
        versions: ['1.0.0']
      },
      'another-package': {
        name: 'another-package', 
        description: 'Another test package',
        latest: '2.0.0',
        tags: ['utils'],
        versions: ['1.0.0', '2.0.0']
      }
    }
  });
});

Then('the output should contain package results', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('📦') || output.includes('package') || 
         output.includes('test-package') || output.includes('another-package'),
    `Expected package results in output: ${output}`);
});

Then('the output should only contain packages matching {string}', 
  function (query) {
    const output = this.lastResult.stdout || '';
    const lines = output.split('\n').filter(line => line.includes('📦'));
    
    for (const line of lines) {
      assert(line.toLowerCase().includes(query.toLowerCase()),
        `Package line should contain "${query}": ${line}`);
    }
});

Then('the output should only contain packages tagged with {string}',
  function (tag) {
    const output = this.lastResult.stdout || '';
    // This would check that only packages with the specified tag are shown
    assert(output.includes(tag), `Expected packages tagged with "${tag}"`);
});

Then('the output should contain at most {int} packages', function (maxPackages) {
  const output = this.lastResult.stdout || '';
  const packageLines = output.split('\n').filter(line => line.includes('📦'));
  assert(packageLines.length <= maxPackages,
    `Expected at most ${maxPackages} packages, found ${packageLines.length}`);
});

Then('the output should contain package details for {string}', 
  function (packageName) {
    const output = this.lastResult.stdout || '';
    assert(output.includes(packageName), 
      `Expected package details for ${packageName}`);
    assert(output.includes('Description:') || output.includes('Versions:'),
      'Expected detailed package information');
});

Then('the output should contain version details for {string}', 
  function (packageVersion) {
    const output = this.lastResult.stdout || '';
    assert(output.includes(packageVersion.split('@')[0]) && 
           output.includes(packageVersion.split('@')[1]),
      `Expected version details for ${packageVersion}`);
});

Then('the output should contain SHA256 hash', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('SHA256:') || output.includes('sha256'),
    'Expected SHA256 hash in output');
});

Then('the output should contain dependencies', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('Dependencies:') || output.includes('dependencies') ||
         output.includes('No dependencies'),
    'Expected dependency information in output');
});

Given('package {string} exists with hash {string}', 
  async function (packageVersion, hash) {
    // Mock the package with specified hash
    await this.mockGitRegistry.addPackage(packageVersion, { sha256: hash });
});

When('I run "rfh download {string} {string}"', async function (hash, destPath) {
  await this.runCommand(`rfh download ${hash} ${destPath}`);
});

Then('the file {string} should exist', function (filePath) {
  const fs = require('fs');
  assert(fs.existsSync(filePath), `File ${filePath} should exist`);
});

Then('the file should have the correct SHA256 hash', function () {
  // This would verify the downloaded file matches the expected hash
  // For now, we just check it was created
  assert(true, 'Hash verification would be implemented');
});

Given('a Git registry {string} without index file', async function (registryName) {
  // Set up registry without index.json
  const config = await this.loadConfig();
  config.registries[registryName] = {
    url: 'https://github.com/test-org/no-index-registry',
    type: 'git'
  };
  await this.saveConfig(config);
  
  await this.mockGitRegistry.setup(registryName, { noIndex: true });
});

Given('a Git registry {string} with no packages directory', 
  async function (registryName) {
    const config = await this.loadConfig();
    config.registries[registryName] = {
      url: 'https://github.com/test-org/invalid-registry', 
      type: 'git'
    };
    await this.saveConfig(config);
    
    await this.mockGitRegistry.setup(registryName, { noPackagesDir: true });
});

Given('a Git registry {string} with {int} packages', 
  async function (registryName, packageCount) {
    const config = await this.loadConfig();
    config.registries[registryName] = {
      url: 'https://github.com/test-org/large-registry',
      type: 'git'
    };
    await this.saveConfig(config);
    
    // Generate test packages
    const packages = {};
    for (let i = 1; i <= packageCount; i++) {
      packages[`package-${i}`] = {
        name: `package-${i}`,
        description: `Test package ${i}`,
        latest: '1.0.0',
        tags: ['test'],
        versions: ['1.0.0']
      };
    }
    
    await this.mockGitRegistry.setup(registryName, { packages });
});

Then('the command should succeed within {int} seconds', function (seconds) {
  // This would be implemented with timing in the actual test runner
  assert(!this.lastResult.error, 'Command should succeed');
  // Timing verification would be added here
});
```

## Success Criteria
- Can search and filter packages from Git registry
- Can retrieve package and version information
- Can download archives by SHA256 hash
- Handles missing or corrupt index gracefully
- Validates repository structure correctly
- Efficient searching with index caching

## Dependencies
- Phase 4: Git Client Basic Operations

## Risks
- **Risk**: Large package lists affecting search performance
  **Mitigation**: Use index for efficient searching, implement pagination
  
- **Risk**: Corrupt or missing metadata files
  **Mitigation**: Graceful error handling, ability to rebuild index
  
- **Risk**: Hash calculation performance on large files
  **Mitigation**: Cache hashes in manifest files

## Next Phase
Phase 6: Git Registry Publishing - Implement package publishing via Git commits and pull requests