package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	outputPath   string
	fileOverride string  // Single file override
	packageName  string  // Non-interactive package name
)

// packCmd represents the pack command
var packCmd = &cobra.Command{
	Use:   "pack [file.mdc]",
	Short: "Pack ruleset files into a distributable archive",
	Long: `Creates a tar.gz archive containing ruleset files and stages it for publishing.

The pack command supports both interactive and non-interactive modes:

Interactive mode:
   - rfh pack my-rule.mdc
   - Prompts to add file to existing package or create new one
   - Handles version incrementing and directory management

Non-interactive mode:
   - rfh pack my-rule.mdc --package="new-package"  # Creates new package at v1.0.0

The pack command:
- Validates .mdc file format
- Updates rulestack.json with new/updated package info  
- Manages .rulestack package directories
- Creates staged archive ready for publishing
- Handles semantic version validation and incrementing

Examples:
  rfh pack my-security-rule.mdc                                    # Interactive
  rfh pack my-rule.mdc --package="new-rules"                      # Create new package`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName := args[0]
		
		// Validate that the file is a .mdc file
		if !isValidMdcFile(fileName) {
			return fmt.Errorf("file must be a valid .mdc file: %s", fileName)
		}
		
		// Check if non-interactive mode
		if packageName != "" {
			return runNonInteractivePack(fileName)
		}
		
		return runInteractivePack(fileName)
	},
}

func runInteractivePack(fileName string) error {
	// Pack is now much simpler - just create a new package
	// No need to read existing manifests, just prompt for package details
	return createNewPackage(fileName)
}

func init() {
	packCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output archive path")
	packCmd.Flags().StringVarP(&fileOverride, "file", "f", "", "override single file to pack")
	
	// Non-interactive mode flags
	packCmd.Flags().StringVarP(&packageName, "package", "p", "", "package name (enables non-interactive mode)")
}