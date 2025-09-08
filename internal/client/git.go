package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

// Interface implementations - Phase 5: Git Registry Discovery

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
	manifestData, _ := os.ReadFile(manifestPath)
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

// Phase 5 Helper Methods

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
	data, err := os.ReadFile(indexPath)
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

	data, err := os.ReadFile(metadataPath)
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

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest GitManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
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

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Write to destination
	return os.WriteFile(dst, data, 0644)
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
	entries, err := os.ReadDir(packagesDir)
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