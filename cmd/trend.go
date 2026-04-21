package cmd

import (
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/spf13/cobra"
)

var trendCmd = &cobra.Command{
	Use:   "trend [services...]",
	Short: "Display a trend report for the last 6 months",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := orchestratorBuilder(true, false)
		if err != nil {
			return err
		}

		var parsedChecks []string

		for _, arg := range args {
			parsedChecks = append(parsedChecks, strings.Split(arg, ",")...)
		}

		flags := model.Flags{
			Region:      region,
			Profile:     profile,
			Output:      outputFormat,
			Trend:       true,
			TrendChecks: parsedChecks,
		}

		return orch.Orchestrate(flags)
	},
}

func init() {
	rootCmd.AddCommand(trendCmd)
}
