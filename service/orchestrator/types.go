package orchestrator

import (
	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	"github.com/elC0mpa/aws-doctor/service/elb"
	"github.com/elC0mpa/aws-doctor/service/output"
	"github.com/elC0mpa/aws-doctor/service/route53"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
)

type service struct {
	stsService     awssts.STSService
	costService    awscostexplorer.CostService
	ec2Service     awsec2.EC2Service
	elbService     elb.ELBService
	route53Service route53.Route53Service
	outputService  output.Service
}

type OrchestratorService interface {
	Orchestrate(model.Flags) error
}
