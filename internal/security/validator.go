package security

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
)

const (
	// Security limits
	MaxFileSize        = 1024 * 1024      // 1MB per file
	MaxTotalSize       = 10 * 1024 * 1024 // 10MB total uncompressed
	MaxFilesPerArchive = 100              // Maximum number of files
)

// SecurityConfig contains security validation settings
type SecurityConfig struct {
	AllowedExtensions []string
	MaxFileSize       int64
	MaxTotalSize      int64
	MaxFiles          int
	RequireUTF8       bool
	SanitizeMarkdown  bool
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedExtensions: []string{".md", ".txt", ".json", ".mdc"},
		MaxFileSize:       MaxFileSize,
		MaxTotalSize:      MaxTotalSize,
		MaxFiles:          MaxFilesPerArchive,
		RequireUTF8:       true,
		SanitizeMarkdown:  true,
	}
}

// PackageValidator handles security validation of packages
type PackageValidator struct {
	config *SecurityConfig
	policy *bluemonday.Policy
}

// NewPackageValidator creates a new package validator
func NewPackageValidator(config *SecurityConfig) *PackageValidator {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	// Create a strict markdown policy - only allow safe markdown elements
	policy := bluemonday.NewPolicy()

	// Allow basic markdown formatting
	policy.AllowElements("h1", "h2", "h3", "h4", "h5", "h6", "p", "br", "hr")
	policy.AllowElements("strong", "b", "em", "i", "code", "pre", "blockquote")
	policy.AllowElements("ul", "ol", "li", "dl", "dt", "dd")
	policy.AllowElements("table", "thead", "tbody", "tr", "th", "td")

	// Allow links but sanitize them
	policy.AllowAttrs("href").OnElements("a")
	policy.RequireNoReferrerOnLinks(true)
	policy.RequireNoFollowOnLinks(true)
	policy.AddTargetBlankToFullyQualifiedLinks(true)

	// Forbid dangerous elements - no scripts, objects, embeds, etc.
	// This is already the default with a strict policy

	return &PackageValidator{
		config: config,
		policy: policy,
	}
}

// ValidateArchive validates the security of a package archive
func (v *PackageValidator) ValidateArchive(archivePath, extractDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var totalSize int64
	var fileCount int

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		fileCount++
		if fileCount > v.config.MaxFiles {
			return fmt.Errorf("archive contains too many files (max %d)", v.config.MaxFiles)
		}

		// Validate file path security
		if err := v.validateFilePath(header.Name, extractDir); err != nil {
			return fmt.Errorf("unsafe file path '%s': %w", header.Name, err)
		}

		// Validate file type
		if err := v.validateFileType(header.Name); err != nil {
			return fmt.Errorf("invalid file type '%s': %w", header.Name, err)
		}

		// Check file size
		if header.Size > v.config.MaxFileSize {
			return fmt.Errorf("file '%s' too large (%d bytes, max %d)",
				header.Name, header.Size, v.config.MaxFileSize)
		}

		totalSize += header.Size
		if totalSize > v.config.MaxTotalSize {
			return fmt.Errorf("archive too large (%d bytes, max %d)",
				totalSize, v.config.MaxTotalSize)
		}

		// Validate file content for regular files
		if header.Typeflag == tar.TypeReg {
			if err := v.validateFileContent(tarReader, header); err != nil {
				return fmt.Errorf("invalid content in '%s': %w", header.Name, err)
			}
		}

		// Reject symlinks and other special file types
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			return fmt.Errorf("unsupported file type for '%s': %c", header.Name, header.Typeflag)
		}
	}

	return nil
}

// validateFilePath checks for path traversal and other path-based attacks
func (v *PackageValidator) validateFilePath(filePath, extractDir string) error {
	// Reject absolute paths (check both Unix and Windows style)
	if filepath.IsAbs(filePath) || strings.HasPrefix(filePath, "/") || strings.HasPrefix(filePath, "\\") {
		return fmt.Errorf("absolute paths not allowed")
	}

	// Reject paths with .. segments (path traversal)
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("path traversal attempt detected")
	}

	// Ensure the final path is within the extract directory
	fullPath := filepath.Join(extractDir, filePath)
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(extractDir)) {
		return fmt.Errorf("path escapes extraction directory")
	}

	// Reject paths with control characters or other dangerous characters
	for _, r := range filePath {
		if r < 32 || r == 127 { // Control characters
			return fmt.Errorf("control characters in path not allowed")
		}
	}

	return nil
}

