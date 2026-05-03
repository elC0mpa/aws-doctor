package orchestrator

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/cloudwatchlogs"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	"github.com/elC0mpa/aws-doctor/service/ecr"
	"github.com/elC0mpa/aws-doctor/service/elb"
	"github.com/elC0mpa/aws-doctor/service/lambda"
	"github.com/elC0mpa/aws-doctor/service/output"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"github.com/elC0mpa/aws-doctor/service/rds"
	"github.com/elC0mpa/aws-doctor/service/report"
	"github.com/elC0mpa/aws-doctor/service/s3"
	"github.com/elC0mpa/aws-doctor/service/sagemaker"
	"github.com/elC0mpa/aws-doctor/service/secretsmanager"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"github.com/elC0mpa/aws-doctor/service/update"
	awsvpc "github.com/elC0mpa/aws-doctor/service/vpc"
)

type service struct {
	stsService            awssts.Service
	costService           awscostexplorer.Service
	ec2Service            awsec2.Service
	elbService            elb.Service
	s3Service             s3.Service
	cloudwatchlogsService cloudwatchlogs.Service
	rdsService            rds.Service
	lambdaService         lambda.Service
	sagemakerService      sagemaker.Service
	secretsmanagerService secretsmanager.Service
	pricingService        pricing.Service
	outputService         output.Service
	updateService         update.Service
	reportService         report.Service
	ecrService            ecr.Service
	versionInfo           model.VersionInfo
	vpcService            awsvpc.Service
}

// Config holds the dependencies for the orchestrator service.
type Config struct {
	STSService            awssts.Service
	CostService           awscostexplorer.Service
	EC2Service            awsec2.Service
	ELBService            elb.Service
	S3Service             s3.Service
	CloudWatchLogsService cloudwatchlogs.Service
	RDSService            rds.Service
	LambdaService         lambda.Service
	SageMakerService      sagemaker.Service
	SecretsManagerService secretsmanager.Service
	PricingService        pricing.Service
	OutputService         output.Service
	UpdateService         update.Service
	ReportService         report.Service
	ECRService            ecr.Service
	VersionInfo           model.VersionInfo
	VPCService            awsvpc.Service
}

// Service is the interface for the orchestrator service.
type Service interface {
	Orchestrate(flags model.Flags) error
}
