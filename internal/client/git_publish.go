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
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
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
	_, err := repo.Head()
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
		return "", fmt.Errorf("failed to create branch: %w", err)
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

// addPackageFiles adds package files to the repository
func (c *GitClient) addPackageFiles(repo *git.Repository, manifestPath, archivePath string) error {
	// Parse manifest to get package info
	manifestData, err := os.ReadFile(manifestPath)
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
	if err := os.WriteFile(manifestDest, updatedManifest, 0644); err != nil {
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
	_, err = w.Add("packages/" + manifest.Name)
	if err != nil {
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
	if data, err := os.ReadFile(metadataPath); err == nil {
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
	return os.WriteFile(metadataPath, data, 0644)
}

// updateRegistryIndex updates the main registry index
func (c *GitClient) updateRegistryIndex(repo *git.Repository, manifest *GitManifest) error {
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	indexPath := filepath.Join(w.Filesystem.Root(), "index.json")

	var index GitRegistryIndex

	// Load existing index
	if data, err := os.ReadFile(indexPath); err == nil {
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
	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	// Stage index changes
	_, err = w.Add("index.json")
	if err != nil {
		return fmt.Errorf("failed to stage index: %w", err)
	}

	return nil
}

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