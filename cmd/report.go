package cmd

import (
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/spf13/cobra"
)

var reportOutPath string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate PDF reports for AWS cost, waste, or trends",
}

var reportCostCmd = &cobra.Command{
	Use:   "cost",
	Short: "Generate a PDF report for cost comparison",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := orchestratorBuilder(true)
		if err != nil {
			return err
		}

		flags := model.Flags{
			Region:     region,
			Profile:    profile,
			Report:     true,
			ReportPath: reportOutPath,
		}

		return orch.Orchestrate(flags)
	},
}

var reportWasteCmd = &cobra.Command{
	Use:   "waste [checks...]",
	Short: "Generate a PDF report for AWS waste detection",
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
			Region:                region,
			Profile:               profile,
			Report:                true,
			ReportPath:            reportOutPath,
			Waste:                 true,
			WasteChecks:           parsedChecks,
			LambdaMemoryThreshold: lambdaMemoryThreshold,
			SageMakerIdleDays:     sageMakerIdleDays,
		}

		return orch.Orchestrate(flags)
	},
}

var reportTrendCmd = &cobra.Command{
	Use:   "trend [services...]",
	Short: "Generate a PDF report for AWS cost trends",
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
			Report:      true,
			ReportPath:  reportOutPath,
			Trend:       true,
			TrendChecks: parsedChecks,
		}

		return orch.Orchestrate(flags)
	},
}

func init() {
	reportCmd.PersistentFlags().StringVar(&reportOutPath, "path", "", "Output path for the PDF report")
	reportCmd.PersistentFlags().Lookup("path").NoOptDefVal = "DEFAULT"

	reportWasteCmd.Flags().IntVar(&lambdaMemoryThreshold, "lambda-memory-threshold", 10,
		"Memory utilization threshold (%) below which Lambda functions are flagged as over-provisioned")
	reportWasteCmd.Flags().IntVar(&sageMakerIdleDays, "sagemaker-idle-days", 14,
		"Lookback window in days for flagging SageMaker endpoints with zero invocations as idle")

	reportCmd.AddCommand(reportCostCmd)
	reportCmd.AddCommand(reportWasteCmd)
	reportCmd.AddCommand(reportTrendCmd)

	rootCmd.AddCommand(reportCmd)
}
