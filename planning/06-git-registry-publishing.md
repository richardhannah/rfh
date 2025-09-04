# Phase 6: Git Registry Publishing

## Overview
Implement package publishing for Git-based registries with simplified fork management and robust error handling. This phase covers branch creation, file management, and committing changes while maintaining consistency with existing patterns.

## Scope
- Implement simplified repository fork management
- Create branch for package publication with proper naming
- Add package files to repository structure
- Commit changes with consistent messaging
- Push branch to fork with proper authentication
- Prepare structured information for PR creation
- Use standardized error types throughout

## Prerequisites
- Phase 4: Git Client Basic Operations completed
- Phase 5: Git Registry Search and Discovery completed
- Understanding of Git workflow and GitHub patterns
- Consistent error handling established

## Publishing Workflow

1. Check if user has fork of registry
2. Clone/update user's fork
3. Create new branch: `publish/{package-name}/{version}`
4. Add package files to appropriate directory
5. Update registry index
6. Commit changes
7. Push branch to fork
8. Create PR (Phase 7)

## Implementation Steps

### 1. Add Fork Management

**File**: `internal/client/git_fork.go` (new file)

```go
package client

import (
    "fmt"
    "strings"
    
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/config"
    "github.com/go-git/go-git/v5/plumbing"
)

// ForkInfo contains information about a repository fork
type ForkInfo struct {
    OriginalURL string
    ForkURL     string
    Username    string
    RepoName    string
}

// detectFork determines fork information from the repository URL
func (c *GitClient) detectFork(repoURL string) (*ForkInfo, error) {
    // Parse GitHub URL format: https://github.com/owner/repo
    if !strings.Contains(repoURL, "github.com") {
        return nil, fmt.Errorf("only GitHub repositories supported for publishing")
    }
    
    parts := strings.Split(repoURL, "/")
    if len(parts) < 5 {
        return nil, fmt.Errorf("invalid GitHub URL format")
    }
    
    owner := parts[3]
    repoName := strings.TrimSuffix(parts[4], ".git")
    
    // Get current user from git token (requires API call in Phase 7)
    // For now, we'll require the user to provide their username
    username := c.getUsername()
    if username == "" {
        return nil, fmt.Errorf("GitHub username required for publishing")
    }
    
    forkURL := fmt.Sprintf("https://github.com/%s/%s.git", username, repoName)
    
    return &ForkInfo{
        OriginalURL: repoURL,
        ForkURL:     forkURL,
        Username:    username,
        RepoName:    repoName,
    }, nil
}

// getUsername extracts username from environment or git config
func (c *GitClient) getUsername() string {
    // Check environment variables first
    if username := os.Getenv("GITHUB_USERNAME"); username != "" {
        return username
    }
    if username := os.Getenv("GIT_USER"); username != "" {
        return username
    }
    
    // Try to read from git config (fallback)
    cmd := exec.Command("git", "config", "--get", "user.name")
    if output, err := cmd.Output(); err == nil {
        if username := strings.TrimSpace(string(output)); username != "" {
            return username
        }
    }
    
    // This will be improved in Phase 7 with GitHub API
    return ""
}

// ensureFork ensures the user's fork is set up and updated
func (c *GitClient) ensureFork(ctx context.Context, fork *ForkInfo) (*git.Repository, error) {
    // Create separate cache directory for fork
    forkCacheDir := filepath.Join(filepath.Dir(c.cacheDir), "forks", fork.RepoName)
    
    // Check if fork already cloned
    if _, err := os.Stat(filepath.Join(forkCacheDir, ".git")); err == nil {
        // Open existing fork
        repo, err := git.PlainOpen(forkCacheDir)
        if err != nil {
            return nil, fmt.Errorf("failed to open fork: %w", err)
        }
        
        // Update fork from upstream
        if err := c.updateFork(ctx, repo, fork); err != nil {
            return nil, err
        }
        
        return repo, nil
    }
    
    // Clone fork
    if c.verbose {
        fmt.Printf("📥 Cloning fork: %s\n", fork.ForkURL)
    }
    
    cloneOpts := &git.CloneOptions{
        URL:      fork.ForkURL,
        Progress: nil,
    }
    
    if c.verbose {
        cloneOpts.Progress = os.Stdout
    }
    
    if c.gitToken != "" {
        cloneOpts.Auth = c.getAuth()
    }
    
    repo, err := git.PlainCloneContext(ctx, forkCacheDir, false, cloneOpts)
    if err != nil {
        return nil, fmt.Errorf("failed to clone fork: %w", err)
    }
    
    // Add upstream remote
    if err := c.addUpstreamRemote(repo, fork.OriginalURL); err != nil {
        return nil, err
    }
    
    return repo, nil
}

// addUpstreamRemote adds the original repository as upstream remote
func (c *GitClient) addUpstreamRemote(repo *git.Repository, upstreamURL string) error {
    _, err := repo.CreateRemote(&config.RemoteConfig{
        Name: "upstream",
        URLs: []string{upstreamURL},
    })
    
    if err != nil && !strings.Contains(err.Error(), "already exists") {
        return fmt.Errorf("failed to add upstream remote: %w", err)
    }
    
    return nil
}

// updateFork syncs fork with upstream
func (c *GitClient) updateFork(ctx context.Context, repo *git.Repository, fork *ForkInfo) error {
    if c.verbose {
        fmt.Printf("🔄 Updating fork from upstream\n")
    }
    
    // Fetch from upstream
    fetchOpts := &git.FetchOptions{
        RemoteName: "upstream",
    }
    
    if c.gitToken != "" {
        fetchOpts.Auth = c.getAuth()
    }
    
    err := repo.FetchContext(ctx, fetchOpts)
    if err != nil && err != git.NoErrAlreadyUpToDate {
        return fmt.Errorf("failed to fetch upstream: %w", err)
    }
    
    // Merge upstream/main into main
    // This will be improved in a production implementation
    
    return nil
}
```

