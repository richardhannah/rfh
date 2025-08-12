package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"rulestack/internal/manifest"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new ruleset project",
	Long: `Creates a new rulestack.json manifest file and basic directory structure
for developing AI rulesets.

This command will create:
- rulestack.json (manifest file)
- rules/ directory (for storing rule files)
- README.md (basic documentation)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func runInit() error {
	manifestPath := "rulestack.json"

	// Check if manifest already exists
	if _, err := os.Stat(manifestPath); err == nil {
		fmt.Printf("rulestack.json already exists. Use --force to overwrite.\n")
		return nil
	}

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

	fmt.Printf("✅ Initialized new ruleset project\n")
	fmt.Printf("📁 Created files:\n")
	fmt.Printf("   - rulestack.json (manifest)\n")
	fmt.Printf("   - rules/example-rule.md (sample rule)\n")
	fmt.Printf("   - README.md (documentation)\n")
	fmt.Printf("\n🚀 Next steps:\n")
	fmt.Printf("   1. Edit rulestack.json with your package details\n")
	fmt.Printf("   2. Add your rule files to rules/\n")
	fmt.Printf("   3. Run 'rfh pack' to create archive\n")
	fmt.Printf("   4. Run 'rfh publish' to publish to registry\n")

	return nil
}

func init() {
	// Add flags if needed
	initCmd.Flags().BoolP("force", "f", false, "force overwrite existing files")
}