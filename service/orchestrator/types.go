package orchestrator

import (
	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	"github.com/elC0mpa/aws-doctor/service/elb"
	"github.com/elC0mpa/aws-doctor/service/output"
	"github.com/elC0mpa/aws-doctor/service/s3"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"github.com/elC0mpa/aws-doctor/service/update"
)

type service struct {
	stsService    awssts.Service
	costService   awscostexplorer.Service
	ec2Service    awsec2.Service
	elbService    elb.Service
	s3Service     s3.Service
	outputService output.Service
	updateService update.Service
	versionInfo   model.VersionInfo
}

// Service is the interface for orchestrator service.
type Service interface {
	Orchestrate(flags model.Flags) error
}
