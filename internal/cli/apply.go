package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply <package>[@version]",
	Short: "Apply a ruleset to an editor workspace",
	Long:  `Apply a previously added ruleset to a specific editor workspace.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("apply command not yet implemented")
	},
}