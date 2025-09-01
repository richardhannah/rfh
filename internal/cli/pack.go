package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	outputPath     string
	fileOverride   string  // Single file override
	packageName    string  // Non-interactive package name
	packageVersion string  // Non-interactive package version
)

// packCmd represents the pack command
var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Pack ruleset files into a distributable archive",
	Long: `Creates a tar.gz archive containing ruleset files and stages it for publishing.

The pack command supports both interactive and non-interactive modes:

Interactive mode:
   - rfh pack --file=my-rule.mdc
   - Prompts to add file to existing package or create new one
   - Handles version incrementing and directory management

Non-interactive mode:
   - rfh pack --file=my-rule.mdc --package="new-package"  # Creates new package at v1.0.0
   - rfh pack --file=my-rule.mdc --package="new-package" --version="1.2.0"  # Creates new package at v1.2.0

The pack command:
- Validates .mdc file format
- Updates rulestack.json with new/updated package info  
- Manages .rulestack package directories
- Creates staged archive ready for publishing
- Handles semantic version validation and incrementing

Examples:
  rfh pack --file=my-security-rule.mdc                                    # Interactive
  rfh pack --file=my-rule.mdc --package="new-rules"                      # Create new package
  rfh pack --file=my-rule.mdc --package="new-rules" --version="2.1.0"    # Create new package with version`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if fileOverride == "" {
			return fmt.Errorf("--file flag is required")
		}
		
		// Validate that the file is a .mdc file
		if !isValidMdcFile(fileOverride) {
			return fmt.Errorf("file must be a valid .mdc file: %s", fileOverride)
		}
		
		// Check if non-interactive mode
		if packageName != "" {
			return runNonInteractivePack(fileOverride)
		}
		
		return runInteractivePack(fileOverride)
	},
}

func runInteractivePack(fileName string) error {
	// Pack is now much simpler - just create a new package
	// No need to read existing manifests, just prompt for package details
	return createNewPackage(fileName)
}

func init() {
	packCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output archive path")
	packCmd.Flags().StringVarP(&fileOverride, "file", "f", "", ".mdc file to pack (required)")
	
	// Non-interactive mode flags
	packCmd.Flags().StringVarP(&packageName, "package", "p", "", "package name (enables non-interactive mode)")
	packCmd.Flags().StringVarP(&packageVersion, "version", "", "", "package version (auto-increments for existing packages, defaults to 1.0.0 for new packages)")
}