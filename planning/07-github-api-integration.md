# Phase 7: GitHub API Integration (Direct Collaborator Mode)

## Overview
**Completely replace** Phase 6 Git Registry Publishing with GitHub API integration for automatic pull request creation. This phase assumes users are collaborators with write access to registry repositories, eliminating the need for fork management and significantly simplifying the implementation.

## Changes from Original Planning (Major Simplification)
- **No Fork Management**: Users are collaborators with direct write access
- **Direct Repository Access**: Work directly on main registry repository
- **Same-Repo PRs**: Create PRs from branch to main on the same repository
- **Simplified Authentication**: Only need GitHub API for PR creation and user info
- **Complete Replacement**: Replace Phase 6 PublishPackage method entirely
- **Remove All Fork Logic**: Eliminate fork detection, creation, and management completely

## Scope
- **Replace Phase 6 PublishPackage method completely**
- Add GitHub API client for pull request creation only
- Create PRs directly from publish branches to main branch (same repository)
- Get authenticated user information for PR author details
- Handle API rate limiting and authentication gracefully
- **Remove all fork-related code from Phase 6**
- Maintain existing Phase 6 helper methods (createPublishBranch, addPackageFiles, etc.)
- Direct repository cloning and branch management only

## Prerequisites
- Phase 6: Git Registry Publishing completed ✅
- Users have collaborator/write access to registry repositories ✅
- GitHub personal access token with collaborator permissions ✅
- google/go-github library dependency (to be added)
- Existing go-git integration confirmed working ✅

## Required GitHub Token Permissions
- `repo` - Write access to repositories (collaborator level)
- OR `public_repo` - Write access to public repositories (collaborator level)
- **Note**: Users must be added as collaborators to registry repositories

## Implementation Steps

### 1. Add GitHub Library Dependency (Required)

```bash
go get github.com/google/go-github/v67/github
go get golang.org/x/oauth2
```

**Implementation Note**: These dependencies will be added as part of Phase 7 implementation.

**Note**: All Git operations (cloning, branching, commits, pushes) will use the go-git library which has been confirmed to work with GitHub personal access tokens. The GitHub API will only be used for:
- Pull request creation (same repository)
- User authentication information
- **No fork management needed** - users work directly on registry repositories

### 2. Enhance GitClient with GitHub API Integration

**File**: `internal/client/github_api.go` (new file)

```go
package client

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/go-github/v67/github"
    "golang.org/x/oauth2"
)

// GitHubClient handles GitHub API operations
// Integrates with existing GitClient from Phase 6
type GitHubClient struct {
    client  *github.Client
    verbose bool
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(token string, verbose bool) *GitHubClient {
    ctx := context.Background()
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(ctx, ts)
    
    client := github.NewClient(tc)
    
    return &GitHubClient{
        client:  client,
        verbose: verbose,
    }
}

```

### 3. Implement User Information Retrieval

**File**: `internal/client/github_api.go`

```go
// GetAuthenticatedUser gets information about the authenticated user
func (g *GitHubClient) GetAuthenticatedUser(ctx context.Context) (*github.User, error) {
    if g.verbose {
        fmt.Printf("🔍 Getting authenticated user info\n")
    }
    
    user, _, err := g.client.Users.Get(ctx, "")
    if err != nil {
        return nil, NewRegistryError(ErrUnauthorized, fmt.Sprintf("failed to get user info: %v", err))
    }
    
    if g.verbose {
        fmt.Printf("✅ Authenticated as: %s\n", user.GetLogin())
    }
    
    return user, nil
}
```

### 4. Direct Repository Access (Replace Fork Management)

**File**: `internal/client/github_api.go` (add repository parsing helper)

