package main

import (
	"context"
	"fmt"
	"os"

	awsconfig "github.com/elC0mpa/aws-doctor/service/aws_config"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	"github.com/elC0mpa/aws-doctor/service/elb"
	"github.com/elC0mpa/aws-doctor/service/flag"
	"github.com/elC0mpa/aws-doctor/service/orchestrator"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"github.com/elC0mpa/aws-doctor/utils"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-version" {
			fmt.Printf("aws-doctor version %s\n", version)
			fmt.Printf("commit: %s\n", commit)
			fmt.Printf("built at: %s\n", date)
			return nil
		}
	}

	utils.DrawBanner()

	flagService := flag.NewService()
	flags, err := flagService.GetParsedFlags()
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	utils.StartSpinner()
	defer utils.StopSpinner()

	cfgService := awsconfig.NewService()
	awsCfg, err := cfgService.GetAWSCfg(context.Background(), flags.Region, flags.Profile)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	costService := awscostexplorer.NewService(awsCfg)
	stsService := awssts.NewService(awsCfg)
	ec2Service := awsec2.NewService(awsCfg)
	elbService := elb.NewService(awsCfg)

	orchestratorService := orchestrator.NewService(stsService, costService, ec2Service, elbService)

	if err := orchestratorService.Orchestrate(flags); err != nil {
		return fmt.Errorf("orchestration failed: %w", err)
	}

	return nil
}
