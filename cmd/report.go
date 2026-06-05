package cmd

import (
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/spf13/cobra"
)

const cmdReportName = "report"

var reportOutPath string

var reportCmd = &cobra.Command{
	Use:   cmdReportName,
	Short: "Generate PDF reports for AWS cost, waste, or trends",
}

var reportCostCmd = &cobra.Command{
	Use:   cmdCostName,
	Short: "Generate a PDF report for cost comparison",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildCostOrchestratorHook()
		if err != nil {
			return err
		}

		return orch.CompareCosts(true, reportOutPath)
	},
}

var reportWasteCmd = &cobra.Command{
	Use:   "waste [checks...]",
	Short: "Generate a PDF report for AWS waste detection",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildWasteOrchestratorHook()
		if err != nil {
			return err
		}

		var parsedChecks []string
		for _, arg := range args {
			parsedChecks = append(parsedChecks, strings.Split(arg, ",")...)
		}

		flags := model.Flags{
			Region:                    region,
			Profile:                   profile,
			Report:                    true,
			ReportPath:                reportOutPath,
			Waste:                     true,
			WasteChecks:               parsedChecks,
			LambdaMemoryThreshold:     lambdaMemoryThreshold,
			SecretsIdleDays:           secretsIdleDays,
			IAMIdleDays:               iamIdleDays,
			EC2StoppedDays:            ec2StoppedDays,
			EC2RiExpiringDays:         ec2RiExpiringDays,
			EC2AmiStaleDays:           ec2AmiStaleDays,
			EC2SnapshotStaleDays:      ec2SnapshotStaleDays,
			EC2IdleDays:               ec2IdleDays,
			EC2IdleCPUPercent:         ec2IdleCPUPercent,
			EC2IdleNetworkBytesPerDay: ec2IdleNetworkBytesPerDay,
			SageMakerIdleDays:         sagemakerIdleDays,
			VPCNatIdleDays:            vpcNatIdleDays,
			ELBIdleDays:               elbIdleDays,
			RDSIdleDays:               rdsIdleDays,
			RDSSnapshotDays:           rdsSnapshotDays,
			LambdaLookbackDays:        lambdaLookbackDays,
		}

		return orch.AnalyzeWaste(flags)
	},
}

var reportTrendCmd = &cobra.Command{
	Use:   "trend [services...]",
	Short: "Generate a PDF report for AWS cost trends",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildTrendOrchestratorHook()
		if err != nil {
			return err
		}

		var parsedChecks []string
		for _, arg := range args {
			parsedChecks = append(parsedChecks, strings.Split(arg, ",")...)
		}

		return orch.AnalyzeTrends(parsedChecks, true, reportOutPath)
	},
}

func init() {
	reportCmd.PersistentFlags().StringVar(&reportOutPath, "path", "", "Output path for the PDF report")
	reportCmd.PersistentFlags().Lookup("path").NoOptDefVal = "DEFAULT"

	addWasteFlags(reportWasteCmd)

	reportCmd.AddCommand(reportCostCmd)
	reportCmd.AddCommand(reportWasteCmd)
	reportCmd.AddCommand(reportTrendCmd)

	rootCmd.AddCommand(reportCmd)
}
