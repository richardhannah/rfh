# Phase 4: Git Client Basic Operations

## Overview
Implement the foundational Git client that can clone, pull, and navigate Git-based package registries. This phase focuses on repository management and basic Git operations while maintaining consistency with the existing HTTP client architecture.

## Scope
- Create GitClient struct implementing RegistryClient interface
- Implement repository cloning and caching with optimized structure
- Add authentication support for private repositories
- Implement basic health check with standardized error handling
- Set up repository structure navigation
- Ensure interface compatibility with existing client architecture

## Prerequisites
- Phase 1: Registry Type Core Architecture completed  
- Phase 2: Registry Client Interface completed
- Phase 3: HTTP Client Refactoring completed (for interface consistency)
- Check and add go-git dependency if needed

## Implementation Steps

### 1. Verify and Add Dependencies

**Check existing dependencies first:**

```bash
go list -m all | grep git
```

**Add go-git dependency if needed:**

```bash
go get github.com/go-git/go-git/v5
go get github.com/go-git/go-git/v5/plumbing/transport/http
```

**Verify compatibility with existing modules.**

### 2. Create Git Client Structure

**File**: `internal/client/git.go` (new file)

```go
package client

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
    
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/transport"
    "github.com/go-git/go-git/v5/plumbing/transport/http"
    
    "rulestack/internal/config"
)

// GitClient implements RegistryClient for Git-based registries
type GitClient struct {
    repoURL   string
    gitToken  string
    verbose   bool
    cacheDir  string
    repo      *git.Repository
    mu        sync.Mutex // Protects repo operations
}

// Ensure GitClient implements RegistryClient
var _ RegistryClient = (*GitClient)(nil)

// NewGitClient creates a new Git registry client
func NewGitClient(repoURL, gitToken string, verbose bool) (*GitClient, error) {
    // Clean up repo URL
    repoURL = strings.TrimRight(repoURL, "/")
    if !strings.HasSuffix(repoURL, ".git") {
        repoURL += ".git"
    }
    
    // Determine cache directory
    cacheDir, err := getGitCacheDir(repoURL)
    if err != nil {
        return nil, fmt.Errorf("failed to determine cache directory: %w", err)
    }
    
    return &GitClient{
        repoURL:  repoURL,
        gitToken: gitToken,
        verbose:  verbose,
        cacheDir: cacheDir,
    }, nil
}

// Type returns the registry type
func (c *GitClient) Type() config.RegistryType {
    return config.RegistryTypeGit
}
```

### 3. Implement Cache Directory Management

**File**: `internal/client/git.go`

```go
// getGitCacheDir returns the cache directory for a Git repository
func getGitCacheDir(repoURL string) (string, error) {
    // Get base cache directory (align with existing patterns)
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    
    // Create a simpler hash of the repo URL
    h := sha256.Sum256([]byte(repoURL))
    dirName := hex.EncodeToString(h[:8])
    
    // Extract repo name for readability
    parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
    repoName := parts[len(parts)-1]
    
    // Use consistent cache structure
    cacheDir := filepath.Join(homeDir, ".rfh", "cache", "git", fmt.Sprintf("%s-%s", repoName, dirName))
    return cacheDir, nil
}

// ensureRepo ensures the repository is cloned and up to date
func (c *GitClient) ensureRepo(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Check if already cloned
    if c.repo != nil {
        return c.pullLatest(ctx)
    }
    
    // Check if cache directory exists
    if _, err := os.Stat(filepath.Join(c.cacheDir, ".git")); err == nil {
        // Repository exists, open it
        if c.verbose {
            fmt.Printf("📂 Opening cached repository at %s\n", c.cacheDir)
        }
        
        repo, err := git.PlainOpen(c.cacheDir)
        if err != nil {
            return fmt.Errorf("failed to open repository: %w", err)
        }
        
        c.repo = repo
        return c.pullLatest(ctx)
    }
    
    // Clone repository
    return c.cloneRepo(ctx)
}
```

### 4. Implement Clone Operation

**File**: `internal/client/git.go`