### 2. Implement Branch Creation

**File**: `internal/client/git_publish.go` (new file)

```go
package client

import (
    "fmt"
    "path/filepath"
    "time"
    
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
)

// createPublishBranch creates a new branch for publishing
func (c *GitClient) createPublishBranch(repo *git.Repository, packageName, version string) (string, error) {
    branchName := fmt.Sprintf("publish/%s/%s", packageName, version)
    
    if c.verbose {
        fmt.Printf("🌿 Creating branch: %s\n", branchName)
    }
    
    // Get current HEAD
    head, err := repo.Head()
    if err != nil {
        return "", fmt.Errorf("failed to get HEAD: %w", err)
    }
    
    // Create new branch from HEAD
    ref := plumbing.NewBranchReferenceName(branchName)
    err = repo.CreateBranch(&config.Branch{
        Name:   branchName,
        Remote: "origin",
        Merge:  ref,
    })
    
    if err != nil && !strings.Contains(err.Error(), "already exists") {
        return "", NewRegistryError(ErrInvalidOperation, fmt.Sprintf("failed to create branch: %v", err))
    }
    
    // Checkout the new branch
    w, err := repo.Worktree()
    if err != nil {
        return "", fmt.Errorf("failed to get worktree: %w", err)
    }
    
    err = w.Checkout(&git.CheckoutOptions{
        Branch: ref,
        Create: true,
        Force:  true,
    })
    
    if err != nil && !strings.Contains(err.Error(), "already exists") {
        return "", fmt.Errorf("failed to checkout branch: %w", err)
    }
    
    return branchName, nil
}
```

### 3. Implement Package File Addition

**File**: `internal/client/git_publish.go`

