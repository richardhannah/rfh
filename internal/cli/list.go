package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List installed rulesets",
	Long:    `List all rulesets that have been installed in the current workspace.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("list command not yet implemented")
	},
}