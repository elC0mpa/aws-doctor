package cmd

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildOrchestrator(false)
		if err != nil {
			return err
		}

		flags := model.Flags{
			Output:  outputFormat,
			Version: true,
		}

		return orch.Orchestrate(flags)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
