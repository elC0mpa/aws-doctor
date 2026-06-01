package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var trendCmd = &cobra.Command{
	Use:   "trend [services...]",
	Short: "Display a trend report for the last 6 months",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildTrendOrchestratorHook()
		if err != nil {
			return err
		}

		var parsedChecks []string

		for _, arg := range args {
			parsedChecks = append(parsedChecks, strings.Split(arg, ",")...)
		}

		return orch.AnalyzeTrends(parsedChecks, false, "")
	},
}

func init() {
	rootCmd.AddCommand(trendCmd)
}
