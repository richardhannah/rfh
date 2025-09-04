package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

// Test to verify go-git library compatibility with Gitea
// This test checks if we can use go-git to perform Git operations against our local Gitea instance
// which should work better than the GitHub API approach since it uses native Git protocols
// DISABLED - compatibility confirmed, keeping as reference
func TestGoGitPackageWithGitea_DISABLED(t *testing.T) {
	t.Skip("Compatibility test disabled - go-git works with Gitea and GitHub PATs")
	const (
		giteaURL     = "http://localhost:3000"
		token        = "6f0867490bf60242d40f052c9a07b247f9816e39"
		testUser     = "richard"
		testRepo     = "test-go-git-compat"
		testBranch   = "feature-branch"
		giteaRepoURL = "http://localhost:3000/richard/test-go-git-compat.git"
	)

	// Skip test if Gitea is not running
	resp, err := http.Get(giteaURL + "/api/v1/version")
	if err != nil {
		t.Skipf("Gitea not running at %s: %v", giteaURL, err)
	}
	resp.Body.Close()

	// Create authentication for go-git
	auth := &githttp.BasicAuth{
		Username: "richard", // Can be any string for token auth
		Password: token,
	}

	t.Run("Clone Repository", func(t *testing.T) {
		// Create a temporary directory for cloning
		tempDir, err := ioutil.TempDir("", "gitea-clone-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// First, we need a repository to clone. Let's check if one exists or skip this test
		// Try to clone a repository (this might fail if no repo exists, which is expected)
		repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
			URL:  giteaRepoURL,
			Auth: auth,
		})

		if err != nil {
			// If clone fails, it might be because the repo doesn't exist yet
			// Let's try to create one by initializing and pushing
			t.Logf("Clone failed (expected if repo doesn't exist): %v", err)
			t.Skip("Skipping clone test - repository may not exist")
			return
		}

		// If we get here, the clone was successful
		worktree, err := repo.Worktree()
		if err != nil {
			t.Fatalf("Failed to get worktree: %v", err)
		}

		// Check if we can read files
		files, err := ioutil.ReadDir(worktree.Filesystem.Root())
		if err != nil {
			t.Fatalf("Failed to read repository files: %v", err)
		}

		t.Logf("✅ Successfully cloned repository with %d files/directories", len(files))
		for _, file := range files {
			t.Logf("  - %s", file.Name())
		}
	})

	t.Run("Memory Clone and Inspect", func(t *testing.T) {
		// Try to clone to memory to test basic connectivity
		repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL:  giteaRepoURL,
			Auth: auth,
		})

		if err != nil {
			t.Logf("Memory clone failed (expected if repo doesn't exist): %v", err)
			t.Skip("Skipping memory clone test - repository may not exist")
			return
		}

		// Get repository information
		head, err := repo.Head()
		if err != nil {
			t.Fatalf("Failed to get HEAD: %v", err)
		}

		t.Logf("✅ Successfully cloned to memory")
		t.Logf("  - HEAD: %s", head.Name().Short())
		t.Logf("  - Commit: %s", head.Hash().String())

		// List branches
		branches, err := repo.Branches()
		if err != nil {
			t.Fatalf("Failed to list branches: %v", err)
		}

		branchCount := 0
		err = branches.ForEach(func(branch *plumbing.Reference) error {
			branchCount++
			t.Logf("  - Branch: %s", branch.Name().Short())
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to iterate branches: %v", err)
		}

		t.Logf("✅ Found %d branches", branchCount)
	})

	t.Run("Initialize and Push Repository", func(t *testing.T) {
		// Create a temporary directory for our test repository
		tempDir, err := ioutil.TempDir("", "gitea-init-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Initialize a new repository
		repo, err := git.PlainInit(tempDir, false)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		t.Logf("✅ Successfully initialized local repository")

		// Get the worktree
		worktree, err := repo.Worktree()
		if err != nil {
			t.Fatalf("Failed to get worktree: %v", err)
		}

		// Create a test file
		testFile := filepath.Join(tempDir, "README.md")
		err = ioutil.WriteFile(testFile, []byte("# Test Repository\n\nThis is a test repository created with go-git"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Add the file to staging
		_, err = worktree.Add("README.md")
		if err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		// Commit the file
		commit, err := worktree.Commit("Initial commit via go-git", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Go-Git Test",
				Email: "test@localhost",
				When:  time.Now(),
			},
		})
		if err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		t.Logf("✅ Successfully created commit: %s", commit.String())

		// Note: We're not pushing to Gitea here because we'd need to create the repository first
		// This test verifies that go-git can perform local Git operations correctly
	})

	t.Run("Branch Operations", func(t *testing.T) {
		// Create a temporary directory for branch testing
		tempDir, err := ioutil.TempDir("", "gitea-branch-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Initialize repository
		repo, err := git.PlainInit(tempDir, false)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		worktree, err := repo.Worktree()
		if err != nil {
			t.Fatalf("Failed to get worktree: %v", err)
		}

		// Create initial commit
		testFile := filepath.Join(tempDir, "initial.txt")
		err = ioutil.WriteFile(testFile, []byte("Initial content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create initial file: %v", err)
		}

		_, err = worktree.Add("initial.txt")
		if err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		initialCommit, err := worktree.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Go-Git Test",
				Email: "test@localhost", 
				When:  time.Now(),
			},
		})
		if err != nil {
			t.Fatalf("Failed to create initial commit: %v", err)
		}

		t.Logf("✅ Created initial commit: %s", initialCommit.String())

		// Create a new branch
		branchRef := plumbing.NewBranchReferenceName(testBranch)
		ref := plumbing.NewHashReference(branchRef, initialCommit)
		
		err = repo.Storer.SetReference(ref)
		if err != nil {
			t.Fatalf("Failed to create branch: %v", err)
		}

		t.Logf("✅ Successfully created branch: %s", testBranch)

		// Checkout the new branch
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
		})
		if err != nil {
			t.Fatalf("Failed to checkout branch: %v", err)
		}

		t.Logf("✅ Successfully checked out branch: %s", testBranch)

		// Make changes on the new branch
		branchFile := filepath.Join(tempDir, "branch-file.txt")
		err = ioutil.WriteFile(branchFile, []byte("Content from feature branch"), 0644)
		if err != nil {
			t.Fatalf("Failed to create branch file: %v", err)
		}

		_, err = worktree.Add("branch-file.txt")
		if err != nil {
			t.Fatalf("Failed to add branch file: %v", err)
		}

		branchCommit, err := worktree.Commit("Add feature from branch", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Go-Git Test",
				Email: "test@localhost",
				When:  time.Now(),
			},
		})
		if err != nil {
			t.Fatalf("Failed to commit to branch: %v", err)
		}

		t.Logf("✅ Successfully committed to branch: %s", branchCommit.String())

		// List all branches
		refs, err := repo.References()
		if err != nil {
			t.Fatalf("Failed to get references: %v", err)
		}

		branchCount := 0
		err = refs.ForEach(func(ref *plumbing.Reference) error {
			if ref.Name().IsBranch() {
				branchCount++
				t.Logf("  - Branch: %s → %s", ref.Name().Short(), ref.Hash().String())
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to iterate references: %v", err)
		}

		t.Logf("✅ Found %d branches total", branchCount)
	})

}

