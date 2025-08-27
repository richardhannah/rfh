package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"rulestack/internal/manifest"
)

var (
	outputPath   string
	fileOverride string  // Single file override
	packageName  string  // Non-interactive package name
	newVersion   string  // Non-interactive version
	addToExisting bool   // Non-interactive: add to existing package
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
   - rfh pack my-rule.mdc --package="my-package" --version="1.0.1" --add-to-existing
   - rfh pack my-rule.mdc --package="new-package"  # Creates new package at v1.0.0

The pack command:
- Validates .mdc file format
- Updates rulestack.json with new/updated package info  
- Manages .rulestack package directories
- Creates staged archive ready for publishing
- Handles semantic version validation and incrementing

Examples:
  rfh pack my-security-rule.mdc                                    # Interactive
  rfh pack my-rule.mdc --package="security-rules" --version="1.0.1" --add-to-existing  # Update existing
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
	// Check if rulestack.json exists
	manifestFile := "rulestack.json"
	manifests, err := manifest.LoadAll(manifestFile)
	if err != nil {
		if os.IsNotExist(err) {
			// No rulestack.json exists, create new package
			return createNewPackage(fileName)
		}
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	if len(manifests) == 0 {
		// Empty manifest file, create new package
		return createNewPackage(fileName)
	}

	// Ask user if they want to add to existing package
	fmt.Printf("Found %d existing package(s).\n", len(manifests))
	addToExisting, err := promptUserChoice("Add file to existing package?")
	if err != nil {
		return err
	}

	if !addToExisting {
		// User chose to create new package
		return createNewPackage(fileName)
	}

	// User chose to add to existing package
	packageIndex, err := promptPackageSelection(manifests)
	if err != nil {
		return err
	}

	return addToExistingPackage(fileName, manifests, packageIndex)
}

func init() {
	packCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output archive path")
	packCmd.Flags().StringVarP(&fileOverride, "file", "f", "", "override single file to pack")
	
	// Non-interactive mode flags
	packCmd.Flags().StringVarP(&packageName, "package", "p", "", "package name (enables non-interactive mode)")
	packCmd.Flags().StringVarP(&newVersion, "version", "", "", "package version (required with --add-to-existing)")
	packCmd.Flags().BoolVar(&addToExisting, "add-to-existing", false, "add to existing package (requires --package and --version)")
}