```go
// cloneRepo clones the repository to the cache directory
func (c *GitClient) cloneRepo(ctx context.Context) error {
    if c.verbose {
        fmt.Printf("📥 Cloning repository %s\n", c.repoURL)
        fmt.Printf("📂 Cache directory: %s\n", c.cacheDir)
    }
    
    // Create cache directory
    if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
        return fmt.Errorf("failed to create cache directory: %w", err)
    }
    
    // Prepare clone options
    cloneOpts := &git.CloneOptions{
        URL:      c.repoURL,
        Progress: nil,
    }
    
    if c.verbose {
        cloneOpts.Progress = os.Stdout
    }
    
    // Add authentication if token provided
    if c.gitToken != "" {
        cloneOpts.Auth = c.getAuth()
    }
    
    // Clone with context
    repo, err := git.PlainCloneContext(ctx, c.cacheDir, false, cloneOpts)
    if err != nil {
        if err == transport.ErrAuthenticationRequired {
            return NewRegistryError(ErrUnauthorized, 
                "authentication required - provide a Git token for private repositories")
        }
        return NewRegistryError(ErrConnectionFailed, fmt.Sprintf("failed to clone repository: %v", err))
    }
    
    c.repo = repo
    
    if c.verbose {
        fmt.Printf("✅ Repository cloned successfully\n")
    }
    
    return nil
}
```

### 5. Implement Pull Operation

**File**: `internal/client/git.go`

```go
// pullLatest pulls the latest changes from the remote repository
func (c *GitClient) pullLatest(ctx context.Context) error {
    if c.repo == nil {
        return fmt.Errorf("repository not initialized")
    }
    
    w, err := c.repo.Worktree()
    if err != nil {
        return fmt.Errorf("failed to get worktree: %w", err)
    }
    
    if c.verbose {
        fmt.Printf("🔄 Pulling latest changes\n")
    }
    
    // Prepare pull options
    pullOpts := &git.PullOptions{
        RemoteName: "origin",
    }
    
    if c.verbose {
        pullOpts.Progress = os.Stdout
    }
    
    // Add authentication if token provided
    if c.gitToken != "" {
        pullOpts.Auth = c.getAuth()
    }
    
    // Pull with context
    err = w.PullContext(ctx, pullOpts)
    if err != nil && err != git.NoErrAlreadyUpToDate {
        if err == transport.ErrAuthenticationRequired {
            return NewRegistryError(ErrUnauthorized,
                "authentication required - provide a Git token for private repositories")
        }
        return fmt.Errorf("failed to pull latest changes: %w", err)
    }
    
    if err == git.NoErrAlreadyUpToDate && c.verbose {
        fmt.Printf("✅ Already up to date\n")
    } else if c.verbose {
        fmt.Printf("✅ Pulled latest changes\n")
    }
    
    return nil
}
```

### 6. Implement Authentication Helper

**File**: `internal/client/git.go`

```go
// getAuth returns authentication configuration
func (c *GitClient) getAuth() transport.AuthMethod {
    if c.gitToken == "" {
        return nil
    }
    
    // Support multiple Git hosting providers
    username := "token"
    
    // Detect provider and use appropriate auth
    switch {
    case strings.Contains(c.repoURL, "gitlab.com"):
        username = "oauth2"
    case strings.Contains(c.repoURL, "bitbucket.org"):
        username = "x-token-auth"
    default: // GitHub and others
        username = "token"
    }
    
    return &http.BasicAuth{
        Username: username,
        Password: c.gitToken,
    }
}
```

### 7. Implement Health Check

**File**: `internal/client/git.go`

```go
// Health checks if the Git registry is accessible
func (c *GitClient) Health(ctx context.Context) error {
    // Try to ensure repository is accessible
    if err := c.ensureRepo(ctx); err != nil {
        return NewRegistryError(ErrConnectionFailed, fmt.Sprintf("registry health check failed: %v", err))
    }
    
    // Verify expected structure exists (packages directory or index.json)
    packagesDir := filepath.Join(c.cacheDir, "packages")
    indexPath := filepath.Join(c.cacheDir, "index.json")
    
    hasPackages := false
    hasIndex := false
    
    if _, err := os.Stat(packagesDir); err == nil {
        hasPackages = true
    }
    if _, err := os.Stat(indexPath); err == nil {
        hasIndex = true
    }
    
    if !hasPackages && !hasIndex {
        return NewRegistryError(ErrInvalidRegistry, "invalid registry structure: neither packages directory nor index.json found")
    }
    
    if c.verbose {
        fmt.Printf("✅ Git registry is healthy (packages: %v, index: %v)\n", hasPackages, hasIndex)
    }
    
    return nil
}
```

### 8. Implement Repository Structure Helpers

**File**: `internal/client/git.go`

