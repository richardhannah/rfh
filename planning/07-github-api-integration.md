# Phase 7: GitHub API Integration

## Overview
Integrate GitHub API minimally to automate pull request creation and user authentication while using go-git for all Git operations. This phase completes the Git registry publishing workflow with lightweight API integration.

## Scope
- Implement minimal GitHub API client using google/go-github library  
- Use go-git library for all Git operations (cloning, branching, commits, pushes)
- Add automatic fork detection and creation via GitHub API
- Implement comprehensive PR creation with proper formatting
- Get authenticated user information reliably
- Handle API rate limiting and errors gracefully  
- Add PR status checking capabilities
- Maintain consistency with existing error handling patterns

## Prerequisites
- Phase 6: Git Registry Publishing completed
- GitHub personal access token with appropriate permissions
- google/go-github library dependency
- go-git library (confirmed working with GitHub/Gitea)

## Required GitHub Token Permissions
- `repo` - Full control of private repositories (if using private registries)
- `public_repo` - Access to public repositories
- `workflow` - Update GitHub Action workflows (if needed)

## Implementation Steps

### 1. Add GitHub Library Dependency

```bash
go get github.com/google/go-github/v67/github
go get golang.org/x/oauth2
```

**Note**: All Git operations (cloning, branching, commits, pushes) will use the go-git library which has been confirmed to work with GitHub personal access tokens. The GitHub API will only be used for:
- Fork detection and creation
- Pull request creation
- User authentication information

### 2. Create GitHub API Client

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

### 4. Implement Fork Management

**File**: `internal/client/github_api.go`

```go

// CreateFork creates a fork of the specified repository
func (g *GitHubClient) CreateFork(ctx context.Context, owner, repo string) (*github.Repository, error) {
    if g.verbose {
        fmt.Printf("🍴 Creating fork of %s/%s\n", owner, repo)
    }
    
    // Create fork with empty options (uses defaults)
    fork, _, err := g.client.Repositories.CreateFork(ctx, owner, repo, nil)
    if err != nil {
        return nil, NewRegistryError(ErrUnauthorized, fmt.Sprintf("failed to create fork: %v", err))
    }
    
    if g.verbose {
        fmt.Printf("✅ Fork created: %s\n", fork.GetFullName())
    }
    
    // Wait for fork to be ready (GitHub needs time to set up the repo)
    time.Sleep(5 * time.Second)
    
    return fork, nil
}

// GetFork checks if a fork exists for the authenticated user
func (g *GitHubClient) GetFork(ctx context.Context, owner, repo string) (*github.Repository, error) {
    // Get authenticated user
    user, err := g.GetAuthenticatedUser(ctx)
    if err != nil {
        return nil, err
    }
    
    // Check if fork exists
    fork, _, err := g.client.Repositories.Get(ctx, user.GetLogin(), repo)
    if err != nil {
        // Check if it's a 404 (fork doesn't exist)
        if _, ok := err.(*github.ErrorResponse); ok {
            return nil, nil // Fork doesn't exist
        }
        return nil, NewRegistryError(ErrConnectionFailed, fmt.Sprintf("failed to check for fork: %v", err))
    }
    
    // Verify it's actually a fork of the target repo
    if !fork.GetFork() || fork.GetParent() == nil ||
       fork.GetParent().GetFullName() != fmt.Sprintf("%s/%s", owner, repo) {
        return nil, nil // Not a fork of the target
    }
    
    return fork, nil
}

// EnsureFork ensures a fork exists, creating one if necessary
func (g *GitHubClient) EnsureFork(ctx context.Context, owner, repo string) (*github.Repository, error) {
    // Check if fork already exists
    fork, err := g.GetFork(ctx, owner, repo)
    if err != nil {
        return nil, err
    }
    
    if fork != nil {
        if g.verbose {
            fmt.Printf("✅ Fork already exists: %s\n", fork.GetFullName())
        }
        return fork, nil
    }
    
    // Create fork
    return g.CreateFork(ctx, owner, repo)
}
```

### 5. Implement Pull Request Creation

**File**: `internal/client/github_api.go`

