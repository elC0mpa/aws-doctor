package cmd

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/spf13/cobra"
)

var trendCmd = &cobra.Command{
	Use:   "trend",
	Short: "Display a trend report for the last 6 months",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildOrchestrator(true)
		if err != nil {
			return err
		}

		flags := model.Flags{
			Region:  region,
			Profile: profile,
			Output:  outputFormat,
			Trend:   true,
		}

		return orch.Orchestrate(flags)
	},
}

func init() {
	rootCmd.AddCommand(trendCmd)
}
