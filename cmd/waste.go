package cmd

import (
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/spf13/cobra"
)

var (
	lambdaMemoryThreshold int
	secretsIdleDays       int
	iamIdleDays           int

	ec2StoppedDays            int
	ec2RiExpiringDays         int
	ec2AmiStaleDays           int
	ec2SnapshotStaleDays      int
	ec2IdleDays               int
	ec2IdleCPUPercent         float64
	ec2IdleNetworkBytesPerDay int
	sagemakerIdleDays         int
	vpcNatIdleDays            int
	elbIdleDays               int
	rdsIdleDays               int
	rdsSnapshotDays           int
	lambdaLookbackDays        int
)

var wasteCmd = &cobra.Command{
	Use:   "waste [checks...]",
	Short: "Display AWS waste report (e.g., ec2 s3 ecr)",
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
			Output:                    outputFormat,
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

func addWasteFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&lambdaMemoryThreshold, "lambda-memory-threshold", 10,
		"Memory utilization threshold (%) below which Lambda functions are flagged as over-provisioned")
	cmd.Flags().IntVar(&secretsIdleDays, "secrets-idle-days", 90,
		"Idle days threshold for flagging unused Secrets Manager secrets")
	cmd.Flags().IntVar(&iamIdleDays, "iam-idle-days", 90,
		"Idle days threshold for flagging unused IAM Users")
	cmd.Flags().IntVar(&ec2StoppedDays, "ec2-stopped-days", 30, "Idle days threshold for flagging stopped EC2 instances")
	cmd.Flags().IntVar(&ec2RiExpiringDays, "ec2-ri-expiring-days", 30, "Days threshold for flagging expiring EC2 Reserved Instances")
	cmd.Flags().IntVar(&ec2AmiStaleDays, "ec2-ami-stale-days", 90, "Days threshold for flagging stale AMIs")
	cmd.Flags().IntVar(&ec2SnapshotStaleDays, "ec2-snapshot-stale-days", 90, "Days threshold for flagging stale EC2 snapshots")
	cmd.Flags().IntVar(&ec2IdleDays, "ec2-idle-days", 14, "Idle days threshold for flagging idle EC2 instances")
	cmd.Flags().Float64Var(&ec2IdleCPUPercent, "ec2-idle-cpu-percent", 5.0, "CPU percentage threshold for flagging idle EC2 instances")
	cmd.Flags().IntVar(&ec2IdleNetworkBytesPerDay, "ec2-idle-network-bytes", 5242880, "Network bytes per day threshold for flagging idle EC2 instances")
	cmd.Flags().IntVar(&sagemakerIdleDays, "sagemaker-idle-days", 14, "Idle days threshold for flagging idle SageMaker endpoints")
	cmd.Flags().IntVar(&vpcNatIdleDays, "vpc-nat-idle-days", 7, "Idle days threshold for flagging idle NAT Gateways")
	cmd.Flags().IntVar(&elbIdleDays, "elb-idle-days", 7, "Idle days threshold for flagging idle Elastic Load Balancers")
	cmd.Flags().IntVar(&rdsIdleDays, "rds-idle-days", 7, "Idle days threshold for flagging idle RDS instances")
	cmd.Flags().IntVar(&rdsSnapshotDays, "rds-snapshot-days", 30, "Days threshold for flagging stale RDS snapshots")
	cmd.Flags().IntVar(&lambdaLookbackDays, "lambda-lookback-days", 14, "Lookback days for analyzing Lambda functions")
}

func init() {
	addWasteFlags(wasteCmd)
	rootCmd.AddCommand(wasteCmd)
}
