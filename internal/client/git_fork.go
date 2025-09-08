package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
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