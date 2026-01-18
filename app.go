package main

import (
	"context"

	awsconfig "github.com/elC0mpa/aws-doctor/service/aws_config"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	"github.com/elC0mpa/aws-doctor/service/flag"
	"github.com/elC0mpa/aws-doctor/service/orchestrator"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"github.com/elC0mpa/aws-doctor/utils"
)

func main() {
	utils.DrawBanner()
	utils.StartSpinner()

	flagService := flag.NewService()
	flags, err := flagService.GetParsedFlags()
	if err != nil {
		panic(err)
	}

	cfgService := awsconfig.NewService()
	awsCfg, err := cfgService.GetAWSCfg(context.Background(), flags.Region, flags.Profile)
	if err != nil {
		panic(err)
	}

	costService := awscostexplorer.NewService(awsCfg)
	stsService := awssts.NewService(awsCfg)
	ec2Service := awsec2.NewService(awsCfg)

	orchestratorService := orchestrator.NewService(stsService, costService, ec2Service)

	err = orchestratorService.Orchestrate(flags)
	if err != nil {
		panic(err)
	}
}
