package csvoutput

import (
	"encoding/csv"
	"errors"
	"os"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
)

// OutputCostComparisonCSV outputs cost comparison data as CSV
func OutputCostComparisonCSV(input model.RenderCostComparisonInput) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	headers := []string{"Service", "Last Month", "Current Month", "Difference"}
	if err := w.Write(headers); err != nil {
		return err
	}

	var rows [][]string

	// Total row
	rows = append(rows, mapTotalRow(input.LastTotalCost, input.CurrentTotalCost))

	// Service breakdown
	if input.LastMonth == nil || input.CurrentMonth == nil {
		return errors.New("both LastMonth and CurrentMonth data must be provided for service breakdown")
	}

	orderedServices := orderCostServices(&input.CurrentMonth.CostGroup)
	for _, service := range orderedServices {
		rows = append(rows, mapServiceRow(*input.LastMonth, service))
	}

	return w.WriteAll(rows)
}

// OutputTrendCSV outputs trend data as CSV
func OutputTrendCSV(monthlyCosts []model.CostInfo, services []string) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	headers := []string{"Period Start", "Period End", "Total Cost", "Unit"}
	if len(services) > 0 {
		headers = append(headers, "Services")
	}

	if err := w.Write(headers); err != nil {
		return err
	}

	rows := mapTrendRows(monthlyCosts, services)

	return w.WriteAll(rows)
}

// OutputWasteCSV outputs waste detection data as CSV
func OutputWasteCSV(input model.RenderWasteInput, pricingSvc pricing.Service) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	headers := []string{"Resource Category", "Resource Identifier", "Estimated Monthly Cost (USD)", "Metric / Size", "Age (Days)", "Additional Details"}
	if err := w.Write(headers); err != nil {
		return err
	}

	var rows [][]string

	rows = append(rows, mapEBSVolumes(input.StoppedVolumes, "stopped", pricingSvc)...)
	rows = append(rows, mapEBSVolumes(input.UnusedVolumes, "unattached", pricingSvc)...)
	rows = append(rows, mapElasticIPs(input.ElasticIPs, pricingSvc)...)
	rows = append(rows, mapS3Buckets(input.S3Buckets)...)
	rows = append(rows, mapS3MultipartUploads(input.S3MultipartUploads)...)
	rows = append(rows, mapStoppedInstances(input.StoppedInstances)...)
	rows = append(rows, mapReservedInstances(input.Ris)...)
	rows = append(rows, mapLoadBalancers(input.LoadBalancers, pricingSvc)...)
	rows = append(rows, mapAMIs(input.UnusedAMIs)...)

	orphaned, stale := mapSnapshots(input.OrphanedSnapshots)
	rows = append(rows, orphaned...)
	rows = append(rows, stale...)

	rows = append(rows, mapKeyPairs(input.UnusedKeyPairs)...)
	rows = append(rows, mapCloudWatchLogGroups(input.CloudWatchLogGroups)...)
	rows = append(rows, mapRDSInstances(input.RDSInstances)...)
	rows = append(rows, mapRDSSnapshots(input.RDSSnapshots)...)
	rows = append(rows, mapRDSIdleInstances(input.RDSIdleInstances)...)
	rows = append(rows, mapNATGateways(input.IdleNATGateways)...)
	rows = append(rows, mapIdleLoadBalancers(input.IdleLoadBalancers)...)
	rows = append(rows, mapLambdaOverProvisioned(input.OverProvisionedLambdas)...)
	rows = append(rows, mapIdleSageMakerEndpoints(input.IdleSageMakerEndpoints)...)
	rows = append(rows, mapECRNoLifecyclePolicies(input.ECRNoLifecyclePolicies)...)
	rows = append(rows, mapECREmptyRepositories(input.ECREmptyRepositories)...)
	rows = append(rows, mapECRUntaggedImages(input.ECRUntaggedImages)...)
	rows = append(rows, mapUnusedSecrets(input.UnusedSecrets, pricingSvc)...)

	return w.WriteAll(rows)
}
