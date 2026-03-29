package cmd

import (
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/spf13/cobra"
)

var wasteCmd = &cobra.Command{
	Use:   "waste [checks...]",
	Short: "Display AWS waste report (e.g., ec2 s3)",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := orchestratorBuilder(true)
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
			Waste:       true,
			WasteChecks: parsedChecks,
		}

		return orch.Orchestrate(flags)
	},
}

func init() {
	rootCmd.AddCommand(wasteCmd)
}
