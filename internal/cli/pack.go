package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"rulestack/internal/manifest"
	"rulestack/internal/pkg"
)

var (
	outputPath   string
	manifestPath string
	fileOverride string  // Single file override
)

// packCmd represents the pack command
var packCmd = &cobra.Command{
	Use:   "pack [project-path]",
	Short: "Pack ruleset files into a distributable archive",
	Long: `Creates a tar.gz archive containing all files specified in the manifest.
The archive is ready for publishing to a registry.

The pack command:
1. Reads rulestack.json manifest (or custom manifest with --manifest)
2. Collects all files matching the patterns in 'files' array (or --file override)
3. Creates a compressed archive
4. Calculates SHA256 hash for integrity

Auto-manifest creation:
- If no manifest exists and --file is specified, creates auto-manifest
- Package name derived from file pattern (e.g., "security-rules" from "security-rules.md")
- Version automatically set to "1.0.0"

Examples:
  rfh pack                                    # Pack current directory
  rfh pack /path/to/project                   # Pack specific project directory
  rfh pack security-rules.md                 # Pack single file (auto-creates manifest)
  rfh pack --manifest custom-rules.json      # Use custom manifest file
  rfh pack --file "rules/security-rules.md"  # Override file to pack (alternative syntax)
  rfh pack --output /tmp/my-package.tgz      # Specify output location`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectPath string
		if len(args) > 0 {
			originalArg := args[0]
			
			// Try multiple path interpretations to handle Windows path quirks
			pathsToTry := []string{
				originalArg,
				filepath.Clean(originalArg),
			}
			
			// If the path starts with a dot and doesn't exist as-is, try without the leading dot
			// This handles cases where shell converts .\filename to .filename
			if strings.HasPrefix(originalArg, ".") && !strings.HasPrefix(originalArg, "./") && !strings.HasPrefix(originalArg, ".\\") {
				// This is likely a file like ".filename" that should be "filename"
				stripped := originalArg[1:]
				pathsToTry = append(pathsToTry, stripped)
			}
			
			var foundPath string
			var foundFile bool
			
			for _, testPath := range pathsToTry {
				if fileInfo, err := os.Stat(testPath); err == nil && !fileInfo.IsDir() {
					foundPath = testPath
					foundFile = true
					break
				}
			}
			
			if foundFile {
				if strings.HasSuffix(strings.ToLower(foundPath), ".md") || 
				   strings.HasSuffix(strings.ToLower(foundPath), ".txt") {
					// User provided a file directly - treat it as --file flag
					if fileOverride != "" {
						return fmt.Errorf("cannot specify both a file argument (%s) and --file flag (%s)", originalArg, fileOverride)
					}
					fileOverride = foundPath
					// Don't set projectPath, use current directory
					projectPath = ""
				} else {
					projectPath = originalArg
				}
			} else {
				// Argument doesn't exist or is a directory
				projectPath = originalArg
			}
		}
		return runPack(projectPath)
	},
}

func runPack(projectPath string) error {
	
	// Determine working directory
	workDir := "."
	if projectPath != "" {
		workDir = projectPath
		if verbose {
			fmt.Printf("📁 Working directory: %s\n", workDir)
		}
	}

	// Determine manifest path
	manifestFile := "rulestack.json"
	if manifestPath != "" {
		manifestFile = manifestPath
	} else if projectPath != "" {
		// If project path is specified, look for manifest in that directory
		manifestFile = filepath.Join(workDir, "rulestack.json")
	}

	if verbose {
		fmt.Printf("📄 Manifest file: %s\n", manifestFile)
	}

	// Load manifest or create one if it doesn't exist and files are specified
	var m *manifest.Manifest
	var err error
	
	// Check if we should create auto-manifest (file specified but no manifest exists)
	if fileOverride != "" {
		if _, statErr := os.Stat(manifestFile); os.IsNotExist(statErr) {
			if verbose {
				fmt.Printf("📝 No manifest found, creating auto-manifest from file\n")
			}
			m = createAutoManifestSingle(fileOverride)
		} else {
			// Manifest exists, load it (file will override later)
			m, err = manifest.Load(manifestFile)
			if err != nil {
				return fmt.Errorf("failed to load manifest: %w", err)
			}
		}
	} else {
		// No file override, must load manifest
		m, err = manifest.Load(manifestFile)
		if err != nil {
			return fmt.Errorf("failed to load manifest: %w", err)
		}
	}

	// Determine files to pack
	filesToPack := m.Files
	if fileOverride != "" {
		// Use single file override
		filesToPack = []string{fileOverride}
		if verbose {
			fmt.Printf("🔄 File overridden from command line: %s\n", fileOverride)
		}
	}

	// Change to working directory if specified
	var originalDir string
	if projectPath != "" {
		originalDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		
		if err := os.Chdir(workDir); err != nil {
			return fmt.Errorf("failed to change to project directory %s: %w", workDir, err)
		}
		defer func() {
			if originalDir != "" {
				os.Chdir(originalDir)
			}
		}()
	}

	// Determine output path
	output := outputPath
	if output == "" {
		// Remove @ and / from package name for filename
		safeName := m.GetPackageName()
		output = fmt.Sprintf("%s-%s.tgz", safeName, m.Version)
		
		// If we changed directories, put output back in original location
		if originalDir != "" {
			output = filepath.Join(originalDir, output)
		}
	}

	if verbose {
		fmt.Printf("📦 Packing %s v%s\n", m.Name, m.Version)
		fmt.Printf("🎯 Targets: %v\n", m.Targets)
		fmt.Printf("🏷️  Tags: %v\n", m.Tags)
		fmt.Printf("📄 File patterns: %v\n", filesToPack)
	}

	// Pack files
	info, err := pkg.Pack(filesToPack, output)
	if err != nil {
		return fmt.Errorf("failed to pack files: %w", err)
	}

	fmt.Printf("✅ Successfully packed %s\n", m.Name)
	fmt.Printf("📦 Archive: %s\n", info.Path)
	fmt.Printf("📏 Size: %d bytes\n", info.SizeBytes)
	fmt.Printf("🔒 SHA256: %s\n", info.SHA256)

	if verbose {
		// List files that were included
		fmt.Printf("\n📋 Files included:\n")
		// This would require modifying Pack to return file list
		// For now, just show the patterns
		for _, pattern := range m.Files {
			fmt.Printf("   - %s\n", pattern)
		}
	}

	return nil
}

