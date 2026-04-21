package cmd

import (
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/spf13/cobra"
)

var lambdaMemoryThreshold int

var wasteCmd = &cobra.Command{
	Use:   "waste [checks...]",
	Short: "Display AWS waste report (e.g., ec2 s3)",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := orchestratorBuilder(true, true)
		if err != nil {
			return err
		}

		var parsedChecks []string

		for _, arg := range args {
			parsedChecks = append(parsedChecks, strings.Split(arg, ",")...)
		}

		flags := model.Flags{
			Region:                region,
			Profile:               profile,
			Output:                outputFormat,
			Waste:                 true,
			WasteChecks:           parsedChecks,
			LambdaMemoryThreshold: lambdaMemoryThreshold,
		}

		return orch.Orchestrate(flags)
	},
}

func init() {
	wasteCmd.Flags().IntVar(&lambdaMemoryThreshold, "lambda-memory-threshold", 10,
		"Memory utilization threshold (%) below which Lambda functions are flagged as over-provisioned")
	rootCmd.AddCommand(wasteCmd)
}