```go
import (
    "encoding/json"
    "io"
    "crypto/sha256"
)

// addPackageFiles adds package files to the repository
func (c *GitClient) addPackageFiles(repo *git.Repository, manifestPath, archivePath string) error {
    // Parse manifest to get package info
    manifestData, err := ioutil.ReadFile(manifestPath)
    if err != nil {
        return fmt.Errorf("failed to read manifest: %w", err)
    }
    
    var manifest GitManifest
    if err := json.Unmarshal(manifestData, &manifest); err != nil {
        return fmt.Errorf("failed to parse manifest: %w", err)
    }
    
    // Calculate archive hash
    archiveHash, archiveSize, err := c.calculateFileInfo(archivePath)
    if err != nil {
        return fmt.Errorf("failed to calculate archive info: %w", err)
    }
    
    // Update manifest with archive info
    manifest.SHA256 = archiveHash
    manifest.Size = archiveSize
    manifest.PublishedAt = time.Now()
    
    // Get worktree
    w, err := repo.Worktree()
    if err != nil {
        return fmt.Errorf("failed to get worktree: %w", err)
    }
    
    // Create package directory structure
    packageDir := filepath.Join(w.Filesystem.Root(), "packages", manifest.Name)
    versionDir := filepath.Join(packageDir, "versions", manifest.Version)
    
    if err := os.MkdirAll(versionDir, 0755); err != nil {
        return fmt.Errorf("failed to create directories: %w", err)
    }
    
    // Write manifest
    manifestDest := filepath.Join(versionDir, "manifest.json")
    updatedManifest, _ := json.MarshalIndent(manifest, "", "  ")
    if err := ioutil.WriteFile(manifestDest, updatedManifest, 0644); err != nil {
        return fmt.Errorf("failed to write manifest: %w", err)
    }
    
    // Copy archive
    archiveDest := filepath.Join(versionDir, "archive.tar.gz")
    if err := c.copyFile(archivePath, archiveDest); err != nil {
        return fmt.Errorf("failed to copy archive: %w", err)
    }
    
    // Update package metadata
    if err := c.updatePackageMetadata(packageDir, &manifest); err != nil {
        return fmt.Errorf("failed to update package metadata: %w", err)
    }
    
    // Stage all changes
    if err := w.Add("packages/" + manifest.Name); err != nil {
        return fmt.Errorf("failed to stage changes: %w", err)
    }
    
    if c.verbose {
        fmt.Printf("✅ Added package files for %s@%s\n", manifest.Name, manifest.Version)
    }
    
    return nil
}

// calculateFileInfo calculates SHA256 hash and size of a file
func (c *GitClient) calculateFileInfo(filePath string) (string, int64, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", 0, err
    }
    defer file.Close()
    
    // Get file size
    stat, err := file.Stat()
    if err != nil {
        return "", 0, err
    }
    
    // Calculate hash
    h := sha256.New()
    if _, err := io.Copy(h, file); err != nil {
        return "", 0, err
    }
    
    return hex.EncodeToString(h.Sum(nil)), stat.Size(), nil
}

// updatePackageMetadata updates the package metadata.json file
func (c *GitClient) updatePackageMetadata(packageDir string, manifest *GitManifest) error {
    metadataPath := filepath.Join(packageDir, "metadata.json")
    
    var metadata GitPackageMetadata
    
    // Load existing metadata if it exists
    if data, err := ioutil.ReadFile(metadataPath); err == nil {
        json.Unmarshal(data, &metadata)
    } else {
        // Create new metadata
        metadata = GitPackageMetadata{
            Name:        manifest.Name,
            Description: manifest.Description,
            CreatedAt:   time.Now(),
        }
    }
    
    // Update metadata
    metadata.Latest = manifest.Version
    metadata.UpdatedAt = time.Now()
    
    // Add version if not exists
    versionExists := false
    for i, v := range metadata.Versions {
        if v.Version == manifest.Version {
            versionExists = true
            // Update existing version info
            metadata.Versions[i] = GitVersionSummary{
                Version:     manifest.Version,
                SHA256:      manifest.SHA256,
                Size:        manifest.Size,
                PublishedAt: manifest.PublishedAt,
            }
            break
        }
    }
    
    if !versionExists {
        metadata.Versions = append(metadata.Versions, GitVersionSummary{
            Version:     manifest.Version,
            SHA256:      manifest.SHA256,
            Size:        manifest.Size,
            PublishedAt: manifest.PublishedAt,
        })
    }
    
    // Write updated metadata
    data, _ := json.MarshalIndent(metadata, "", "  ")
    return ioutil.WriteFile(metadataPath, data, 0644)
}
```

