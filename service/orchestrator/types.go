package orchestrator

import (
	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
)

type service struct {
	stsService  awssts.STSService
	costService awscostexplorer.CostService
	ec2Service  awsec2.EC2Service
}

type OrchestratorService interface {
	Orchestrate(model.Flags) error
}