```go
// parseGitHubURL extracts owner and repo from GitHub URL
func parseGitHubURL(repoURL string) (owner, repo string, err error) {
    if !strings.Contains(repoURL, "github.com") {
        return "", "", fmt.Errorf("not a GitHub URL")
    }
    
    // Handle different URL formats
    repoURL = strings.TrimSuffix(repoURL, ".git")
    
    // Parse URL - handle both https://github.com/owner/repo and git@github.com:owner/repo
    var parts []string
    if strings.Contains(repoURL, "github.com/") {
        parts = strings.Split(repoURL, "/")
    } else if strings.Contains(repoURL, "github.com:") {
        parts = strings.Split(strings.Replace(repoURL, ":", "/", -1), "/")
    } else {
        return "", "", fmt.Errorf("invalid GitHub URL format")
    }
    
    if len(parts) < 2 {
        return "", "", fmt.Errorf("invalid GitHub URL format")
    }
    
    // Find github.com in parts and extract owner/repo
    for i, part := range parts {
        if part == "github.com" && i+2 < len(parts) {
            owner = parts[i+1]
            repo = parts[i+2]
            break
        }
    }
    
    if owner == "" || repo == "" {
        return "", "", fmt.Errorf("could not parse owner/repo from URL")
    }
    
    return owner, repo, nil
}

// GetRepository gets repository information (no fork needed)
func (g *GitHubClient) GetRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
    repository, _, err := g.client.Repositories.Get(ctx, owner, repo)
    if err != nil {
        return nil, NewRegistryError(ErrNotFound, fmt.Sprintf("failed to get repository %s/%s: %v", owner, repo, err))
    }
    
    if g.verbose {
        fmt.Printf("📁 Repository: %s (default branch: %s)\n", 
            repository.GetFullName(), repository.GetDefaultBranch())
    }
    
    return repository, nil
}
```

### 5. Implement Same-Repository Pull Request Creation

**File**: `internal/client/github_api.go`

```go
// CreatePullRequest creates a new pull request on the same repository
// This is for collaborators creating PRs from branch to main on the same repo
func (g *GitHubClient) CreatePullRequest(ctx context.Context, owner, repo, title, branchName, baseBranch, body string) (*github.PullRequest, error) {
    if g.verbose {
        fmt.Printf("📝 Creating pull request: %s\n", title)
        fmt.Printf("   Repository: %s/%s\n", owner, repo)
        fmt.Printf("   Branch: %s -> %s\n", branchName, baseBranch)
    }
    
    newPR := &github.NewPullRequest{
        Title:               github.String(title),
        Head:                github.String(branchName), // Just branch name (same repo)
        Base:                github.String(baseBranch), // Usually "main"
        Body:                github.String(body),
        MaintainerCanModify: github.Bool(true),
        Draft:               github.Bool(false),
    }
    
    pr, _, err := g.client.PullRequests.Create(ctx, owner, repo, newPR)
    if err != nil {
        // Check for common errors
        if strings.Contains(err.Error(), "pull request already exists") {
            return nil, fmt.Errorf("pull request already exists for branch %s", branchName)
        }
        if strings.Contains(err.Error(), "No commits between") {
            return nil, fmt.Errorf("no commits found on branch %s", branchName)
        }
        return nil, NewRegistryError(ErrInvalidOperation, fmt.Sprintf("failed to create PR: %v", err))
    }
    
    if g.verbose {
        fmt.Printf("✅ Pull request created: %s\n", pr.GetHTMLURL())
        fmt.Printf("   PR #%d: %s\n", pr.GetNumber(), pr.GetTitle())
    }
    
    return pr, nil
}

// GetPullRequest gets information about a pull request
func (g *GitHubClient) GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
    pr, _, err := g.client.PullRequests.Get(ctx, owner, repo, number)
    if err != nil {
        return nil, NewRegistryError(ErrNotFound, fmt.Sprintf("failed to get PR #%d: %v", number, err))
    }
    
    return pr, nil
}

// CheckCollaboratorAccess verifies user has write access to the repository
func (g *GitHubClient) CheckCollaboratorAccess(ctx context.Context, owner, repo string) error {
    user, err := g.GetAuthenticatedUser(ctx)
    if err != nil {
        return err
    }
    
    // Check if user is a collaborator with write access
    _, _, err = g.client.Repositories.GetPermissionLevel(ctx, owner, repo, user.GetLogin())
    if err != nil {
        return NewRegistryError(ErrUnauthorized, 
            fmt.Sprintf("user %s does not have access to %s/%s", user.GetLogin(), owner, repo))
    }
    
    if g.verbose {
        fmt.Printf("✅ User %s has access to %s/%s\n", user.GetLogin(), owner, repo)
    }
    
    return nil
}
```

