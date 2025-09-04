package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"rulestack/internal/client"
	"rulestack/internal/config"
)

var (
	searchTag    string
	searchTarget string
	searchLimit  int
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for rulesets in the registry",
	Long: `Search for rulesets in the configured registry.

You can filter results by tags and targets to find rulesets that match
your specific needs.

Examples:
  rfh search security
  rfh search "secure coding" --tag=javascript
  rfh search linting --target=cursor
  rfh search react --limit=10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearch(args[0])
	},
}

func runSearch(query string) error {
	// Get registry configuration
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Use current registry (no overrides)
	registryName := cfg.Current

	if registryName == "" {
		return fmt.Errorf("no registry configured. Use 'rfh registry add' to add a registry")
	}

	reg, exists := cfg.Registries[registryName]
	if !exists {
		return fmt.Errorf("registry '%s' not found. Use 'rfh registry list' to see available registries", registryName)
	}

	if verbose {
		fmt.Printf("🔍 Searching for: %s\n", query)
		fmt.Printf("🌐 Registry: %s (%s)\n", registryName, reg.URL)
		if searchTag != "" {
			fmt.Printf("🏷️  Tag filter: %s\n", searchTag)
		}
		if searchTarget != "" {
			fmt.Printf("🎯 Target filter: %s\n", searchTarget)
		}
	}

	// Create client using new factory
	c, err := client.GetClient(cfg, verbose)
	if err != nil {
		return err
	}

	// Search packages using new interface
	ctx, cancel := client.WithTimeout(context.Background())
	defer cancel()
	
	opts := client.SearchOptions{
		Query:  query,
		Tag:    searchTag,
		Target: searchTarget,
		Limit:  searchLimit,
	}
	
	packages, err := c.SearchPackages(ctx, opts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(packages) == 0 {
		fmt.Printf("No rulesets found matching '%s'\n", query)
		if searchTag != "" || searchTarget != "" {
			fmt.Printf("Try removing filters or using different search terms.\n")
		}
		return nil
	}

	// Display results
	fmt.Printf("📋 Found %d ruleset(s):\n\n", len(packages))

	for _, pkg := range packages {
		name := pkg.Name
		version := pkg.Latest
		description := pkg.Description

		fmt.Printf("📦 %s@%s\n", name, version)

		if description != "" {
			fmt.Printf("   %s\n", description)
		}

		// Display versions
		if len(pkg.Versions) > 1 {
			fmt.Printf("   📋 Versions: %s\n", strings.Join(pkg.Versions, ", "))
		}

		// Display tags
		if len(pkg.Tags) > 0 {
			fmt.Printf("   🏷️  Tags: %s\n", strings.Join(pkg.Tags, ", "))
		}

		fmt.Printf("\n")
	}

	fmt.Printf("💡 Install with: rfh add <package-name>@<version>\n")

	return nil
}

func init() {
	searchCmd.Flags().StringVar(&searchTag, "tag", "", "filter by tag")
	searchCmd.Flags().StringVar(&searchTarget, "target", "", "filter by target (cursor, claude-code, etc.)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 20, "limit number of results")
}
