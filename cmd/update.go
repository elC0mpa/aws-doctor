package cmd

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update aws-doctor to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildOrchestrator(false)
		if err != nil {
			return err
		}
		
		flags := model.Flags{
			Output: Output,
			Update: true,
		}
		return orch.Orchestrate(flags)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