### 5. Update Git Client for GitHub Integration

**File**: `internal/client/git_github.go` (new file)

```go
package client

import (
    "context"
    "fmt"
    "strings"
)

// parseGitHubURL extracts owner and repo from GitHub URL
func parseGitHubURL(repoURL string) (owner, repo string, err error) {
    if !strings.Contains(repoURL, "github.com") {
        return "", "", fmt.Errorf("not a GitHub URL")
    }
    
    // Handle different URL formats
    repoURL = strings.TrimSuffix(repoURL, ".git")
    
    // Parse URL
    parts := strings.Split(repoURL, "/")
    if len(parts) < 2 {
        return "", "", fmt.Errorf("invalid GitHub URL format")
    }
    
    owner = parts[len(parts)-2]
    repo = parts[len(parts)-1]
    
    // Handle github.com in path
    for i, part := range parts {
        if part == "github.com" && i+2 < len(parts) {
            owner = parts[i+1]
            repo = parts[i+2]
            break
        }
    }
    
    return owner, repo, nil
}

// createPullRequestForPackage creates a PR for package publication
func (c *GitClient) createPullRequestForPackage(ctx context.Context, branchName string, manifest *GitManifest) (*PullRequest, error) {
    // Parse repository URL
    owner, repo, err := parseGitHubURL(c.repoURL)
    if err != nil {
        return nil, fmt.Errorf("failed to parse repository URL: %w", err)
    }
    
    // Create GitHub client
    github := NewGitHubClient(c.gitToken, c.verbose)
    
    // This method is REMOVED - replaced by createPullRequestForPackage in the main git.go file
    // See the complete replacement PublishPackage method below
}
```

### 6. Replace Phase 6 PublishPackage Method Completely

**File**: Update `internal/client/git.go` (complete replacement of Phase 6 method)

