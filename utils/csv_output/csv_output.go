package csvoutput

import (
	"encoding/csv"
	"os"

	"github.com/elC0mpa/aws-doctor/model"
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
func OutputWasteCSV(input model.RenderWasteInput) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	headers := []string{"Resource Category", "Resource Identifier", "Estimated Monthly Cost (USD)", "Metric / Size", "Age (Days)", "Additional Details"}
	if err := w.Write(headers); err != nil {
		return err
	}

	var rows [][]string
	rows = append(rows, mapEBSVolumes(input.StoppedVolumes, "stopped")...)
	rows = append(rows, mapEBSVolumes(input.UnusedVolumes, "unattached")...)
	rows = append(rows, mapElasticIPs(input.ElasticIPs)...)
	rows = append(rows, mapS3Buckets(input.S3Buckets)...)
	rows = append(rows, mapS3MultipartUploads(input.S3MultipartUploads)...)
	rows = append(rows, mapStoppedInstances(input.StoppedInstances)...)
	rows = append(rows, mapReservedInstances(input.Ris)...)
	rows = append(rows, mapLoadBalancers(input.LoadBalancers)...)
	rows = append(rows, mapAMIs(input.UnusedAMIs)...)

	orphaned, stale := mapSnapshots(input.OrphanedSnapshots)
	rows = append(rows, orphaned...)
	rows = append(rows, stale...)

	rows = append(rows, mapKeyPairs(input.UnusedKeyPairs)...)
	rows = append(rows, mapCloudWatchLogGroups(input.CloudWatchLogGroups)...)
	rows = append(rows, mapRDSInstances(input.RDSInstances)...)
	rows = append(rows, mapRDSSnapshots(input.RDSSnapshots)...)
	rows = append(rows, mapRDSIdleInstances(input.RDSIdleInstances)...)

	return w.WriteAll(rows)
}