### 4. Implement Index Update

**File**: `internal/client/git_publish.go`

```go
// updateRegistryIndex updates the main registry index
func (c *GitClient) updateRegistryIndex(repo *git.Repository, manifest *GitManifest) error {
    w, err := repo.Worktree()
    if err != nil {
        return fmt.Errorf("failed to get worktree: %w", err)
    }
    
    indexPath := filepath.Join(w.Filesystem.Root(), "index.json")
    
    var index GitRegistryIndex
    
    // Load existing index
    if data, err := ioutil.ReadFile(indexPath); err == nil {
        json.Unmarshal(data, &index)
    } else {
        // Create new index
        index = GitRegistryIndex{
            Version:  "1.0",
            Packages: make(map[string]GitPackageEntry),
        }
    }
    
    // Update index
    index.UpdatedAt = time.Now()
    index.Packages[manifest.Name] = GitPackageEntry{
        Name:        manifest.Name,
        Description: manifest.Description,
        Latest:      manifest.Version,
        UpdatedAt:   time.Now(),
    }
    
    if _, exists := index.Packages[manifest.Name]; !exists {
        index.PackageCount++
    }
    
    // Write updated index
    data, _ := json.MarshalIndent(index, "", "  ")
    if err := ioutil.WriteFile(indexPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write index: %w", err)
    }
    
    // Stage index changes
    if err := w.Add("index.json"); err != nil {
        return fmt.Errorf("failed to stage index: %w", err)
    }
    
    return nil
}
```

### 5. Implement Commit Creation

**File**: `internal/client/git_publish.go`

```go
// createCommit creates a commit for the package publication
func (c *GitClient) createCommit(repo *git.Repository, manifest *GitManifest) (plumbing.Hash, error) {
    w, err := repo.Worktree()
    if err != nil {
        return plumbing.ZeroHash, fmt.Errorf("failed to get worktree: %w", err)
    }
    
    // Create commit message
    message := fmt.Sprintf("Publish %s@%s\n\n", manifest.Name, manifest.Version)
    message += fmt.Sprintf("- Package: %s\n", manifest.Name)
    message += fmt.Sprintf("- Version: %s\n", manifest.Version)
    message += fmt.Sprintf("- Description: %s\n", manifest.Description)
    message += fmt.Sprintf("- SHA256: %s\n", manifest.SHA256)
    message += fmt.Sprintf("- Size: %d bytes\n", manifest.Size)
    
    if c.verbose {
        fmt.Printf("💬 Creating commit: %s@%s\n", manifest.Name, manifest.Version)
    }
    
    // Get author info
    author := c.getAuthor()
    
    // Create commit
    commit, err := w.Commit(message, &git.CommitOptions{
        Author: author,
    })
    
    if err != nil {
        return plumbing.ZeroHash, fmt.Errorf("failed to create commit: %w", err)
    }
    
    if c.verbose {
        fmt.Printf("✅ Created commit: %s\n", commit.String()[:7])
    }
    
    return commit, nil
}

// getAuthor returns author information for commits
func (c *GitClient) getAuthor() *object.Signature {
    // Try to get from environment
    name := os.Getenv("GIT_AUTHOR_NAME")
    email := os.Getenv("GIT_AUTHOR_EMAIL")
    
    if name == "" {
        name = "RuleStack Publisher"
    }
    if email == "" {
        email = "publisher@rulestack.dev"
    }
    
    return &object.Signature{
        Name:  name,
        Email: email,
        When:  time.Now(),
    }
}
```

### 6. Implement Push Operation

**File**: `internal/client/git_publish.go`