```go
// getPackagePath returns the path to a package directory
func (c *GitClient) getPackagePath(packageName string) string {
    return filepath.Join(c.cacheDir, "packages", packageName)
}

// getVersionPath returns the path to a package version directory
func (c *GitClient) getVersionPath(packageName, version string) string {
    return filepath.Join(c.getPackagePath(packageName), "versions", version)
}

// getIndexPath returns the path to the registry index file
func (c *GitClient) getIndexPath() string {
    return filepath.Join(c.cacheDir, "index.json")
}

// packageExists checks if a package exists in the repository
func (c *GitClient) packageExists(packageName string) bool {
    path := c.getPackagePath(packageName)
    _, err := os.Stat(path)
    return err == nil
}

// versionExists checks if a package version exists
func (c *GitClient) versionExists(packageName, version string) bool {
    path := c.getVersionPath(packageName, version)
    _, err := os.Stat(path)
    return err == nil
}
```

### 9. Implement Cleanup Method

**File**: `internal/client/git.go`

```go
// Clean removes the cached repository
func (c *GitClient) Clean() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.verbose {
        fmt.Printf("🧹 Cleaning cache directory: %s\n", c.cacheDir)
    }
    
    c.repo = nil
    return os.RemoveAll(c.cacheDir)
}

// SetVerbose enables or disables verbose output
func (c *GitClient) SetVerbose(verbose bool) {
    c.verbose = verbose
}
```

### 10. Add Interface Compliance Methods

**File**: `internal/client/git.go`

```go
// Interface implementations - completed in subsequent phases

func (c *GitClient) SearchPackages(ctx context.Context, opts SearchOptions) ([]Package, error) {
    return nil, NewRegistryError(ErrNotImplemented, "Git registry search not yet implemented - see Phase 5")
}

func (c *GitClient) GetPackage(ctx context.Context, name string) (*Package, error) {
    return nil, NewRegistryError(ErrNotImplemented, "Git registry package retrieval not yet implemented - see Phase 5")
}

func (c *GitClient) GetPackageVersion(ctx context.Context, name, version string) (*PackageVersion, error) {
    return nil, NewRegistryError(ErrNotImplemented, "Git registry version retrieval not yet implemented - see Phase 5")
}

func (c *GitClient) PublishPackage(ctx context.Context, manifestPath, archivePath string) (*PublishResult, error) {
    return nil, NewRegistryError(ErrNotImplemented, "Git registry publishing not yet implemented - see Phase 6")
}

func (c *GitClient) DownloadBlob(ctx context.Context, sha256Hash, destPath string) error {
    return NewRegistryError(ErrNotImplemented, "Git registry blob download not yet implemented - see Phase 5")
}
```

### 11. Update Factory to Support Git Client

**File**: `internal/client/factory.go`

```go
// Update GetClient function to include Git support
func GetClient(cfg *config.Config, verbose bool) (RegistryClient, error) {
    registry := cfg.GetCurrentRegistry()
    if registry == nil {
        return nil, fmt.Errorf("no active registry configured")
    }

    switch registry.Type {
    case config.RegistryTypeHTTP:
        return NewHTTPClient(registry.URL, registry.JWTToken, verbose), nil
    case config.RegistryTypeGit:
        return NewGitClient(registry.URL, registry.GitToken, verbose)
    default:
        return nil, fmt.Errorf("unsupported registry type: %s", registry.Type)
    }
}
```

## Testing Requirements

### Unit Tests
1. Test cache directory generation
2. Test authentication configuration
3. Test repository structure helpers
4. Mock git operations for testing

### Integration Tests
1. Test cloning public repository
2. Test cloning with authentication
3. Test pull operations
4. Test health check on valid/invalid repos

### Manual Testing Checklist
- [ ] Clone public GitHub repository
- [ ] Clone private GitHub repository with token
- [ ] Pull latest changes
- [ ] Handle authentication failures gracefully
- [ ] Verify cache directory structure
- [ ] Test with GitLab repositories

### Cucumber Test Amendments

**File**: `features/git-registry-basic.feature` (new file)

