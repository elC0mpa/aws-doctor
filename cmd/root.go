package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/elC0mpa/aws-doctor/model"
	awsconfig "github.com/elC0mpa/aws-doctor/service/aws_config"
	"github.com/elC0mpa/aws-doctor/service/cloudwatchlogs"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	"github.com/elC0mpa/aws-doctor/service/elb"
	"github.com/elC0mpa/aws-doctor/service/orchestrator"
	"github.com/elC0mpa/aws-doctor/service/output"
	"github.com/elC0mpa/aws-doctor/service/s3"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"github.com/elC0mpa/aws-doctor/service/update"
	"github.com/elC0mpa/aws-doctor/utils/banner"
	"github.com/elC0mpa/aws-doctor/utils/spinner"
)

var (
	Region      string
	Profile     string
	Output      string
	VersionInfo model.VersionInfo
)

func buildOrchestrator(needsAWS bool) (orchestrator.Service, error) {
	outputService := output.NewService(Output)
	updateService := update.NewService()

	if !needsAWS {
		return orchestrator.NewService(nil, nil, nil, nil, nil, nil, outputService, updateService, VersionInfo), nil
	}

	banner.DrawBannerTitle()

	cfgService := awsconfig.NewService()
	awsCfg, err := cfgService.GetAWSCfg(context.Background(), Region, Profile)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	spinner.StartSpinner()

	costService := awscostexplorer.NewService(awsCfg)
	stsService := awssts.NewService(awsCfg)
	ec2Service := awsec2.NewService(awsCfg)
	elbService := elb.NewService(awsCfg)
	s3Service := s3.NewService(awsCfg)
	cloudwatchlogsService := cloudwatchlogs.NewService(awsCfg)

	return orchestrator.NewService(stsService, costService, ec2Service, elbService, s3Service, cloudwatchlogsService, outputService, updateService, VersionInfo), nil
}

var rootCmd = &cobra.Command{
	Use:   "aws-doctor",
	Short: "A comprehensive health check for your AWS accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := buildOrchestrator(true)
		if err != nil {
			return err
		}
		flags := model.Flags{
			Region:  Region,
			Profile: Profile,
			Output:  Output,
		}
		return orch.Orchestrate(flags)
	},
}

func Execute(version, commit, date string) error {
	VersionInfo = model.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&Region, "region", "", "AWS region (defaults to AWS_REGION, AWS_DEFAULT_REGION, or ~/.aws/config)")
	rootCmd.PersistentFlags().StringVar(&Profile, "profile", "", "AWS profile configuration")
	rootCmd.PersistentFlags().StringVar(&Output, "output", "table", "Output format: table or json")
}