```go
// PublishPackage publishes a package to the Git registry (Phase 7 - Direct Collaborator Mode)
// This completely replaces the Phase 6 fork-based implementation
func (c *GitClient) PublishPackage(ctx context.Context, manifestPath, archivePath string) (*PublishResult, error) {
    if c.verbose {
        fmt.Printf("📦 Publishing package to Git registry (direct collaborator mode)\n")
    }

    // Parse manifest for package info
    manifestData, err := os.ReadFile(manifestPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read manifest: %w", err)
    }
    var manifest GitManifest
    if err := json.Unmarshal(manifestData, &manifest); err != nil {
        return nil, fmt.Errorf("failed to parse manifest: %w", err)
    }

    // Work directly with the target repository (no fork management)
    repo, err := c.cloneRepository(ctx, c.repoURL)
    if err != nil {
        return nil, fmt.Errorf("failed to prepare repository: %w", err)
    }

    // Create publish branch (reuse existing Phase 6 helper)
    branchName, err := c.createPublishBranch(repo, manifest.Name, manifest.Version)
    if err != nil {
        return nil, fmt.Errorf("failed to create branch: %w", err)
    }

    // Add package files (reuse existing Phase 6 helper)
    if err := c.addPackageFiles(repo, manifestPath, archivePath); err != nil {
        return nil, fmt.Errorf("failed to add package files: %w", err)
    }

    // Update registry index (reuse existing Phase 6 helper)
    if err := c.updateRegistryIndex(repo, &manifest); err != nil {
        return nil, fmt.Errorf("failed to update index: %w", err)
    }

    // Create commit (reuse existing Phase 6 helper)
    _, err = c.createCommit(repo, &manifest)
    if err != nil {
        return nil, fmt.Errorf("failed to create commit: %w", err)
    }

    // Push branch to origin (same repository)
    if err := c.pushBranch(ctx, repo, branchName); err != nil {
        return nil, fmt.Errorf("failed to push branch: %w", err)
    }

    // Create pull request via GitHub API (same repository)
    pr, err := c.createPullRequestForPackage(ctx, branchName, &manifest)
    if err != nil {
        // If GitHub API fails, provide manual URL for same repository
        owner, repoName, _ := parseGitHubURL(c.repoURL)
        manualURL := fmt.Sprintf("https://github.com/%s/%s/compare/main...%s", 
            owner, repoName, branchName) // Same repo - direct collaborator access
        
        if c.verbose {
            fmt.Printf("⚠️ GitHub API PR creation failed: %v\n", err)
            fmt.Printf("💡 Branch pushed successfully. Create PR manually: %s\n", manualURL)
        }
        
        return &PublishResult{
            Name:    manifest.Name,
            Version: manifest.Version,
            SHA256:  manifest.SHA256,
            PRUrl:   manualURL,
            Message: fmt.Sprintf("Branch pushed. Create PR manually: %s", manualURL),
        }, nil
    }

    return &PublishResult{
        Name:    manifest.Name,
        Version: manifest.Version,
        SHA256:  manifest.SHA256,
        PRUrl:   pr.GetHTMLURL(),
        Message: fmt.Sprintf("Pull request created successfully: %s", pr.GetHTMLURL()),
    }, nil
}

// cloneRepository clones the target repository directly (no fork management)
func (c *GitClient) cloneRepository(ctx context.Context, repoURL string) (*git.Repository, error) {
    // Create cache directory for the repository
    cacheDir := c.cacheDir
    
    // Check if repository already cloned
    if _, err := os.Stat(filepath.Join(cacheDir, ".git")); err == nil {
        // Open existing repository
        repo, err := git.PlainOpen(cacheDir)
        if err != nil {
            return nil, fmt.Errorf("failed to open repository: %w", err)
        }
        
        // Update from remote
        if err := c.updateRepository(ctx, repo); err != nil {
            return nil, err
        }
        
        return repo, nil
    }
    
    // Clone repository
    if c.verbose {
        fmt.Printf("📥 Cloning repository: %s\n", repoURL)
    }
    
    cloneOpts := &git.CloneOptions{
        URL:      repoURL,
        Progress: nil,
    }
    
    if c.verbose {
        cloneOpts.Progress = os.Stdout
    }
    
    if c.gitToken != "" {
        cloneOpts.Auth = c.getAuth()
    }
    
    repo, err := git.PlainCloneContext(ctx, cacheDir, false, cloneOpts)
    if err != nil {
        return nil, fmt.Errorf("failed to clone repository: %w", err)
    }
    
    return repo, nil
}

// updateRepository updates the repository from remote
func (c *GitClient) updateRepository(ctx context.Context, repo *git.Repository) error {
    if c.verbose {
        fmt.Printf("🔄 Updating repository from remote\n")
    }
    
    // Fetch latest changes
    fetchOpts := &git.FetchOptions{
        RemoteName: "origin",
    }
    
    if c.gitToken != "" {
        fetchOpts.Auth = c.getAuth()
    }
    
    err := repo.FetchContext(ctx, fetchOpts)
    if err != nil && err != git.NoErrAlreadyUpToDate {
        return fmt.Errorf("failed to fetch from remote: %w", err)
    }
    
    return nil
}

// createPullRequestForPackage creates a PR for package publication (same repository)
func (c *GitClient) createPullRequestForPackage(ctx context.Context, branchName string, manifest *GitManifest) (*github.PullRequest, error) {
    // Parse repository URL directly
    owner, repo, err := parseGitHubURL(c.repoURL)
    if err != nil {
        return nil, fmt.Errorf("failed to parse repository URL: %w", err)
    }
    
    // Create GitHub client
    githubClient := NewGitHubClient(c.gitToken, c.verbose)
    
    // Verify collaborator access
    if err := githubClient.CheckCollaboratorAccess(ctx, owner, repo); err != nil {
        return nil, fmt.Errorf("access check failed: %w", err)
    }
    
    // Get repository information
    repository, err := githubClient.GetRepository(ctx, owner, repo)
    if err != nil {
        return nil, fmt.Errorf("failed to get repository info: %w", err)
    }
    
    // Get authenticated user
    user, err := githubClient.GetAuthenticatedUser(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get user info: %w", err)
    }
    
    // Create PR body
    body := fmt.Sprintf(`## 📦 Package Publication Request