// createAutoManifestSingle creates a manifest from a single file when no manifest exists
func createAutoManifestSingle(filePath string) *manifest.Manifest {
	// Extract package name from single file
	packageName := extractPackageNameFromPattern(filePath)
	
	// Fallback name if extraction fails
	if packageName == "" {
		packageName = "auto-generated-package"
	}
	
	return &manifest.Manifest{
		Name:        packageName,
		Version:     "1.0.0",
		Description: fmt.Sprintf("Auto-generated package from file: %s", filePath),
		Files:       []string{filePath},
	}
}

// createAutoManifest creates a manifest from file patterns when no manifest exists (legacy - for future multi-file support)
func createAutoManifest(filePatterns []string) *manifest.Manifest {
	// Parse file patterns to extract package name
	var packageName string
	if len(filePatterns) > 0 {
		// Take the first file pattern and try to extract a meaningful name
		firstPattern := filePatterns[0]
		packageName = extractPackageNameFromPattern(firstPattern)
	}
	
	// Fallback name if extraction fails
	if packageName == "" {
		packageName = "auto-generated-package"
	}
	
	// Flatten and parse all file patterns
	var allFiles []string
	for _, pattern := range filePatterns {
		// Split comma-separated patterns
		parts := strings.Split(pattern, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				allFiles = append(allFiles, part)
			}
		}
	}
	
	return &manifest.Manifest{
		Name:        packageName,
		Version:     "1.0.0",
		Description: fmt.Sprintf("Auto-generated package from files: %v", allFiles),
		Files:       allFiles,
	}
}

// extractPackageNameFromPattern extracts a package name from a file pattern
func extractPackageNameFromPattern(pattern string) string {
	// Remove common prefixes and suffixes
	name := pattern
	
	// Remove file extensions
	name = strings.TrimSuffix(name, ".md")
	name = strings.TrimSuffix(name, ".txt")
	name = strings.TrimSuffix(name, ".*")
	
	// Remove glob patterns
	name = strings.ReplaceAll(name, "*", "")
	name = strings.ReplaceAll(name, "**", "")
	
	// Remove directory separators and take the last meaningful part
	parts := strings.FieldsFunc(name, func(c rune) bool {
		return c == '/' || c == '\\'
	})
	
	// Find the most meaningful part
	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		
		if part != "" && part != "." && part != ".." {
			// Clean up the name to be package-name compliant
			part = strings.ReplaceAll(part, "_", "-")
			part = strings.ToLower(part)
			
			// Remove any remaining invalid characters
			var cleaned strings.Builder
			for _, r := range part {
				if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
					cleaned.WriteRune(r)
				}
			}
			
			result := cleaned.String()
			if result != "" && result != "-" {
				return result
			}
		}
	}
	
	return ""
}

func init() {
	packCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output archive path")
	packCmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "path to manifest file (default: rulestack.json)")
	packCmd.Flags().StringVarP(&fileOverride, "file", "f", "", "override single file to pack (auto-creates manifest if none exists)")
}