package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"rulestack/internal/manifest"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new RuleStack project",
	Long: `Initialize a new RuleStack project in the current directory.

This establishes the current directory as the project root and creates:
- rulestack.json (package manifest file)
- rules/ directory (for storing rule files)  
- README.md (basic documentation)
- .rulestack/ directory (for dependency management)

Similar to 'git init', this command must be run before using other RFH commands
in this directory. It explicitly sets the project root to the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		return runInit(force)
	},
}

func runInit(force bool) error {
	// Get current directory as project root
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manifestPath := "rulestack.json"

	// Check if already initialized
	if _, err := os.Stat(manifestPath); err == nil {
		if !force {
			fmt.Printf("RuleStack project already initialized (rulestack.json exists).\n")
			fmt.Printf("Use --force to reinitialize.\n")
			return nil
		}
	}

	fmt.Printf("Initializing RuleStack project in: %s\n", projectRoot)

	// Create sample manifest
	sample := manifest.CreateSample()

	// Save manifest
	if err := sample.Save(manifestPath); err != nil {
		return fmt.Errorf("failed to create manifest: %w", err)
	}

	// Create rules directory
	if err := os.MkdirAll("rules", 0o755); err != nil {
		return fmt.Errorf("failed to create rules directory: %w", err)
	}

	// Create .rulestack directory for dependency management
	if err := os.MkdirAll(".rulestack", 0o755); err != nil {
		return fmt.Errorf("failed to create .rulestack directory: %w", err)
	}

	// Create sample rule file
	sampleRule := `# Example Rule

This is an example AI rule file. You can write rules in Markdown format.

## Rule Description
- This rule helps with secure coding practices
- It applies to JavaScript and TypeScript files
- It suggests using const instead of var

## Example
` + "```javascript" + `
// Bad
var userName = "alice";

// Good  
const userName = "alice";
` + "```" + `
`

	ruleFile := "rules/example-rule.md"
	if err := os.WriteFile(ruleFile, []byte(sampleRule), 0o644); err != nil {
		return fmt.Errorf("failed to create sample rule: %w", err)
	}

	// Create README
	readme := fmt.Sprintf(`# %s

%s

## Installation

` + "```bash" + `
rfh add %s
` + "```" + `

## Usage

This ruleset provides AI rules for:
%s

## Files

- ` + "`rules/`" + ` - Rule files in Markdown format
- ` + "`rulestack.json`" + ` - Package manifest

## Publishing

1. Update version in rulestack.json
2. Run ` + "`rfh pack`" + ` to create archive
3. Run ` + "`rfh publish`" + ` to publish to registry
`,
		sample.Name,
		sample.Description,
		sample.Name,
		"- "+sample.Targets[0],
	)

	if err := os.WriteFile("README.md", []byte(readme), 0o644); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	fmt.Printf("✅ Initialized RuleStack project in: %s\n", filepath.Base(projectRoot))
	fmt.Printf("📁 Created:\n")
	fmt.Printf("   - rulestack.json (package manifest)\n")
	fmt.Printf("   - rules/example-rule.md (sample rule)\n")
	fmt.Printf("   - README.md (documentation)\n")
	fmt.Printf("   - .rulestack/ (dependency directory)\n")
	fmt.Printf("\n🚀 Next steps:\n")
	fmt.Printf("   1. Edit rulestack.json with your package details\n")
	fmt.Printf("   2. Add your rule files to rules/\n")
	fmt.Printf("   3. Run 'rfh add <package>' to install dependencies\n")
	fmt.Printf("   4. Run 'rfh pack' to create archive\n")
	fmt.Printf("   5. Run 'rfh publish' to publish to registry\n")

	return nil
}

func init() {
	// Add flags if needed
	initCmd.Flags().BoolP("force", "f", false, "force overwrite existing files")
}