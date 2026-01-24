package orchestrator

import (
	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	"github.com/elC0mpa/aws-doctor/service/elb"
	"github.com/elC0mpa/aws-doctor/service/output"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
)

type service struct {
	stsService    awssts.Service
	costService   awscostexplorer.CostService
	ec2Service    awsec2.EC2Service
	elbService    elb.Service
	outputService output.Service
}

// Service is the interface for orchestrator service.
type Service interface {
	Orchestrate(flags model.Flags) error
}