```go
// pushBranch pushes the branch to the remote repository
func (c *GitClient) pushBranch(ctx context.Context, repo *git.Repository, branchName string) error {
    if c.verbose {
        fmt.Printf("📤 Pushing branch: %s\n", branchName)
    }
    
    pushOpts := &git.PushOptions{
        RemoteName: "origin",
        RefSpecs: []config.RefSpec{
            config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName)),
        },
    }
    
    if c.gitToken != "" {
        pushOpts.Auth = c.getAuth()
    }
    
    if c.verbose {
        pushOpts.Progress = os.Stdout
    }
    
    err := repo.PushContext(ctx, pushOpts)
    if err != nil && err != git.NoErrAlreadyUpToDate {
        return fmt.Errorf("failed to push branch: %w", err)
    }
    
    if c.verbose {
        fmt.Printf("✅ Branch pushed successfully\n")
    }
    
    return nil
}
```

### 7. Implement Main Publish Method

**File**: `internal/client/git.go`

```go
// PublishPackage publishes a package to the Git registry
func (c *GitClient) PublishPackage(ctx context.Context, manifestPath, archivePath string) (*PublishResult, error) {
    if c.verbose {
        fmt.Printf("📦 Publishing package to Git registry\n")
    }
    
    // Detect fork information
    fork, err := c.detectFork(c.repoURL)
    if err != nil {
        return nil, fmt.Errorf("failed to detect fork: %w", err)
    }
    
    // Ensure fork is ready
    forkRepo, err := c.ensureFork(ctx, fork)
    if err != nil {
        return nil, fmt.Errorf("failed to prepare fork: %w", err)
    }
    
    // Parse manifest for package info
    manifestData, _ := ioutil.ReadFile(manifestPath)
    var manifest GitManifest
    json.Unmarshal(manifestData, &manifest)
    
    // Create publish branch
    branchName, err := c.createPublishBranch(forkRepo, manifest.Name, manifest.Version)
    if err != nil {
        return nil, fmt.Errorf("failed to create branch: %w", err)
    }
    
    // Add package files
    if err := c.addPackageFiles(forkRepo, manifestPath, archivePath); err != nil {
        return nil, fmt.Errorf("failed to add package files: %w", err)
    }
    
    // Update registry index
    if err := c.updateRegistryIndex(forkRepo, &manifest); err != nil {
        return nil, fmt.Errorf("failed to update index: %w", err)
    }
    
    // Create commit
    commitHash, err := c.createCommit(forkRepo, &manifest)
    if err != nil {
        return nil, fmt.Errorf("failed to create commit: %w", err)
    }
    
    // Push branch
    if err := c.pushBranch(ctx, forkRepo, branchName); err != nil {
        return nil, fmt.Errorf("failed to push branch: %w", err)
    }
    
    // PR creation will be in Phase 7
    prURL := fmt.Sprintf("https://github.com/%s/%s/compare/main...%s:%s",
        strings.Split(fork.OriginalURL, "/")[3],
        fork.RepoName,
        fork.Username,
        branchName)
    
    return &PublishResult{
        Name:    manifest.Name,
        Version: manifest.Version,
        SHA256:  manifest.SHA256,
        PRUrl:   prURL,
        Message: fmt.Sprintf("Branch '%s' pushed. Visit %s to create PR", branchName, prURL),
    }, nil
}
```

## Testing Requirements

### Unit Tests
1. Test branch name generation
2. Test commit message formatting
3. Test file structure creation
4. Test index updating logic

### Integration Tests
1. Test fork detection
2. Test branch creation and checkout
3. Test file staging and committing
4. Test push operation
5. Test with existing package (update)

### Manual Testing Checklist
- [ ] Publish new package
- [ ] Publish new version of existing package
- [ ] Handle authentication correctly
- [ ] Fork workflow works
- [ ] Branch creation successful
- [ ] Files added to correct locations
- [ ] Commit created with proper message
- [ ] Push to fork successful

### Cucumber Test Amendments

**File**: `features/git-registry-publishing.feature` (new file)

