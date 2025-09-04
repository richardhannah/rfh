# Task 8: CLI Foundation and Commands (1 hour)

## Objective
Create the CLI foundation using Cobra and implement basic commands for the RuleStack CLI tool.

## Prerequisites
- Tasks 1-7 completed
- Manifest and archive handling working
- Configuration system established

## Checklist

### 1. Create CLI Main Entry Point (10 minutes)
Create `cmd/cli/main.go`:
```go
package main

import (
    "fmt"
    "os"
    
    "rulestack/internal/cli"
)

func main() {
    if err := cli.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

- [ ] Create main.go file
- [ ] Verify import path is correct

### 2. Create CLI Root Command (15 minutes)
Create `internal/cli/root.go`:
```go
package cli

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    
    "rulestack/internal/config"
)

var (
    cfgFile   string
    registry  string
    token     string
    verbose   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
    Use:   "rfh",
    Short: "RFH - Registry for Humans (AI ruleset manager)",
    Long: `RFH is a package manager for AI rulesets, allowing you to publish,
discover, and install AI rules for use with tools like Claude Code, Cursor, and Windsurf.

Registry for Humans - making AI rulesets accessible and shareable.`,
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // Load .env file if it exists
        config.LoadEnvFile(".env")
        
        if verbose {
            fmt.Printf("RFH version: 1.0.0\n")
            fmt.Printf("Config file: %s\n", cfgFile)
        }
    },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
    return rootCmd.Execute()
}

func init() {
    cobra.OnInitialize(initConfig)
    
    // Global flags
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.rfh/config.toml)")
    rootCmd.PersistentFlags().StringVar(&registry, "registry", "", "registry URL override")
    rootCmd.PersistentFlags().StringVar(&token, "token", "", "auth token override")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
    
    // Add subcommands
    rootCmd.AddCommand(initCmd)
    rootCmd.AddCommand(packCmd)
    rootCmd.AddCommand(publishCmd)
    rootCmd.AddCommand(searchCmd)
    rootCmd.AddCommand(addCmd)
    rootCmd.AddCommand(applyCmd)
    rootCmd.AddCommand(registryCmd)
    rootCmd.AddCommand(listCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
    if cfgFile != "" {
        // Use config file from the flag if provided
        // (Implementation would go here if needed)
    }
}

// Helper function to handle errors
func checkErr(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

- [ ] Create root.go file
- [ ] Verify cobra import works

### 3. Create Init Command (15 minutes)
Create `internal/cli/init.go`:
```go
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
\`\`\`javascript
// Bad
var userName = "alice";

// Good  
const userName = "alice";
\`\`\`
`
    
    ruleFile := "rules/example-rule.md"
    if err := os.WriteFile(ruleFile, []byte(sampleRule), 0o644); err != nil {
        return fmt.Errorf("failed to create sample rule: %w", err)
    }
    
    // Create README
    readme := fmt.Sprintf(`# %s

%s

## Installation

\`\`\`bash
rfh add %s
\`\`\`

## Usage

This ruleset provides AI rules for:
%s

## Files

- \`rules/\` - Rule files in Markdown format
- \`rulestack.json\` - Package manifest

## Publishing

1. Update version in rulestack.json
2. Run \`rfh pack\` to create archive
3. Run \`rfh publish\` to publish to registry
`,
        sample.Name,
        sample.Description,
        sample.Name,
        "- " + sample.Targets[0],
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
```

- [ ] Create init.go file
- [ ] Test that it creates proper directory structure

### 4. Create Pack Command (15 minutes)
Create `internal/cli/pack.go`:
```go
package cli

import (
    "fmt"
    "path/filepath"
    
    "github.com/spf13/cobra"
    
    "rulestack/internal/manifest"
    "rulestack/internal/pkg"
)

var (
    outputPath string
)

// packCmd represents the pack command
var packCmd = &cobra.Command{
    Use:   "pack",
    Short: "Pack ruleset files into a distributable archive",
    Long: `Creates a tar.gz archive containing all files specified in the manifest.
The archive is ready for publishing to a registry.

The pack command:
1. Reads rulestack.json manifest
2. Collects all files matching the patterns in 'files' array
3. Creates a compressed archive
4. Calculates SHA256 hash for integrity`,
    RunE: func(cmd *cobra.Command, args []string) error {
        return runPack()
    },
}

func runPack() error {
    // Load manifest
    manifest, err := manifest.Load("rulestack.json")
    if err != nil {
        return fmt.Errorf("failed to load manifest: %w", err)
    }
    
    // Determine output path
    output := outputPath
    if output == "" {
        // Remove @ and / from package name for filename
        safeName := manifest.GetPackageName()
        if scope := manifest.GetScope(); scope != "" {
            safeName = scope + "-" + safeName
        }
        output = fmt.Sprintf("%s-%s.tgz", safeName, manifest.Version)
    }
    
    if verbose {
        fmt.Printf("📦 Packing %s v%s\n", manifest.Name, manifest.Version)
        fmt.Printf("🎯 Targets: %v\n", manifest.Targets)
        fmt.Printf("🏷️  Tags: %v\n", manifest.Tags)
        fmt.Printf("📄 File patterns: %v\n", manifest.Files)
    }
    
    // Pack files
    info, err := pkg.Pack(manifest.Files, output)
    if err != nil {
        return fmt.Errorf("failed to pack files: %w", err)
    }
    
    fmt.Printf("✅ Successfully packed %s\n", manifest.Name)
    fmt.Printf("📦 Archive: %s\n", info.Path)
    fmt.Printf("📏 Size: %d bytes\n", info.SizeBytes)
    fmt.Printf("🔒 SHA256: %s\n", info.SHA256)
    
    if verbose {
        // List files that were included
        fmt.Printf("\n📋 Files included:\n")
        // This would require modifying Pack to return file list
        // For now, just show the patterns
        for _, pattern := range manifest.Files {
            fmt.Printf("   - %s\n", pattern)
        }
    }
    
    return nil
}

func init() {
    packCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output archive path")
}
```

- [ ] Create pack.go file
- [ ] Verify archive creation works

### 5. Create Registry Command (15 minutes)
Create `internal/cli/registry.go`:
```go
package cli

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    
    "rulestack/internal/config"
)

// registryCmd represents the registry command
var registryCmd = &cobra.Command{
    Use:   "registry",
    Short: "Manage registries",
    Long: `Manage registry configurations for publishing and installing rulesets.

Registries are where rulesets are published and downloaded from. You can
configure multiple registries including public and private ones.`,
}

// registryAddCmd adds a new registry
var registryAddCmd = &cobra.Command{
    Use:   "add <name> <url> [token]",
    Short: "Add a new registry",
    Long: `Add a new registry configuration.

Examples:
  rfh registry add public https://registry.rulestack.dev
  rfh registry add company https://rulestack.company.com my-token
  rfh registry add local http://localhost:8080 dev-token`,
    Args: cobra.RangeArgs(2, 3),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        url := args[1]
        token := ""
        if len(args) > 2 {
            token = args[2]
        }
        
        return runRegistryAdd(name, url, token)
    },
}