// Benchmark test to compare Git operation performance 
// DISABLED - compatibility confirmed, keeping as reference
func BenchmarkGoGitOperations_DISABLED(b *testing.B) {
	b.Skip("Benchmark disabled - compatibility confirmed")
	const (
		giteaURL = "http://localhost:3000"
		token    = "6f0867490bf60242d40f052c9a07b247f9816e39"
	)

	// Skip if Gitea is not running
	resp, err := http.Get(giteaURL + "/api/v1/version")
	if err != nil {
		b.Skipf("Gitea not running at %s: %v", giteaURL, err)
	}
	resp.Body.Close()

	b.Run("InitializeRepository", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tempDir, err := ioutil.TempDir("", "bench-init-*")
			if err != nil {
				b.Fatalf("Failed to create temp dir: %v", err)
			}
			
			_, err = git.PlainInit(tempDir, false)
			if err != nil {
				b.Fatalf("Failed to initialize repository: %v", err)
			}
			
			os.RemoveAll(tempDir)
		}
	})

	b.Run("CreateCommit", func(b *testing.B) {
		// Setup: Create a repository once
		tempDir, err := ioutil.TempDir("", "bench-commit-*")
		if err != nil {
			b.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo, err := git.PlainInit(tempDir, false)
		if err != nil {
			b.Fatalf("Failed to initialize repository: %v", err)
		}

		worktree, err := repo.Worktree()
		if err != nil {
			b.Fatalf("Failed to get worktree: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create a unique file for each benchmark iteration
			fileName := filepath.Join(tempDir, "bench-file-"+fmt.Sprintf("%d", i)+".txt")
			err = ioutil.WriteFile(fileName, []byte("Benchmark content"), 0644)
			if err != nil {
				b.Fatalf("Failed to create file: %v", err)
			}

			_, err = worktree.Add(".")
			if err != nil {
				b.Fatalf("Failed to add files: %v", err)
			}

			_, err = worktree.Commit("Benchmark commit", &git.CommitOptions{
				Author: &object.Signature{
					Name:  "Benchmark",
					Email: "bench@localhost",
					When:  time.Now(),
				},
			})
			if err != nil {
				b.Fatalf("Failed to commit: %v", err)
			}
		}
	})
}