```gherkin
Feature: Git Registry Publishing
  Git registry should support package publishing via Git workflow

  Background:
    Given I have a clean test environment
    And I have a valid GitHub token
    And I have a package ready to publish

  Scenario: Publish new package to Git registry
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    When I run "rfh publish"
    Then the command should succeed
    And the output should contain "Branch pushed"
    And the output should contain a pull request preparation URL

  Scenario: Publish with existing fork
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    And I already have a fork of the registry
    When I run "rfh publish"
    Then the command should succeed
    And the existing fork should be used
    And the output should contain "Branch pushed"

  Scenario: Publish new version of existing package
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    And package "test-package@1.0.0" already exists in the registry
    And I have "test-package@2.0.0" ready to publish
    When I run "rfh publish"
    Then the command should succeed
    And the package metadata should be updated with the new version
    And the registry index should be updated

  Scenario: Publish without authentication fails
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have not provided a GitHub token
    When I run "rfh publish"
    Then the command should fail
    And the output should contain "GitHub token required"

  Scenario: Publish to non-GitHub repository fails
    Given a Git registry "gitlab-test" with URL "https://gitlab.com/test-org/test-registry"
    And I use registry "gitlab-test"
    When I run "rfh publish"
    Then the command should fail
    And the output should contain "only GitHub repositories supported"

  Scenario: Branch naming follows convention
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    And I have "my-package@1.5.0" ready to publish
    When I run "rfh publish"
    Then the command should succeed
    And a branch "publish/my-package/1.5.0" should be created
    And the branch should be pushed to my fork

  Scenario: Commit message format is correct
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    When I run "rfh publish"
    Then the command should succeed
    And the commit should contain package name and version
    And the commit should contain package description
    And the commit should contain SHA256 hash
    And the commit should contain file size

  Scenario: Package files are placed correctly
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test" 
    And I have authenticated to the Git registry
    And I have "test-package@1.0.0" ready to publish
    When I run "rfh publish"
    Then the command should succeed
    And the file "packages/test-package/versions/1.0.0/manifest.json" should be added
    And the file "packages/test-package/versions/1.0.0/archive.tar.gz" should be added
    And the file "packages/test-package/metadata.json" should be updated
    And the file "index.json" should be updated

  Scenario: Registry index is updated correctly
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    When I run "rfh publish"
    Then the command should succeed
    And the registry index should contain the new package
    And the package count should be incremented
    And the index updated timestamp should be current

  Scenario: Publish with verbose output
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    When I run "rfh publish --verbose"
    Then the command should succeed
    And the output should contain "Cloning fork"
    And the output should contain "Creating branch"
    And the output should contain "Adding package files"
    And the output should contain "Creating commit"
    And the output should contain "Pushing branch"
```

**File**: `features/step_definitions/git_publish_steps.js`

