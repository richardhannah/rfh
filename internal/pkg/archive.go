package pkg

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"rulestack/internal/security"
)

// ArchiveInfo contains information about a created archive
type ArchiveInfo struct {
	Path      string
	SHA256    string
	SizeBytes int64
}

// Pack creates a tar.gz archive from file patterns
func Pack(patterns []string, outputPath string) (*ArchiveInfo, error) {
	// Collect all files matching the patterns
	var files []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to match pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			// Skip directories
			if info, err := os.Stat(match); err != nil || info.IsDir() {
				continue
			}

			// Clean path and avoid duplicates
			cleanPath := filepath.Clean(match)
			if !seen[cleanPath] {
				files = append(files, cleanPath)
				seen[cleanPath] = true
			}
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files matched the specified patterns")
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create archive file: %w", err)
	}
	defer outFile.Close()

	// Create hash writer
	hasher := sha256.New()
	multiWriter := io.MultiWriter(outFile, hasher)

	// Create gzip writer
	gzWriter := gzip.NewWriter(multiWriter)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	var totalSize int64

	// Add each file to the archive
	for _, filePath := range files {
		if err := addFileToArchive(tarWriter, filePath); err != nil {
			return nil, fmt.Errorf("failed to add file %s: %w", filePath, err)
		}

		if info, err := os.Stat(filePath); err == nil {
			totalSize += info.Size()
		}
	}

	// Close writers to flush data before calculating hash
	tarWriter.Close()
	gzWriter.Close()
	outFile.Close()

	// Get final archive info
	info, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat archive: %w", err)
	}

	return &ArchiveInfo{
		Path:      outputPath,
		SHA256:    fmt.Sprintf("%x", hasher.Sum(nil)),
		SizeBytes: info.Size(),
	}, nil
}

// addFileToArchive adds a single file to the tar archive
func addFileToArchive(tarWriter *tar.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create tar header
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	// Use forward slashes in archive
	header.Name = filepath.ToSlash(filePath)

	// Write header
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	// Write file content
	_, err = io.Copy(tarWriter, file)
	return err
}

// Unpack extracts a tar.gz archive to a destination directory with security validation
func Unpack(archivePath string, destDir string) error {
	// First, validate the archive for security
	validator := security.NewPackageValidator(nil)
	if err := validator.ValidateArchive(archivePath, destDir); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	// If validation passes, proceed with extraction
	return UnpackValidated(archivePath, destDir)
}

// UnpackValidated extracts a pre-validated archive (internal use)
func UnpackValidated(archivePath string, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		if err := extractFileSecure(tarReader, header, destDir); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", header.Name, err)
		}
	}

	return nil
}

// extractFileSecure extracts a single file from tar archive with enhanced security
func extractFileSecure(tarReader *tar.Reader, header *tar.Header, destDir string) error {
	// Validate file path (redundant with validator, but defense in depth)
	if err := validateExtractionPath(header.Name, destDir); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, header.Name)

	// Handle directories
	if header.Typeflag == tar.TypeDir {
		return os.MkdirAll(destPath, 0o755)
	}

	// Only handle regular files (symlinks and other types rejected by validator)
	if header.Typeflag != tar.TypeReg {
		return fmt.Errorf("unsupported file type: %c", header.Typeflag)
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	// Create file with safe permissions
	outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Copy file content with size limit (defense in depth)
	_, err = io.CopyN(outFile, tarReader, security.MaxFileSize)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

// validateExtractionPath validates the extraction path for security
func validateExtractionPath(filePath, destDir string) error {
	// Reject absolute paths
	if filepath.IsAbs(filePath) {
		return fmt.Errorf("absolute paths not allowed: %s", filePath)
	}

	// Reject paths with .. segments
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("path traversal attempt: %s", filePath)
	}

	// Ensure the final path is within the destination directory
	destPath := filepath.Join(destDir, filePath)
	cleanDest := filepath.Clean(destPath)
	cleanDestDir := filepath.Clean(destDir)
	
	if !strings.HasPrefix(cleanDest, cleanDestDir) {
		return fmt.Errorf("path escapes destination directory: %s", filePath)
	}

	return nil
}

// extractFile extracts a single file from tar archive (legacy function for compatibility)
func extractFile(tarReader *tar.Reader, header *tar.Header, destDir string) error {
	return extractFileSecure(tarReader, header, destDir)
}

// CalculateSHA256 calculates SHA256 hash of a file
func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}