// registryListCmd lists configured registries
var registryListCmd = &cobra.Command{
    Use:   "list",
    Short: "List configured registries",
    Long:  `List all configured registries showing name, URL, and active status.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        return runRegistryList()
    },
}

// registryUseCmd sets the active registry
var registryUseCmd = &cobra.Command{
    Use:   "use <name>",
    Short: "Set active registry",
    Long: `Set the active registry for publishing and installing packages.

The active registry is used when no --registry flag is specified.`,
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return runRegistryUse(args[0])
    },
}

func runRegistryAdd(name, url, token string) error {
    cfg, err := config.LoadCLI()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Add registry
    cfg.Registries[name] = config.Registry{
        URL:   url,
        Token: token,
    }
    
    // Set as current if it's the first one
    if cfg.Current == "" {
        cfg.Current = name
    }
    
    // Save config
    if err := config.SaveCLI(cfg); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }
    
    fmt.Printf("✅ Added registry '%s'\n", name)
    fmt.Printf("🌐 URL: %s\n", url)
    if token != "" {
        fmt.Printf("🔑 Token: [configured]\n")
    }
    
    if cfg.Current == name {
        fmt.Printf("⭐ Set as active registry\n")
    }
    
    return nil
}

func runRegistryList() error {
    cfg, err := config.LoadCLI()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    if len(cfg.Registries) == 0 {
        fmt.Printf("No registries configured.\n")
        fmt.Printf("Add a registry with: rfh registry add <name> <url> [token]\n")
        return nil
    }
    
    fmt.Printf("📋 Configured registries:\n\n")
    for name, reg := range cfg.Registries {
        marker := "  "
        if cfg.Current == name {
            marker = "* "
        }
        
        fmt.Printf("%s%s\n", marker, name)
        fmt.Printf("    URL: %s\n", reg.URL)
        if reg.Token != "" {
            fmt.Printf("    Token: [configured]\n")
        }
        fmt.Printf("\n")
    }
    
    if cfg.Current != "" {
        fmt.Printf("* = active registry\n")
    }
    
    return nil
}

func runRegistryUse(name string) error {
    cfg, err := config.LoadCLI()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Check if registry exists
    if _, exists := cfg.Registries[name]; !exists {
        return fmt.Errorf("registry '%s' not found. Use 'rfh registry list' to see available registries", name)
    }
    
    // Set as current
    cfg.Current = name
    
    // Save config
    if err := config.SaveCLI(cfg); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }
    
    fmt.Printf("✅ Set '%s' as active registry\n", name)
    fmt.Printf("🌐 URL: %s\n", cfg.Registries[name].URL)
    
    return nil
}

func init() {
    registryCmd.AddCommand(registryAddCmd)
    registryCmd.AddCommand(registryListCmd)
    registryCmd.AddCommand(registryUseCmd)
}
```

- [ ] Create registry.go file
- [ ] Test registry configuration commands

## Validation
Test the CLI foundation:
```bash
# Build CLI
go build -o rfh ./cmd/cli

# Test commands
./rfh --help
./rfh init
./rfh registry add local http://localhost:8080 test-token
./rfh registry list
./rfh pack
```

## Acceptance Criteria
- [ ] CLI builds without errors
- [ ] Help text displays correctly for all commands
- [ ] Init command creates proper project structure
- [ ] Pack command creates archive from manifest
- [ ] Registry commands manage configuration properly
- [ ] Verbose flag provides additional output
- [ ] Error handling provides helpful messages
- [ ] Config file operations work correctly

## Time Estimate: ~60 minutes

## Next Task
Task 9: HTTP Client and Publishing