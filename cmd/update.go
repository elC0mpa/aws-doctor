package cmd

import (
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update aws-doctor to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildSystemOrchestratorHook()
		if err != nil {
			return err
		}

		return orch.Update()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
