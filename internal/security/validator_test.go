package security

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a test archive
func createTestArchive(files map[string][]byte) (string, error) {
	tmpFile, err := os.CreateTemp("", "test-archive-*.tgz")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	gzWriter := gzip.NewWriter(tmpFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for filename, content := range files {
		header := &tar.Header{
			Name:     filename,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
			Mode:     0644,
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return "", err
		}

		if _, err := tarWriter.Write(content); err != nil {
			return "", err
		}
	}

	return tmpFile.Name(), nil
}

func TestPackageValidator_ValidFiles(t *testing.T) {
	validator := NewPackageValidator(nil)

	// Create a valid archive
	files := map[string][]byte{
		"rules/test.md":    []byte("# Test Rule\n\nThis is a valid markdown file."),
		"docs/readme.txt":  []byte("This is a readme file."),
		"config/data.json": []byte(`{"test": "value"}`),
	}

	archivePath, err := createTestArchive(files)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}
	defer os.Remove(archivePath)

	tmpDir, err := os.MkdirTemp("", "test-extract-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Should pass validation
	err = validator.ValidateArchive(archivePath, tmpDir)
	if err != nil {
		t.Errorf("Valid archive failed validation: %v", err)
	}
}

func TestPackageValidator_PathTraversal(t *testing.T) {
	validator := NewPackageValidator(nil)

	testCases := []struct {
		name     string
		filename string
	}{
		{"Relative path traversal", "../../../etc/passwd"},
		{"Absolute path", "/etc/passwd"},
		{"Current dir traversal", "./../../secret.txt"},
		{"Windows path traversal", "..\\..\\windows\\system32\\config\\sam"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string][]byte{
				tc.filename: []byte("malicious content"),
			}

			archivePath, err := createTestArchive(files)
			if err != nil {
				t.Fatalf("Failed to create test archive: %v", err)
			}
			defer os.Remove(archivePath)

			tmpDir, err := os.MkdirTemp("", "test-extract-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			err = validator.ValidateArchive(archivePath, tmpDir)
			if err == nil {
				t.Errorf("Path traversal attack was not detected for: %s", tc.filename)
			}
		})
	}
}

func TestPackageValidator_DisallowedFileTypes(t *testing.T) {
	validator := NewPackageValidator(nil)

	testCases := []string{
		"script.sh",
		"executable.exe",
		"library.dll",
		"binary.so",
		"image.png",
		"script.py",
		"batch.bat",
		"powershell.ps1",
	}

	for _, filename := range testCases {
		t.Run(filename, func(t *testing.T) {
			files := map[string][]byte{
				filename: []byte("potentially dangerous content"),
			}

			archivePath, err := createTestArchive(files)
			if err != nil {
				t.Fatalf("Failed to create test archive: %v", err)
			}
			defer os.Remove(archivePath)

			tmpDir, err := os.MkdirTemp("", "test-extract-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			err = validator.ValidateArchive(archivePath, tmpDir)
			if err == nil {
				t.Errorf("Disallowed file type was not rejected: %s", filename)
			}
		})
	}
}

func TestPackageValidator_ExecutableHeaders(t *testing.T) {
	validator := NewPackageValidator(nil)

	testCases := []struct {
		name    string
		content []byte
	}{
		{"ELF executable", []byte{0x7F, 0x45, 0x4C, 0x46, 0x01, 0x01, 0x01, 0x00}},
		{"PE executable", []byte{0x4D, 0x5A, 0x90, 0x00, 0x03, 0x00, 0x00, 0x00}},
		{"Shebang script", []byte("#!/bin/bash\necho 'dangerous'")},
		{"Java class", []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x00, 0x00, 0x34}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string][]byte{
				"file.txt": tc.content, // Disguised as text file
			}

			archivePath, err := createTestArchive(files)
			if err != nil {
				t.Fatalf("Failed to create test archive: %v", err)
			}
			defer os.Remove(archivePath)

			tmpDir, err := os.MkdirTemp("", "test-extract-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			err = validator.ValidateArchive(archivePath, tmpDir)
			if err == nil {
				t.Errorf("Executable content was not detected: %s", tc.name)
			}
		})
	}
}

func TestPackageValidator_MaliciousMarkdown(t *testing.T) {
	validator := NewPackageValidator(nil)

	testCases := []struct {
		name    string
		content string
	}{
		{"Script tag", "<script>alert('xss')</script>"},
		{"JavaScript URL", "[Click me](javascript:alert('xss'))"},
		{"Iframe", "<iframe src='http://malicious.com'></iframe>"},
		{"Object tag", "<object data='malicious.swf'></object>"},
		{"Onload handler", "<img onload='alert(1)' src='x'>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string][]byte{
				"rules/malicious.md": []byte("# Test\n\n" + tc.content),
			}

			archivePath, err := createTestArchive(files)
			if err != nil {
				t.Fatalf("Failed to create test archive: %v", err)
			}
			defer os.Remove(archivePath)

			tmpDir, err := os.MkdirTemp("", "test-extract-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			err = validator.ValidateArchive(archivePath, tmpDir)
			if err == nil {
				t.Errorf("Malicious markdown was not detected: %s", tc.name)
			}
		})
	}
}

func TestPackageValidator_SizeLimits(t *testing.T) {
	validator := NewPackageValidator(nil)

	// Test file too large
	largeContent := bytes.Repeat([]byte("A"), int(MaxFileSize+1))
	files := map[string][]byte{
		"large.md": largeContent,
	}

	archivePath, err := createTestArchive(files)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}
	defer os.Remove(archivePath)

	tmpDir, err := os.MkdirTemp("", "test-extract-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = validator.ValidateArchive(archivePath, tmpDir)
	if err == nil {
		t.Error("Oversized file was not rejected")
	}
}

func TestPackageValidator_NulBytes(t *testing.T) {
	validator := NewPackageValidator(nil)

	files := map[string][]byte{
		"file.txt": []byte("normal content\x00hidden content"),
	}

	archivePath, err := createTestArchive(files)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}
	defer os.Remove(archivePath)

	tmpDir, err := os.MkdirTemp("", "test-extract-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = validator.ValidateArchive(archivePath, tmpDir)
	if err == nil {
		t.Error("File with NUL bytes was not rejected")
	}
}

func TestPackageValidator_InvalidUTF8(t *testing.T) {
	validator := NewPackageValidator(nil)

	// Invalid UTF-8 sequence
	invalidUTF8 := []byte{0xFF, 0xFE, 0xFD} // Invalid UTF-8
	files := map[string][]byte{
		"file.txt": append([]byte("Some text "), invalidUTF8...),
	}

	archivePath, err := createTestArchive(files)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}
	defer os.Remove(archivePath)

	tmpDir, err := os.MkdirTemp("", "test-extract-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = validator.ValidateArchive(archivePath, tmpDir)
	if err == nil {
		t.Error("File with invalid UTF-8 was not rejected")
	}
}

func TestPackageValidator_TooManyFiles(t *testing.T) {
	validator := NewPackageValidator(nil)

	files := make(map[string][]byte)
	for i := 0; i < MaxFilesPerArchive+1; i++ {
		files[filepath.Join("files", fmt.Sprintf("file%d.md", i))] = []byte("content")
	}

	archivePath, err := createTestArchive(files)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}
	defer os.Remove(archivePath)

	tmpDir, err := os.MkdirTemp("", "test-extract-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = validator.ValidateArchive(archivePath, tmpDir)
	if err == nil {
		t.Error("Archive with too many files was not rejected")
	}
}