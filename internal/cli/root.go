package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"rulestack/internal/config"
)

var (
	verbose  bool
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
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(registryCmd)
	rootCmd.AddCommand(authCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// No custom config file support - use defaults
}

// Helper function to handle errors
func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}