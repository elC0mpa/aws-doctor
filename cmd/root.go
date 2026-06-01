package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/analyzer"
	awsconfig "github.com/elC0mpa/aws-doctor/service/aws_config"
	"github.com/elC0mpa/aws-doctor/service/cache"
	"github.com/elC0mpa/aws-doctor/service/cloudwatchlogs"
	"github.com/elC0mpa/aws-doctor/service/cloudwatchmetrics"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	"github.com/elC0mpa/aws-doctor/service/ecr"
	"github.com/elC0mpa/aws-doctor/service/elb"
	"github.com/elC0mpa/aws-doctor/service/iam"
	awslambda "github.com/elC0mpa/aws-doctor/service/lambda"
	"github.com/elC0mpa/aws-doctor/service/orchestrator"
	"github.com/elC0mpa/aws-doctor/service/output"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"github.com/elC0mpa/aws-doctor/service/rds"
	"github.com/elC0mpa/aws-doctor/service/report"
	"github.com/elC0mpa/aws-doctor/service/s3"
	awssagemaker "github.com/elC0mpa/aws-doctor/service/sagemaker"
	awssecretsmanager "github.com/elC0mpa/aws-doctor/service/secretsmanager"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"github.com/elC0mpa/aws-doctor/service/update"
	awsvpc "github.com/elC0mpa/aws-doctor/service/vpc"
	"github.com/elC0mpa/aws-doctor/utils/banner"
	"github.com/elC0mpa/aws-doctor/utils/spinner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	region       string
	profile      string
	outputFormat string
	versionInfo  model.VersionInfo

	// Test hooks
	buildWasteOrchestratorHook  = buildWasteOrchestrator
	buildCostOrchestratorHook   = buildCostOrchestrator
	buildTrendOrchestratorHook  = buildTrendOrchestrator
	buildSystemOrchestratorHook = buildSystemOrchestrator
)

func buildSystemOrchestrator() (orchestrator.SystemService, error) {
	outputService := output.NewService(outputFormat)
	cacheService := cache.NewService()
	updateService := update.NewService(versionInfo, cacheService)

	cfg := orchestrator.SystemConfig{
		OutputService: outputService,
		UpdateService: updateService,
		VersionInfo:   versionInfo,
	}

	return orchestrator.NewSystemService(cfg), nil
}

func buildWasteOrchestrator() (orchestrator.WasteService, error) {
	cfgService := awsconfig.NewService()

	awsCfg, err := cfgService.GetAWSCfg(context.Background(), region, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	banner.DrawBannerTitle()
	spinner.StartSpinner()

	outputService := output.NewService(outputFormat)
	stsService := awssts.NewService(awsCfg)
	cwMetricsService := cloudwatchmetrics.NewService(awsCfg)
	pricingSvc := pricing.NewService(awsCfg)
	cwLogsService := cloudwatchlogs.NewService(awsCfg, pricingSvc)
	reportService := report.NewService()

	reg := analyzer.NewRegistry()
	reg.Register(awsec2.NewService(awsCfg, cwMetricsService, pricingSvc))
	reg.Register(awsvpc.NewService(awsCfg, cwMetricsService, pricingSvc))
	reg.Register(elb.NewService(awsCfg, cwMetricsService, pricingSvc))
	reg.Register(s3.NewService(awsCfg))
	reg.Register(cwLogsService)
	reg.Register(rds.NewService(awsCfg, cwMetricsService, pricingSvc))
	reg.Register(awslambda.NewService(awsCfg, cwLogsService))
	reg.Register(awssagemaker.NewService(awsCfg, cwMetricsService, pricingSvc))
	reg.Register(ecr.NewService(awsCfg, pricingSvc))
	reg.Register(awssecretsmanager.NewService(awsCfg, pricingSvc))
	reg.Register(iam.NewService(awsCfg))

	cfg := orchestrator.WasteConfig{
		OutputService:  outputService,
		STSService:     stsService,
		PricingService: pricingSvc,
		ReportService:  reportService,
		Registry:       reg,
	}

	return orchestrator.NewWasteService(cfg), nil
}

func buildCostOrchestrator() (orchestrator.CostService, error) {
	cfgService := awsconfig.NewService()

	awsCfg, err := cfgService.GetAWSCfg(context.Background(), region, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	banner.DrawBannerTitle()
	spinner.StartSpinner()

	cfg := orchestrator.CostConfig{
		OutputService: output.NewService(outputFormat),
		STSService:    awssts.NewService(awsCfg),
		CostService:   awscostexplorer.NewService(awsCfg),
		ReportService: report.NewService(),
	}

	return orchestrator.NewCostService(cfg), nil
}

func buildTrendOrchestrator() (orchestrator.TrendService, error) {
	cfgService := awsconfig.NewService()

	awsCfg, err := cfgService.GetAWSCfg(context.Background(), region, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	banner.DrawBannerTitle()
	spinner.StartSpinner()

	cfg := orchestrator.TrendConfig{
		OutputService: output.NewService(outputFormat),
		STSService:    awssts.NewService(awsCfg),
		CostService:   awscostexplorer.NewService(awsCfg),
		ReportService: report.NewService(),
	}

	return orchestrator.NewTrendService(cfg), nil
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

	// Trigger background update
	sysOrch, _ := buildSystemOrchestratorHook()
	updateCh := sysOrch.CheckForUpdateInBackground()

	err := rootCmd.Execute()

	// Wait briefly for update check to complete and print if necessary
	select {
	case res := <-updateCh:
		if res.Err == nil && res.LatestVersion != nil {
			output.NewService(outputFormat).PrintNewVersionAvailable(versionInfo.Version, *res.LatestVersion)
		}
	case <-time.After(500 * time.Millisecond):
	}

	return err
}

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if !term.IsTerminal(int(os.Stdout.Fd())) && !rootCmd.PersistentFlags().Changed("output") {
			outputFormat = "json"
		}

		return nil
	}

	rootCmd.PersistentFlags().StringVar(&region, "region", "", "AWS region (defaults to AWS_REGION, AWS_DEFAULT_REGION, or ~/.aws/config)")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "AWS profile configuration")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "table", "Output format: table, json or csv")
}
