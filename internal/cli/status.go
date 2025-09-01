package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show packages staged for publishing",
	Long:  `Lists .tgz packages in the staging directory that are ready for publishing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

func runStatus() error {
	stagingDir := ".rulestack/staged"
	pattern := filepath.Join(stagingDir, "*.tgz")
	
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan staging directory: %w", err)
	}
	
	if len(files) == 0 {
		fmt.Println("No staged packages found")
		return nil
	}
	
	for _, file := range files {
		filename := filepath.Base(file)
		fmt.Println(filename)
	}
	
	return nil
}