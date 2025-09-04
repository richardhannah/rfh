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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	rfhconfig "rulestack/internal/config"
)

// GitClient implements RegistryClient for Git-based registries
type GitClient struct {
	repoURL  string
	gitToken string
	verbose  bool
	cacheDir string
	repo     *git.Repository
	mu       sync.Mutex // Protects repo operations
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
func (c *GitClient) Type() rfhconfig.RegistryType {
	return rfhconfig.RegistryTypeGit
}

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