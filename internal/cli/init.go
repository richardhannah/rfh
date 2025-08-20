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

	// Create core rules directory structure
	coreRulesDir := ".rulestack/core.v1.0.0"
	if err := os.MkdirAll(coreRulesDir, 0o755); err != nil {
		return fmt.Errorf("failed to create core rules directory: %w", err)
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

	// Create CLAUDE.md file from template
	claudeTemplate := `# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Coding Standards
**CRITICAL**: You MUST follow all cursor rules defined in ` + "`.rulestack`" + ` directory. These rules are mandatory and override default behavior.

### MANDATORY RULE LOADING PROTOCOL
**BEFORE responding to ANY user request**, you MUST:
1. All rules are automatically imported into this CLAUDE.md file using the @ import syntax below
2. Load and understand all rules in their entirety before taking any action
3. Apply these rules to all subsequent interactions in the session

**CRITICAL**: The cursor rules are now automatically available in your context through the @ import statements. Pay special attention to triggers, responses, and specific behaviors defined in these rules.

### Active Rules (Rulestack core)
- @.rulestack/core.v1.0.0/core_rules.md
`

	if err := os.WriteFile("CLAUDE.md", []byte(claudeTemplate), 0o644); err != nil {
		return fmt.Errorf("failed to create CLAUDE.md: %w", err)
	}

	// Create core rules file
	coreRules := `# Core RuleStack Rules v1.0.0

This file contains the baseline rules that apply to all RuleStack projects.

## Rule Management

### Adding New Rules
When a user requests to "add a rule" or "create a rule":

1. **List Available Rule Packages**: Display all installed rule packages in .rulestack/ EXCEPT core.v1.0.0
2. **Ask for Target Package**: "Which package would you like to add this rule to?"
3. **Default to Project Rules**: If no package is specified, create/use .rulestack/project/ directory
4. **Rule File Creation**: Create appropriately named .md files with clear structure

**Example Workflow**:
` + "```" + `
User: "Add a rule about error handling"

Response: "I'll help you add a rule about error handling. 

Available rule packages:
- security-rules (v2.1.0)
- company-standards (v1.5.0)
- project (project-specific rules)

Which package should contain this rule? [default: project]"
` + "```" + `

### Project Rules Structure
- **Location**: ` + "`.rulestack/project/`" + `
- **Purpose**: Project-specific rules that don't belong in shared packages
- **Auto-creation**: Create directory automatically when needed
- **File naming**: Use descriptive names like ` + "`error_handling.md`" + `, ` + "`api_conventions.md`" + `

### Rule Package Guidelines
- **Core rules** (core.v1.0.0): NEVER modify - system managed
- **Installed packages**: Add rules only with user confirmation
- **Project rules**: Default location for new project-specific rules
- **Rule organization**: Group related rules in appropriate packages

## Code Quality Rules

### Defensive Programming
- Always validate inputs and handle edge cases
- Use explicit error handling rather than silent failures
- Write clear, self-documenting code with meaningful variable names
- Include appropriate logging for debugging and monitoring

### Security Rules
- Never commit secrets, API keys, or sensitive data to repositories
- Validate and sanitize all user inputs
- Use secure coding practices appropriate for the technology stack
- Follow principle of least privilege for permissions and access

### Documentation Rules
- Document all public APIs and interfaces
- Include usage examples in code comments where helpful
- Keep README files up to date with current functionality
- Document any non-obvious business logic or algorithms

## RuleStack-Specific Rules

### Package Management
- Always run 'rfh init' before using other RuleStack commands
- Use semantic versioning for all packages
- Include clear descriptions in package manifests
- Test packages thoroughly before publishing

### Rule Development
- Write rules that are clear and actionable
- Provide examples in rule documentation
- Test rules against real-world scenarios
- Keep rules focused and single-purpose

## Integration Rules

### Claude Code Integration
- Use descriptive commit messages
- Break down large tasks into smaller, manageable steps
- Provide context when asking for code modifications
- Review generated code for correctness and style

### Version Control
- Make atomic commits with clear purposes
- Use meaningful branch names
- Keep commit history clean and readable
- Tag releases appropriately

---

*These core rules are maintained by the RuleStack system and should not be modified directly.*
`

	coreRulesPath := filepath.Join(coreRulesDir, "core_rules.md")
	if err := os.WriteFile(coreRulesPath, []byte(coreRules), 0o644); err != nil {
		return fmt.Errorf("failed to create core rules: %w", err)
	}

	fmt.Printf("✅ Initialized RuleStack project in: %s\n", filepath.Base(projectRoot))
	fmt.Printf("📁 Created:\n")
	fmt.Printf("   - rulestack.json (package manifest)\n")
	fmt.Printf("   - rules/example-rule.md (sample rule)\n")
	fmt.Printf("   - README.md (documentation)\n")
	fmt.Printf("   - CLAUDE.md (Claude Code integration)\n")
	fmt.Printf("   - .rulestack/ (dependency directory)\n")
	fmt.Printf("   - .rulestack/core.v1.0.0/core_rules.md (baseline rules)\n")
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