```go

// CreatePullRequest creates a new pull request
func (g *GitHubClient) CreatePullRequest(ctx context.Context, owner, repo, title, head, base, body string) (*github.PullRequest, error) {
    if g.verbose {
        fmt.Printf("📝 Creating pull request: %s\n", title)
        fmt.Printf("   Base: %s <- Head: %s\n", base, head)
    }
    
    newPR := &github.NewPullRequest{
        Title:               github.String(title),
        Head:                github.String(head),
        Base:                github.String(base),
        Body:                github.String(body),
        MaintainerCanModify: github.Bool(true),
        Draft:               github.Bool(false),
    }
    
    pr, _, err := g.client.PullRequests.Create(ctx, owner, repo, newPR)
    if err != nil {
        return nil, NewRegistryError(ErrInvalidOperation, fmt.Sprintf("failed to create PR: %v", err))
    }
    
    if g.verbose {
        fmt.Printf("✅ Pull request created: %s\n", pr.GetHTMLURL())
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

// GetRateLimit gets current rate limit status
func (g *GitHubClient) GetRateLimit(ctx context.Context) (*github.RateLimits, error) {
    rateLimit, _, err := g.client.RateLimits.Get(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get rate limit: %v", err)
    }
    
    return rateLimit, nil
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
    
    // Get authenticated user
    user, err := github.GetAuthenticatedUser(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get user info: %w", err)
    }
    
    // Ensure fork exists
    fork, err := github.EnsureFork(ctx, owner, repo)
    if err != nil {
        return nil, fmt.Errorf("failed to ensure fork: %w", err)
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
- Added package files to ` + "`packages/%s/versions/%s/`" + `
- Updated package metadata
- Updated registry index

---
*This pull request was automatically generated by RuleStack CLI*`,
        manifest.Name,
        manifest.Version,
        manifest.Description,
        manifest.SHA256,
        manifest.Size,
        user.Login,
        manifest.Name,
        manifest.Version)
    
    // Create PR request
    prRequest := &PullRequestRequest{
        Title: fmt.Sprintf("Publish %s@%s", manifest.Name, manifest.Version),
        Head:  fmt.Sprintf("%s:%s", user.Login, branchName),
        Base:  fork.Parent.DefaultBranch,
        Body:  body,
        MaintainerCanModify: true,
        Draft: false,
    }
    
    // Create pull request
    pr, err := github.CreatePullRequest(ctx, owner, repo, prRequest)
    if err != nil {
        // Check if PR already exists
        if strings.Contains(err.Error(), "pull request already exists") {
            return nil, fmt.Errorf("pull request already exists for this branch")
        }
        return nil, err
    }
    
    return pr, nil
}
```

### 6. Update Main Publish Method

**File**: Update `internal/client/git.go`

```go
// PublishPackage publishes a package to the Git registry (updated)
func (c *GitClient) PublishPackage(ctx context.Context, manifestPath, archivePath string) (*PublishResult, error) {
    if c.verbose {
        fmt.Printf("📦 Publishing package to Git registry\n")
    }
    
    // Check if GitHub repository
    owner, repo, err := parseGitHubURL(c.repoURL)
    if err != nil {
        return nil, fmt.Errorf("only GitHub repositories supported for publishing: %w", err)
    }
    
    // Verify token is provided
    if c.gitToken == "" {
        return nil, fmt.Errorf("GitHub token required for publishing")
    }
    
    // Create GitHub client and ensure fork
    github := NewGitHubClient(c.gitToken, c.verbose)
    fork, err := github.EnsureFork(ctx, owner, repo)
    if err != nil {
        return nil, fmt.Errorf("failed to ensure fork: %w", err)
    }
    
    // Update fork URL in client
    forkURL := fork.CloneURL
    
    // Clone/update fork
    forkRepo, err := c.cloneOrUpdateFork(ctx, forkURL)
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
    _, err = c.createCommit(forkRepo, &manifest)
    if err != nil {
        return nil, fmt.Errorf("failed to create commit: %w", err)
    }
    
    // Push branch
    if err := c.pushBranch(ctx, forkRepo, branchName); err != nil {
        return nil, fmt.Errorf("failed to push branch: %w", err)
    }
    
    // Create pull request
    pr, err := c.createPullRequestForPackage(ctx, branchName, &manifest)
    if err != nil {
        // If PR creation fails, still return success with manual URL
        manualURL := fmt.Sprintf("%s/compare/%s...%s:%s",
            strings.TrimSuffix(c.repoURL, ".git"),
            fork.Parent.DefaultBranch,
            fork.Owner.Login,
            branchName)
        
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
        PRUrl:   pr.HTMLURL,
        Message: fmt.Sprintf("Pull request created successfully: %s", pr.HTMLURL),
    }, nil
}
```

### 7. Add Rate Limit Handling

**File**: `internal/client/github_api.go`

```go
// RateLimitInfo contains GitHub API rate limit information
type RateLimitInfo struct {
    Limit     int       `json:"limit"`
    Remaining int       `json:"remaining"`
    Reset     time.Time `json:"reset"`
}

