package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildSystemOrchestratorHook()
		if err != nil {
			return err
		}

		return orch.Version()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
