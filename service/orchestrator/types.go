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

// WasteConfig holds the dependencies for the waste orchestrator service.
type WasteConfig struct {
	STSService     awssts.Service
	PricingService pricing.Service
	Renderer       output.Renderer
	ReportService  report.Service
	Registry       analyzer.Registry
}

// WasteService is the interface for the waste orchestrator service.
type WasteService interface {
	AnalyzeWaste(flags model.Flags) error
}

// CostConfig holds the dependencies for the cost orchestrator service.
type CostConfig struct {
	STSService    awssts.Service
	CostService   awscostexplorer.Service
	Renderer      output.Renderer
	ReportService report.Service
}

// CostService is the interface for the cost orchestrator service.
type CostService interface {
	CompareCosts(generateReport bool, reportPath string) error
}

// TrendConfig holds the dependencies for the trend orchestrator service.
type TrendConfig struct {
	STSService    awssts.Service
	CostService   awscostexplorer.Service
	Renderer      output.Renderer
	ReportService report.Service
}

// TrendService is the interface for the trend orchestrator service.
type TrendService interface {
	AnalyzeTrends(trendChecks []string, generateReport bool, reportPath string) error
}

// SystemConfig holds the dependencies for the system orchestrator service.
type SystemConfig struct {
	UpdateService update.Service
	Renderer      output.Renderer
	VersionInfo   model.VersionInfo
}

// SystemService is the interface for the system orchestrator service.
type SystemService interface {
	Update() error
	Version() error
	CheckForUpdateInBackground() <-chan model.VersionCheckResult
}
