package client

import (
	"context"
	"fmt"
	"strings"
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

// WaitForRateLimit waits if rate limit is low
func (g *GitHubClient) WaitForRateLimit(ctx context.Context) error {
	rateLimit, _, err := g.client.RateLimits(ctx)
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
	rateLimit, _, err := g.client.RateLimits(ctx)
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