**Package**: %s  
**Version**: %s  
**Description**: %s

### Package Details
- **SHA256**: %s
- **Size**: %d bytes
- **Publisher**: %s

### Changes
- Added package files to `+"`packages/%s/versions/%s/`"+`
- Updated package metadata  
- Updated registry index

---
*This pull request was automatically generated by RuleStack CLI*`,
        manifest.Name,
        manifest.Version,
        manifest.Description,
        manifest.SHA256,
        manifest.Size,
        user.GetLogin(),
        manifest.Name,
        manifest.Version)
    
    // Create pull request (same repository: branch -> main)
    title := fmt.Sprintf("Publish %s@%s", manifest.Name, manifest.Version)
    baseBranch := repository.GetDefaultBranch()
    
    pr, err := githubClient.CreatePullRequest(ctx, owner, repo, title, branchName, baseBranch, body)
    if err != nil {
        return nil, err
    }
    
    return pr, nil
}
```

### 7. Add Rate Limit Handling

**File**: `internal/client/github_api.go` (update the GetRateLimit method already defined above)

```go
// WaitForRateLimit waits if rate limit is low
func (g *GitHubClient) WaitForRateLimit(ctx context.Context) error {
    rateLimit, _, err := g.client.RateLimits.Get(ctx)
    if err != nil {
        return fmt.Errorf("failed to get rate limit: %v", err)
    }
    
    core := rateLimit.GetCore()
    if core.Remaining < 10 {
        waitTime := time.Until(core.Reset.Time)
        if g.verbose {
            fmt.Printf("⏳ Rate limit low (%d remaining). Waiting %v\n", 
                core.Remaining, waitTime)
        }
        
        select {
        case <-time.After(waitTime):
            return nil
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return nil
}

// CheckRateLimit provides rate limit information for debugging
func (g *GitHubClient) CheckRateLimit(ctx context.Context) error {
    rateLimit, _, err := g.client.RateLimits.Get(ctx)
    if err != nil {
        return fmt.Errorf("failed to get rate limit: %v", err)
    }
    
    core := rateLimit.GetCore()
    if g.verbose {
        fmt.Printf("📊 Rate limit: %d/%d remaining (resets at %v)\n", 
            core.Remaining, core.Limit, core.Reset.Time)
    }
    
    return nil
}
```

## Testing Requirements

### Unit Tests (Phase 7 Implementation)
1. Test GitHub URL parsing
2. Test PR body generation  
3. Test rate limit handling
4. Mock GitHub API responses

### Integration Tests (Phase 7 Implementation)
1. Test user authentication
2. Test collaborator access verification
3. Test PR creation (same repository)
4. Test rate limit checking
5. Test direct repository access

### Manual Testing Checklist
- [ ] Authenticate with GitHub token
- [ ] Verify collaborator access to registry repository
- [ ] Clone repository directly
- [ ] Create pull request (same repository)
- [ ] Handle rate limiting
- [ ] PR contains correct information
- [ ] Works with private repositories
- [ ] Multiple registries with different tokens work correctly

### Cucumber Test Amendments (Deferred)

**Note**: Cucumber testing for GitHub integration will be deferred until real GitHub repositories and test keys are set up for integration testing.

## Success Criteria
- Can authenticate with GitHub token and verify collaborator access
- Creates well-formatted pull requests on the same repository (branch → main)
- Works directly with target repository (no fork management)
- Handles rate limiting gracefully
- Provides clear error messages for access issues
- Returns PR URL for user reference
- Gracefully falls back to manual same-repository PR creation if API fails
- **Supports multiple registries with different GitHub tokens (hard requirement)**
- **Maintains backwards compatibility with HTTP registries (no changes)**

## Dependencies
- Phase 6: Git Registry Publishing (completed)
- GitHub personal access token with collaborator permissions
- Network access to GitHub API
- User must be added as collaborator to registry repositories