```gherkin
Feature: Git Registry Basic Operations
  Git registry should support basic repository operations

  Background:
    Given I have a clean test environment

  Scenario: Add public Git registry
    When I run "rfh registry add github-public https://github.com/test-org/public-registry --type git"
    Then the command should succeed
    And the config file should contain registry "github-public" with type "git"

  Scenario: Health check for public Git registry
    Given a registry "github-public" with URL "https://github.com/test-org/test-registry" and type "git"
    And I use registry "github-public"
    When I run "rfh registry health"
    Then the command should succeed
    And the output should contain "registry is healthy"

  Scenario: Health check for private Git registry without token
    Given a registry "github-private" with URL "https://github.com/test-org/private-registry" and type "git"
    And I use registry "github-private"
    When I run "rfh registry health"
    Then the command should fail
    And the output should contain "authentication required"

  Scenario: Health check for private Git registry with token
    Given a registry "github-private" with URL "https://github.com/test-org/private-registry" and type "git"
    And the registry has a valid git token
    And I use registry "github-private"
    When I run "rfh registry health"
    Then the command should succeed

  Scenario: Cache directory is created
    Given a registry "github-test" with URL "https://github.com/test-org/test-registry" and type "git"
    And I use registry "github-test"
    When I run "rfh registry health"
    Then the command should succeed
    And the cache directory should exist for the repository

  Scenario: Invalid Git repository URL
    Given a registry "invalid-git" with URL "https://example.com/not-a-git-repo" and type "git"
    And I use registry "invalid-git"
    When I run "rfh registry health"
    Then the command should fail
    And the output should contain "failed to clone repository"

  Scenario: GitLab repository support
    Given a registry "gitlab-test" with URL "https://gitlab.com/test-org/test-registry" and type "git"
    And I use registry "gitlab-test"
    When I run "rfh registry health"
    Then the command should succeed

  Scenario: Repository update on subsequent operations
    Given a registry "github-test" with URL "https://github.com/test-org/test-registry" and type "git"
    And I use registry "github-test"
    And I run "rfh registry health" successfully once
    When I run "rfh registry health" again
    Then the command should succeed
    And the repository should be updated from remote
```

**File**: `features/step_definitions/git_client_steps.js`

Add new step definitions:
```javascript
Given('the registry has a valid git token', async function () {
  // Set environment variable or config
  process.env.GITHUB_TOKEN = 'test-token-123';
  
  // Update config to include token
  const config = await this.loadConfig();
  const currentRegistry = config.registries[config.current];
  if (currentRegistry) {
    currentRegistry.git_token = 'test-token-123';
    await this.saveConfig(config);
  }
});

Then('the cache directory should exist for the repository', async function () {
  // Check if cache directory was created
  const fs = require('fs');
  const path = require('path');
  const os = require('os');
  
  const cacheDir = path.join(os.homedir(), '.rfh', 'cache', 'git');
  assert(fs.existsSync(cacheDir), 'Git cache directory should exist');
  
  // Check if specific repo cache exists
  const repoDirs = fs.readdirSync(cacheDir);
  assert(repoDirs.length > 0, 'Repository cache directory should be created');
});

Then('the repository should be updated from remote', async function () {
  // This would be verified through logging in verbose mode
  // For now, we just ensure the command succeeded
  assert(!this.lastResult.error, 'Repository update should succeed');
});

When('I run {string} successfully once', async function (command) {
  await this.runCommand(command);
  assert(!this.lastResult.error, `First run should succeed: ${this.lastResult.error}`);
});

When('I run {string} again', async function (command) {
  await this.runCommand(command);
});
```

**File**: `features/step_definitions/registry_steps.js` (add to existing)

```javascript
When('I run {string}', async function (command) {
  this.lastResult = await this.runCommand(command);
});

Then('the command should succeed', function () {
  assert(!this.lastResult.error, `Command should succeed but got error: ${this.lastResult.error}`);
  assert(this.lastResult.exitCode === 0 || this.lastResult.exitCode === undefined, 
    `Expected exit code 0, got ${this.lastResult.exitCode}`);
});

Then('the command should fail', function () {
  assert(this.lastResult.error || this.lastResult.exitCode !== 0, 
    'Command should fail but succeeded');
});

Then('the output should contain {string}', function (expectedText) {
  const output = this.lastResult.stdout || this.lastResult.stderr || '';
  assert(output.includes(expectedText), 
    `Expected output to contain "${expectedText}", got: ${output}`);
});
```

## Success Criteria
- Can clone and cache Git repositories
- Authentication works for private repositories
- Repository updates via pull work correctly
- Health check validates repository structure
- Thread-safe repository operations
- Verbose mode provides useful debugging output

## Dependencies
- Phase 1: Registry Type Core Architecture
- Phase 2: Registry Client Interface
- go-git package

## Risks
- **Risk**: Large repository size affecting performance
  **Mitigation**: Implement shallow clones in future enhancement
  
- **Risk**: Network issues during clone/pull
  **Mitigation**: Proper error handling and retry logic
  
- **Risk**: Disk space usage from cached repositories
  **Mitigation**: Add cleanup command and cache management

## Next Phase
Phase 5: Git Registry Search and Discovery - Implement package search, retrieval, and download operations