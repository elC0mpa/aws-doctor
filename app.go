// Package main is the entry point for the aws-doctor application.
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
	"github.com/elC0mpa/aws-doctor/service/output"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"github.com/elC0mpa/aws-doctor/utils"
	"github.com/elC0mpa/aws-doctor/model"
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
	flagService := flag.NewService()

	flags, err := flagService.GetParsedFlags()
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	versionInfo := model.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	if flags.Version {
		outputService := output.NewService(flags.Output)
		orchestratorService := orchestrator.NewService(nil, nil, nil, nil, outputService, versionInfo)

		return orchestratorService.Orchestrate(flags)
	}

	utils.DrawBanner()
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
	outputService := output.NewService(flags.Output)

	orchestratorService := orchestrator.NewService(stsService, costService, ec2Service, elbService, outputService, versionInfo)

	if err := orchestratorService.Orchestrate(flags); err != nil {
		return fmt.Errorf("orchestration failed: %w", err)
	}

	return nil
}