## Risks
- **Risk**: GitHub API rate limits
  **Mitigation**: Implement rate limit checking and waiting
  
- **Risk**: Insufficient collaborator permissions
  **Mitigation**: Access verification before operations, clear error messages
  
- **Risk**: API failures during PR creation
  **Mitigation**: Graceful fallback to manual PR creation workflow
  
- **Risk**: Branch conflicts or existing PRs
  **Mitigation**: Check for existing branches/PRs, provide clear error messages

## Configuration

Add to CLI config:
```toml
[registries.github-registry]
url = "https://github.com/org/registry"  # Registry repo where user is collaborator
type = "git"
git_token = "ghp_xxxxxxxxxxxx"         # Token with repo access
```

Environment variables:
```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx    # Personal access token
export GIT_AUTHOR_NAME="Your Name"      # Commit author name
export GIT_AUTHOR_EMAIL="your.email@example.com"  # Commit author email
```

**Important**: Users must be added as collaborators to the registry repository with write access.

## Backwards Compatibility Requirements

### Remote-HTTP Registry Support (Must Maintain)
- **Keep existing HTTP registry functionality** unchanged
- Phase 7 changes apply **only to Git registries** 
- HTTP registries continue to work as before (no changes needed)

### Multi-Registry Token Support (Hard Requirement)
- Support different GitHub tokens per registry
- Each Git registry can have its own `git_token` configuration
- Token selection per registry for authentication and API calls

Example multi-registry configuration:
```toml
[registries.company-internal]
url = "https://github.com/company/internal-registry"
type = "git"
git_token = "ghp_company_internal_token"

[registries.open-source]
url = "https://github.com/oss/public-registry"  
type = "git"
git_token = "ghp_oss_token"

[registries.http-registry]
url = "https://registry.example.com/api"
type = "http"
# No changes to HTTP registries
```

### Fix-Forward Approach
- **No backwards compatibility** for Phase 6 fork-based Git publishing
- Complete replacement approach - clean break from fork logic
- Focus on maintaining HTTP registry compatibility only

## Summary of Simplified Workflow

The Phase 7 implementation creates a **direct collaborator workflow** that is much simpler than the original fork-based approach:

### **Original Complex Workflow (Avoided):**
1. Detect if fork exists → Create fork if needed → Clone fork → Sync with upstream → Create branch on fork → Push to fork → Create PR from fork to upstream

### **New Simplified Workflow:**
1. Clone registry repository directly → Create branch → Add files → Commit → Push branch → Create PR (same repo)

### **Key Benefits:**
✅ **Eliminates Fork Management**: No fork detection, creation, or synchronization  
✅ **Direct Repository Access**: Work directly on the target repository  
✅ **Simplified Authentication**: Only need collaborator access, no fork permissions  
✅ **Reduced API Calls**: Fewer GitHub API operations required  
✅ **Better Performance**: No fork cloning or synchronization delays  
✅ **Clearer Error Messages**: Direct feedback about repository access issues  

This approach is **perfect for organizational workflows** where team members are collaborators on shared registry repositories.

## Implementation Notes

### Complete Phase 6 Replacement
- **Remove** all fork-related code from Phase 6 (`git_fork.go`)
- **Replace** the entire `PublishPackage` method in `internal/client/git.go`
- **Keep** existing Phase 6 helper methods (`createPublishBranch`, `addPackageFiles`, `updateRegistryIndex`, `createCommit`, `pushBranch`)
- **Add** new GitHub API integration files (`github_api.go`)
- **Update** fallback URLs to point to same-repository comparisons (no fork URLs)
- **Branch cleanup** is deferred - Phase 7 will not handle cleanup of publish branches after PR merge

### Files to Remove
- `internal/client/git_fork.go` (entire file - fork management no longer needed)

### Files to Create
- `internal/client/github_api.go` (GitHub API client and operations)

### Files to Modify
- `internal/client/git.go` (replace `PublishPackage` method completely)
- `go.mod` (add GitHub API dependencies)

The result is a **much simpler and more reliable** publishing workflow focused on direct collaborator access.