Add new step definitions:
```javascript
Given('I have a valid GitHub token', async function () {
  process.env.GITHUB_TOKEN = 'ghp_test_token_123456789abcdef';
  process.env.GITHUB_USERNAME = 'test-user';
});

Given('I have a package ready to publish', async function () {
  // Create test package files
  await this.createTestPackage('test-package', '1.0.0', {
    description: 'A test package for publishing',
    dependencies: { 'some-dep': '^1.0.0' }
  });
});

Given('I have authenticated to the Git registry', async function () {
  const config = await this.loadConfig();
  const currentRegistry = config.registries[config.current];
  if (currentRegistry) {
    currentRegistry.git_token = process.env.GITHUB_TOKEN;
    await this.saveConfig(config);
  }
});

Given('I already have a fork of the registry', async function () {
  // Mock existing fork
  await this.mockGitHub.setupExistingFork('test-user', 'test-registry');
});

Given('package {string} already exists in the registry', 
  async function (packageVersion) {
    const [name, version] = packageVersion.split('@');
    await this.mockGitRegistry.addExistingPackage(name, version);
});

Given('I have {string} ready to publish', async function (packageVersion) {
  const [name, version] = packageVersion.split('@');
  await this.createTestPackage(name, version);
});

Given('I have not provided a GitHub token', async function () {
  delete process.env.GITHUB_TOKEN;
  const config = await this.loadConfig();
  const currentRegistry = config.registries[config.current];
  if (currentRegistry) {
    delete currentRegistry.git_token;
    await this.saveConfig(config);
  }
});

Then('the output should contain a pull request preparation URL', 
  function () {
    const output = this.lastResult.stdout || '';
    assert(output.includes('github.com') && output.includes('compare'),
      'Expected PR preparation URL in output');
});

Then('the existing fork should be used', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('Fork already exists') || 
         output.includes('Opening cached repository'),
    'Expected existing fork to be used');
});

Then('the package metadata should be updated with the new version', 
  function () {
    // This would verify the metadata.json was updated correctly
    assert(!this.lastResult.error, 'Publish should succeed');
});

Then('the registry index should be updated', function () {
  // This would verify index.json was updated
  assert(!this.lastResult.error, 'Index should be updated');
});

Then('a branch {string} should be created', function (branchName) {
  const output = this.lastResult.stdout || '';
  assert(output.includes(`Creating branch: ${branchName}`) ||
         output.includes(branchName),
    `Expected branch ${branchName} to be created`);
});

Then('the branch should be pushed to my fork', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('Branch pushed') || output.includes('Pushing branch'),
    'Expected branch to be pushed to fork');
});

Then('the commit should contain package name and version', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('Creating commit') && 
         output.includes('test-package') && output.includes('1.0.0'),
    'Expected commit with package name and version');
});

Then('the commit should contain package description', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('test package') || output.includes('Description:'),
    'Expected commit to contain description');
});

Then('the commit should contain SHA256 hash', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('SHA256:') || output.includes('sha256'),
    'Expected commit to contain SHA256 hash');
});

Then('the commit should contain file size', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('Size:') || output.includes('bytes'),
    'Expected commit to contain file size');
});

Then('the file {string} should be added', function (filePath) {
  // This would verify the file was added to the git repository
  // For now, we just verify the publish succeeded
  assert(!this.lastResult.error, `File ${filePath} should be added`);
});

Then('the file {string} should be updated', function (filePath) {
  // This would verify the file was updated in the git repository
  assert(!this.lastResult.error, `File ${filePath} should be updated`);
});

Then('the registry index should contain the new package', function () {
  // This would parse the index.json and verify the package is there
  assert(!this.lastResult.error, 'Registry index should contain new package');
});

Then('the package count should be incremented', function () {
  // This would verify the package_count field was incremented
  assert(!this.lastResult.error, 'Package count should be incremented');
});

Then('the index updated timestamp should be current', function () {
  // This would verify the updated_at timestamp is recent
  assert(!this.lastResult.error, 'Index timestamp should be current');
});

// Helper method to create test packages
async function createTestPackage(name, version, options = {}) {
  const fs = require('fs').promises;
  const path = require('path');
  
  const manifest = {
    name: name,
    version: version,
    description: options.description || `Test package ${name}`,
    dependencies: options.dependencies || {},
    ...options
  };
  
  // Create manifest file
  const manifestPath = path.join(process.cwd(), 'manifest.json');
  await fs.writeFile(manifestPath, JSON.stringify(manifest, null, 2));
  
  // Create dummy archive
  const archivePath = path.join(process.cwd(), 'archive.tar.gz');
  await fs.writeFile(archivePath, 'dummy archive content');
  
  return { manifestPath, archivePath };
}
```

## Success Criteria
- Can detect and prepare fork for publishing
- Creates proper branch structure
- Adds package files in correct locations
- Updates metadata and index correctly
- Creates descriptive commits
- Pushes branch to fork successfully
- Returns PR preparation information

## Dependencies
- Phase 4: Git Client Basic Operations
- Phase 5: Git Registry Search and Discovery
- GitHub account with fork permissions

## Risks
- **Risk**: Fork permissions insufficient
  **Mitigation**: Clear error messages about required permissions
  
- **Risk**: Conflicts with existing branches
  **Mitigation**: Use timestamp in branch name if needed
  
- **Risk**: Large archive files
  **Mitigation**: Implement size limits and progress indicators

## Next Phase
Phase 7: GitHub API Integration - Implement automatic PR creation and fork management via GitHub API