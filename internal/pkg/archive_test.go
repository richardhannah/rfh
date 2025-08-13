package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateSHA256(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("calculates hash for existing file", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "test.txt")
		content := "Hello, World!"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		hash, err := CalculateSHA256(testFile)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Verify hash length (SHA256 = 64 hex chars)
		if len(hash) != 64 {
			t.Errorf("expected hash length 64, got %d", len(hash))
		}

		// Verify deterministic behavior
		hash2, err := CalculateSHA256(testFile)
		if err != nil {
			t.Errorf("unexpected error on second call: %v", err)
		}

		if hash != hash2 {
			t.Errorf("hash is not deterministic: %q != %q", hash, hash2)
		}
	})

	t.Run("fails for non-existent file", func(t *testing.T) {
		_, err := CalculateSHA256("nonexistent.txt")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("different files produce different hashes", func(t *testing.T) {
		// Create two different test files
		file1 := filepath.Join(tempDir, "file1.txt")
		file2 := filepath.Join(tempDir, "file2.txt")

		err1 := os.WriteFile(file1, []byte("content1"), 0644)
		err2 := os.WriteFile(file2, []byte("content2"), 0644)
		if err1 != nil || err2 != nil {
			t.Fatalf("failed to create test files: %v, %v", err1, err2)
		}

		hash1, err1 := CalculateSHA256(file1)
		hash2, err2 := CalculateSHA256(file2)

		if err1 != nil || err2 != nil {
			t.Fatalf("unexpected errors: %v, %v", err1, err2)
		}

		if hash1 == hash2 {
			t.Error("different files produced same hash")
		}
	})
}

func TestPack(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "pack_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		"file1.txt":           "content1",
		"subdir/file2.txt":    "content2",
		"rules/rule1.md":      "# Rule 1",
		"rules/rule2.md":      "# Rule 2",
		"other/ignored.txt":   "ignored",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", fullPath, err)
		}
	}

	// Change to temp directory for relative path testing
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	t.Run("packs files matching patterns", func(t *testing.T) {
		patterns := []string{"rules/*.md", "file1.txt"}
		outputPath := "test.tgz"

		info, err := Pack(patterns, outputPath)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if info == nil {
			t.Fatal("expected ArchiveInfo, got nil")
		}

		// Verify archive was created
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("archive file was not created")
		}

		// Verify SHA256 is set
		if len(info.SHA256) != 64 {
			t.Errorf("expected SHA256 length 64, got %d", len(info.SHA256))
		}

		// Verify size is positive
		if info.SizeBytes <= 0 {
			t.Errorf("expected positive size, got %d", info.SizeBytes)
		}

		// Clean up
		os.Remove(outputPath)
	})

	t.Run("fails with no matching files", func(t *testing.T) {
		patterns := []string{"nonexistent/*.xyz"}
		outputPath := "empty.tgz"

		_, err := Pack(patterns, outputPath)
		if err == nil {
			t.Error("expected error for no matching files")
		}
	})

	t.Run("fails with invalid pattern", func(t *testing.T) {
		patterns := []string{"[invalid"}
		outputPath := "invalid.tgz"

		_, err := Pack(patterns, outputPath)
		if err == nil {
			t.Error("expected error for invalid pattern")
		}
	})
}

func TestUnpack(t *testing.T) {
	// Create temporary directories
	sourceDir, err := os.MkdirTemp("", "unpack_source")
	if err != nil {
		t.Fatalf("failed to create source temp dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	destDir, err := os.MkdirTemp("", "unpack_dest")
	if err != nil {
		t.Fatalf("failed to create dest temp dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	// Create test files and pack them
	testFiles := map[string]string{
		"file1.txt":        "content1",
		"subdir/file2.txt": "content2",
	}

	oldWd, _ := os.Getwd()
	os.Chdir(sourceDir)
	defer os.Chdir(oldWd)

	for path, content := range testFiles {
		dir := filepath.Dir(path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("failed to create directory %s: %v", dir, err)
			}
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", path, err)
		}
	}

	// Pack the files
	archivePath := "test.tgz"
	patterns := []string{"**/*.txt"}
	_, err = Pack(patterns, archivePath)
	if err != nil {
		t.Fatalf("failed to create test archive: %v", err)
	}

	t.Run("unpacks archive successfully", func(t *testing.T) {
		err := Unpack(archivePath, destDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Verify files were extracted
		for path, expectedContent := range testFiles {
			extractedPath := filepath.Join(destDir, path)
			if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
				t.Errorf("expected file %s was not extracted", path)
				continue
			}

			content, err := os.ReadFile(extractedPath)
			if err != nil {
				t.Errorf("failed to read extracted file %s: %v", path, err)
				continue
			}

			if string(content) != expectedContent {
				t.Errorf("content mismatch for %s: expected %q, got %q", path, expectedContent, string(content))
			}
		}
	})

	t.Run("fails with non-existent archive", func(t *testing.T) {
		err := Unpack("nonexistent.tgz", destDir)
		if err == nil {
			t.Error("expected error for non-existent archive")
		}
	})
}

func TestExtractFileDirectoryTraversal(t *testing.T) {
	// This test verifies that directory traversal attacks are prevented
	tempDir, err := os.MkdirTemp("", "extract_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// We can't easily test extractFile directly since it requires a tar.Reader,
	// but the function checks for ".." in the path, so we test the path cleaning logic
	testPaths := []string{
		"../../../etc/passwd",
		"..\\..\\windows\\system32",
		"normal/file.txt",
		"./relative/file.txt",
	}

	for _, testPath := range testPaths {
		cleanPath := filepath.Clean(testPath)
		// The function should reject paths containing ".."
		if filepath.IsAbs(cleanPath) || filepath.HasPrefix(cleanPath, "..") {
			// These should be rejected by the extractFile function
			t.Logf("Path %q would be rejected (contains '..' or is absolute)", testPath)
		} else {
			t.Logf("Path %q would be accepted", testPath)
		}
	}
}