// validateFileType checks if the file extension is allowed
func (v *PackageValidator) validateFileType(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Allow directories (no extension)
	if ext == "" {
		return nil
	}

	for _, allowed := range v.config.AllowedExtensions {
		if ext == allowed {
			return nil
		}
	}

	return fmt.Errorf("file extension '%s' not allowed", ext)
}

// validateFileContent validates the content of a file
func (v *PackageValidator) validateFileContent(reader io.Reader, header *tar.Header) error {
	// Read file content
	content, err := io.ReadAll(io.LimitReader(reader, v.config.MaxFileSize))
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	// Check for NUL bytes
	if bytes.Contains(content, []byte{0}) {
		return fmt.Errorf("file contains NUL bytes")
	}

	// Check for executable headers
	if err := v.checkExecutableHeaders(content, header.Name); err != nil {
		return err
	}

	// Validate UTF-8 encoding for text files
	if v.config.RequireUTF8 && isTextFile(header.Name) {
		if !utf8.Valid(content) {
			return fmt.Errorf("file is not valid UTF-8")
		}
	}

	// Sanitize markdown content
	if v.config.SanitizeMarkdown && strings.HasSuffix(strings.ToLower(header.Name), ".md") {
		if err := v.validateMarkdownContent(content); err != nil {
			return fmt.Errorf("markdown validation failed: %w", err)
		}
	}

	return nil
}

// checkExecutableHeaders checks for executable file headers
func (v *PackageValidator) checkExecutableHeaders(content []byte, filename string) error {
	if len(content) < 4 {
		return nil
	}

	// Check for common executable signatures
	signatures := map[string][]byte{
		"ELF":       {0x7F, 0x45, 0x4C, 0x46}, // ELF executables
		"PE":        {0x4D, 0x5A},             // PE executables (MZ header)
		"Mach-O 32": {0xFE, 0xED, 0xFA, 0xCE}, // Mach-O 32-bit
		"Mach-O 64": {0xFE, 0xED, 0xFA, 0xCF}, // Mach-O 64-bit
		"Java":      {0xCA, 0xFE, 0xBA, 0xBE}, // Java class files
		"Shebang":   {0x23, 0x21},             // #! scripts
	}

	for sigName, sig := range signatures {
		if len(content) >= len(sig) && bytes.HasPrefix(content, sig) {
			return fmt.Errorf("executable file detected (%s signature)", sigName)
		}
	}

	// Check for script extensions
	ext := strings.ToLower(filepath.Ext(filename))
	scriptExts := []string{".sh", ".bat", ".cmd", ".ps1", ".py", ".rb", ".pl", ".js", ".exe", ".dll", ".so", ".dylib"}
	for _, scriptExt := range scriptExts {
		if ext == scriptExt {
			return fmt.Errorf("executable/script file extension not allowed: %s", ext)
		}
	}

	return nil
}

// validateMarkdownContent validates markdown content using bluemonday
func (v *PackageValidator) validateMarkdownContent(content []byte) error {
	// Check if the content becomes significantly different after sanitization
	original := string(content)
	sanitized := v.policy.Sanitize(original)

	// If the sanitized version is very different, it likely contained dangerous content
	originalLines := strings.Split(original, "\n")
	sanitizedLines := strings.Split(sanitized, "\n")

	// Allow some difference due to HTML cleanup, but reject major changes
	if len(sanitizedLines) < len(originalLines)/2 {
		return fmt.Errorf("markdown content contains potentially dangerous elements")
	}

	// Check for suspicious patterns in the original content
	suspiciousPatterns := []string{
		"<script",
		"<iframe",
		"<object",
		"<embed",
		"<applet",
		"javascript:",
		"data:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
	}

	lowerContent := strings.ToLower(original)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerContent, pattern) {
			return fmt.Errorf("markdown content contains suspicious pattern: %s", pattern)
		}
	}

	return nil
}

// isTextFile determines if a file should be treated as text based on extension
func isTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	textExts := []string{".md", ".txt", ".json", ".yaml", ".yml", ".toml", ".ini", ".mdc"}

	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}

	return false
}
