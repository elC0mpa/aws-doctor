package orchestrator

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/analyzer"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	"github.com/elC0mpa/aws-doctor/service/output"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"github.com/elC0mpa/aws-doctor/service/report"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"github.com/elC0mpa/aws-doctor/service/update"
)

type service struct {
	stsService     awssts.Service
	costService    awscostexplorer.Service
	pricingService pricing.Service
	outputService  output.Service
	updateService  update.Service
	reportService  report.Service
	registry       analyzer.Registry
	versionInfo    model.VersionInfo
}

// Config holds the dependencies for the orchestrator service.
type Config struct {
	STSService     awssts.Service
	CostService    awscostexplorer.Service
	PricingService pricing.Service
	OutputService  output.Service
	UpdateService  update.Service
	ReportService  report.Service
	Registry       analyzer.Registry
	VersionInfo    model.VersionInfo
}

// Service is the interface for the orchestrator service.
type Service interface {
	Orchestrate(flags model.Flags) error
}
