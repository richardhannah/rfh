package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <package>[@version]",
	Short: "Add (download) a ruleset package",
	Long:  `Download and add a ruleset package to the current workspace.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("add command not yet implemented")
	},
}