// GetRateLimit gets current rate limit status
func (g *GitHubClient) GetRateLimit(ctx context.Context) (*RateLimitInfo, error) {
    resp, err := g.makeRequest(ctx, "GET", "/rate_limit", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Core RateLimitInfo `json:"core"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return &result.Core, nil
}

// WaitForRateLimit waits if rate limit is low
func (g *GitHubClient) WaitForRateLimit(ctx context.Context) error {
    info, err := g.GetRateLimit(ctx)
    if err != nil {
        return err
    }
    
    if info.Remaining < 10 {
        waitTime := time.Until(info.Reset)
        if g.verbose {
            fmt.Printf("⏳ Rate limit low (%d remaining). Waiting %v\n", 
                info.Remaining, waitTime)
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
```

## Testing Requirements

### Unit Tests
1. Test GitHub URL parsing
2. Test PR body generation
3. Test rate limit handling
4. Mock GitHub API responses

### Integration Tests
1. Test user authentication
2. Test fork creation
3. Test PR creation
4. Test rate limit checking
5. Test with existing fork

### Manual Testing Checklist
- [ ] Authenticate with GitHub token
- [ ] Create fork of repository
- [ ] Use existing fork
- [ ] Create pull request
- [ ] Handle rate limiting
- [ ] PR contains correct information
- [ ] Works with private repositories

### Cucumber Test Amendments

**File**: `features/github-integration.feature` (new file)

```gherkin
Feature: GitHub API Integration
  Git registry publishing should integrate with GitHub API for automated PR creation

  Background:
    Given I have a clean test environment
    And I have a valid GitHub token with appropriate permissions
    And I have a package ready to publish

  Scenario: Automatic fork creation
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    And I do not have an existing fork
    When I run "rfh publish"
    Then the command should succeed
    And a new fork should be created automatically
    And the output should contain "Fork created"
    And the output should contain a pull request URL

  Scenario: Use existing fork
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    And I already have an existing fork
    When I run "rfh publish"
    Then the command should succeed
    And the existing fork should be used
    And the output should contain "Fork already exists"
    And the output should contain a pull request URL

  Scenario: Automatic pull request creation
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    When I run "rfh publish"
    Then the command should succeed
    And a pull request should be created automatically
    And the PR title should contain the package name and version
    And the PR body should contain package details
    And the PR body should contain SHA256 hash and size
    And the output should contain the PR URL

  Scenario: Pull request format is correct
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    And I have "my-package@2.1.0" ready to publish
    When I run "rfh publish"
    Then the command should succeed
    And the PR title should be "Publish my-package@2.1.0"
    And the PR head should be "test-user:publish/my-package/2.1.0"
    And the PR base should be "main"
    And the PR should allow maintainer modifications

  Scenario: Authentication failure handling
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have an invalid GitHub token
    When I run "rfh publish"
    Then the command should fail
    And the output should contain "authentication failed" or "invalid token"

  Scenario: Rate limiting handling
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    And the GitHub API is rate limited
    When I run "rfh publish"
    Then the command should wait for rate limit reset
    And the output should contain "Rate limit"
    And the command should eventually succeed

  Scenario: Insufficient permissions handling
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have a GitHub token with insufficient permissions
    When I run "rfh publish"
    Then the command should fail
    And the output should contain "insufficient permissions" or "access denied"

  Scenario: PR creation failure fallback
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    And PR creation will fail due to API issues
    When I run "rfh publish"
    Then the command should succeed with warnings
    And the output should contain "Branch pushed"
    And the output should contain a manual PR creation URL
    And the output should contain "Create PR manually"

  Scenario: Private repository support
    Given a Git registry "github-private" with URL "https://github.com/test-org/private-registry"
    And the registry is a private repository
    And I use registry "github-private"
    And I have authenticated to the Git registry
    When I run "rfh publish"
    Then the command should succeed
    And a pull request should be created
    And the fork should be private

  Scenario: User information retrieval
    Given a Git registry "github-test" with URL "https://github.com/test-org/test-registry"
    And I use registry "github-test"
    And I have authenticated to the Git registry
    When I run "rfh publish --verbose"
    Then the command should succeed
    And the output should contain "Authenticated as: test-user"
    And the user information should be used in PR creation

  Scenario: Multiple registries with different tokens
    Given a Git registry "github-org1" with URL "https://github.com/org1/registry"
    And a Git registry "github-org2" with URL "https://github.com/org2/registry"
    And "github-org1" has token "token1"
    And "github-org2" has token "token2"
    When I publish to "github-org1"
    Then the command should use "token1" for authentication
    When I publish to "github-org2"
    Then the command should use "token2" for authentication
```

**File**: `features/step_definitions/github_api_steps.js`

Add new step definitions:
```javascript
Given('I have a valid GitHub token with appropriate permissions', async function () {
  process.env.GITHUB_TOKEN = 'ghp_valid_token_with_permissions';
  // Mock GitHub API to return successful responses
  await this.mockGitHubAPI.setup({
    user: { login: 'test-user', name: 'Test User' },
    permissions: ['repo', 'public_repo']
  });
});

Given('I do not have an existing fork', async function () {
  // Mock GitHub API to return 404 for fork check
  await this.mockGitHubAPI.setupNoFork('test-org', 'test-registry');
});

Given('I already have an existing fork', async function () {
  // Mock existing fork in GitHub API
  await this.mockGitHubAPI.setupExistingFork('test-user', 'test-registry', {
    parent: { full_name: 'test-org/test-registry' }
  });
});

Then('a new fork should be created automatically', function () {
  assert(this.mockGitHubAPI.forkWasCreated(), 'Fork should have been created');
});

Then('the existing fork should be used', function () {
  assert(!this.mockGitHubAPI.forkWasCreated(), 'Should not create new fork');
  const output = this.lastResult.stdout || '';
  assert(output.includes('Fork already exists'), 'Should use existing fork');
});

Then('a pull request should be created automatically', function () {
  assert(this.mockGitHubAPI.prWasCreated(), 'Pull request should have been created');
});

Then('the PR title should contain the package name and version', function () {
  const prData = this.mockGitHubAPI.getCreatedPR();
  assert(prData.title.includes('test-package') && prData.title.includes('1.0.0'),
    `PR title should contain package info: ${prData.title}`);
});

Then('the PR body should contain package details', function () {
  const prData = this.mockGitHubAPI.getCreatedPR();
  assert(prData.body.includes('Package:') && prData.body.includes('test-package'),
    'PR body should contain package details');
});

Then('the PR body should contain SHA256 hash and size', function () {
  const prData = this.mockGitHubAPI.getCreatedPR();
  assert(prData.body.includes('SHA256:') && prData.body.includes('Size:'),
    'PR body should contain hash and size');
});

Then('the output should contain the PR URL', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('pull request created') && output.includes('github.com'),
    'Output should contain PR URL');
});

Then('the PR title should be {string}', function (expectedTitle) {
  const prData = this.mockGitHubAPI.getCreatedPR();
  assert(prData.title === expectedTitle, 
    `Expected PR title "${expectedTitle}", got "${prData.title}"`);
});

Then('the PR head should be {string}', function (expectedHead) {
  const prData = this.mockGitHubAPI.getCreatedPR();
  assert(prData.head === expectedHead,
    `Expected PR head "${expectedHead}", got "${prData.head}"`);
});

Then('the PR base should be {string}', function (expectedBase) {
  const prData = this.mockGitHubAPI.getCreatedPR();
  assert(prData.base === expectedBase,
    `Expected PR base "${expectedBase}", got "${prData.base}"`);
});

Then('the PR should allow maintainer modifications', function () {
  const prData = this.mockGitHubAPI.getCreatedPR();
  assert(prData.maintainer_can_modify === true,
    'PR should allow maintainer modifications');
});

Given('I have an invalid GitHub token', async function () {
  process.env.GITHUB_TOKEN = 'invalid_token';
  await this.mockGitHubAPI.setupInvalidToken();
});

Given('the GitHub API is rate limited', async function () {
  await this.mockGitHubAPI.setupRateLimit({
    remaining: 0,
    reset: Math.floor(Date.now() / 1000) + 60 // Reset in 1 minute
  });
});

Then('the command should wait for rate limit reset', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('Rate limit') && output.includes('Waiting'),
    'Should wait for rate limit reset');
});

Given('I have a GitHub token with insufficient permissions', async function () {
  process.env.GITHUB_TOKEN = 'ghp_limited_token';
  await this.mockGitHubAPI.setup({
    user: { login: 'test-user' },
    permissions: [] // No permissions
  });
});

Given('PR creation will fail due to API issues', async function () {
  await this.mockGitHubAPI.setupPRCreationFailure('API temporarily unavailable');
});

Then('the command should succeed with warnings', function () {
  assert(!this.lastResult.error || this.lastResult.exitCode === 0,
    'Command should succeed despite PR creation failure');
});

Then('the output should contain a manual PR creation URL', function () {
  const output = this.lastResult.stdout || '';
  assert(output.includes('github.com') && output.includes('compare'),
    'Should provide manual PR creation URL');
});

Given('the registry is a private repository', async function () {
  await this.mockGitHubAPI.setupPrivateRepo('test-org', 'private-registry');
});

Then('the fork should be private', function () {
  const forkData = this.mockGitHubAPI.getCreatedFork();
  assert(forkData.private === true, 'Fork should be private');
});

Then('the output should contain {string}', function (expectedText) {
  const output = this.lastResult.stdout || '';
  assert(output.includes(expectedText),
    `Expected "${expectedText}" in output: ${output}`);
});

Then('the user information should be used in PR creation', function () {
  const prData = this.mockGitHubAPI.getCreatedPR();
  assert(prData.body.includes('test-user'), 'PR should contain user info');
});

Given('{string} has token {string}', async function (registryName, token) {
  const config = await this.loadConfig();
  if (!config.registries[registryName]) {
    config.registries[registryName] = {};
  }
  config.registries[registryName].git_token = token;
  await this.saveConfig(config);
});

When('I publish to {string}', async function (registryName) {
  await this.runCommand(`rfh publish --registry ${registryName}`);
});

Then('the command should use {string} for authentication', function (expectedToken) {
  const usedToken = this.mockGitHubAPI.getUsedToken();
  assert(usedToken === expectedToken,
    `Expected token "${expectedToken}", got "${usedToken}"`);
});
```

**File**: `features/support/mock-github-api.js` (new helper file)

```javascript
class MockGitHubAPI {
  constructor() {
    this.reset();
  }
  
  reset() {
    this.forkCreated = false;
    this.prCreated = false;
    this.createdPR = null;
    this.createdFork = null;
    this.usedToken = null;
    this.user = null;
  }
  
  setup(options) {
    this.user = options.user;
    this.permissions = options.permissions || [];
  }
  
  setupNoFork(owner, repo) {
    // Mock API to return 404 for fork check
  }
  
  setupExistingFork(username, repo, forkData) {
    this.createdFork = { ...forkData, owner: { login: username } };
  }
  
  setupInvalidToken() {
    // Mock API to return 401 for requests
  }
  
  setupRateLimit(rateLimitInfo) {
    this.rateLimitInfo = rateLimitInfo;
  }
  
  setupPRCreationFailure(reason) {
    this.prCreationWillFail = reason;
  }
  
  setupPrivateRepo(owner, repo) {
    this.privateRepos = this.privateRepos || [];
    this.privateRepos.push(`${owner}/${repo}`);
  }
  
  forkWasCreated() {
    return this.forkCreated;
  }
  
  prWasCreated() {
    return this.prCreated;
  }
  
  getCreatedPR() {
    return this.createdPR;
  }
  
  getCreatedFork() {
    return this.createdFork;
  }
  
  getUsedToken() {
    return this.usedToken;
  }
}

module.exports = { MockGitHubAPI };
```

## Success Criteria
- Can authenticate with GitHub token
- Automatically creates fork if needed
- Creates well-formatted pull requests
- Handles rate limiting gracefully
- Provides clear error messages
- Returns PR URL for user reference

## Dependencies
- Phase 6: Git Registry Publishing
- GitHub personal access token
- Network access to GitHub API

## Risks
- **Risk**: GitHub API rate limits
  **Mitigation**: Implement rate limit checking and waiting
  
- **Risk**: Token permissions insufficient
  **Mitigation**: Clear error messages about required permissions
  
- **Risk**: API changes
  **Mitigation**: Use stable v3 API, version headers

## Configuration

Add to CLI config:
```toml
[registries.github-registry]
url = "https://github.com/org/registry"
type = "git"
git_token = "ghp_xxxxxxxxxxxx"
```

Environment variables:
```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
export GITHUB_USERNAME=myusername
export GIT_AUTHOR_NAME="Your Name"
export GIT_AUTHOR_EMAIL="your.email@example.com"
```

## Next Steps
This completes the Git registry implementation. Consider future enhancements:
- GitLab API support
- Bitbucket integration
- PR status checking
- Automatic merging for trusted publishers
- Webhook integration for CI/CD