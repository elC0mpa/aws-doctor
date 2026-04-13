package cmd

import (
	"context"
	"fmt"

	"github.com/elC0mpa/aws-doctor/model"
	awsconfig "github.com/elC0mpa/aws-doctor/service/aws_config"
	"github.com/elC0mpa/aws-doctor/service/cloudwatchlogs"
	"github.com/elC0mpa/aws-doctor/service/cloudwatchmetrics"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	"github.com/elC0mpa/aws-doctor/service/elb"
	awslambda "github.com/elC0mpa/aws-doctor/service/lambda"
	"github.com/elC0mpa/aws-doctor/service/orchestrator"
	"github.com/elC0mpa/aws-doctor/service/output"
	"github.com/elC0mpa/aws-doctor/service/rds"
	"github.com/elC0mpa/aws-doctor/service/report"
	"github.com/elC0mpa/aws-doctor/service/s3"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"github.com/elC0mpa/aws-doctor/service/update"
	awsvpc "github.com/elC0mpa/aws-doctor/service/vpc"
	"github.com/elC0mpa/aws-doctor/utils/banner"
	"github.com/elC0mpa/aws-doctor/utils/spinner"
	"github.com/spf13/cobra"
)

var (
	region              string
	profile             string
	outputFormat        string
	versionInfo         model.VersionInfo
	orchestratorBuilder = buildOrchestrator
)

func buildOrchestrator(needsAWS bool) (orchestrator.Service, error) {
	outputService := output.NewService(outputFormat)
	updateService := update.NewService(versionInfo)

	config := orchestrator.Config{
		OutputService: outputService,
		UpdateService: updateService,
		VersionInfo:   versionInfo,
	}

	if !needsAWS {
		return orchestrator.NewService(config), nil
	}

	banner.DrawBannerTitle()

	cfgService := awsconfig.NewService()

	awsCfg, err := cfgService.GetAWSCfg(context.Background(), region, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	spinner.StartSpinner()

	cwMetricsService := cloudwatchmetrics.NewService(awsCfg)

	config.STSService = awssts.NewService(awsCfg)
	config.CostService = awscostexplorer.NewService(awsCfg)
	config.EC2Service = awsec2.NewService(awsCfg)
	config.ELBService = elb.NewService(awsCfg, cwMetricsService)
	config.S3Service = s3.NewService(awsCfg)
	config.CloudWatchLogsService = cloudwatchlogs.NewService(awsCfg)
	config.RDSService = rds.NewService(awsCfg, cwMetricsService)
	config.VPCService = awsvpc.NewService(awsCfg, cwMetricsService)
	config.LambdaService = awslambda.NewService(awsCfg)
	config.ReportService = report.NewService()

	return orchestrator.NewService(config), nil
}

var rootCmd = &cobra.Command{
	Use:   "aws-doctor",
	Short: "A comprehensive health check for your AWS accounts",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(version, commit, date string) error {
	versionInfo = model.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "AWS region (defaults to AWS_REGION, AWS_DEFAULT_REGION, or ~/.aws/config)")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "AWS profile configuration")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "table", "Output format: table, json or csv")
}
