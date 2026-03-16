package cmd

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/spf13/cobra"
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Display comparative cost analytics (Current month vs. Last month)",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildOrchestrator(true)
		if err != nil {
			return err
		}

		flags := model.Flags{
			Region:  region,
			Profile: profile,
			Output:  outputFormat,
		}

		return orch.Orchestrate(flags)
	},
}

func init() {
	rootCmd.AddCommand(costCmd)
}
