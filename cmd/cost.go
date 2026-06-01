package cmd

import (
	"github.com/spf13/cobra"
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Display comparative cost analytics (Current month vs. Last month)",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildCostOrchestratorHook()
		if err != nil {
			return err
		}

		return orch.CompareCosts(false, "")
	},
}

func init() {
	rootCmd.AddCommand(costCmd)
}
