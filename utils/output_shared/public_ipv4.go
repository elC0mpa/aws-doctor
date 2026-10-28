package outputshared

import (
	"fmt"
	"sort"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
)

// SummarizePublicIPv4 groups in-use public IPv4 addresses by resource type and prices each group.
// The result is sorted by resource type so every output format lists the rows in the same order.
func SummarizePublicIPv4(addresses []model.PublicIPv4Address, pricingSvc pricing.Service) []model.CategorySummary {
	if len(addresses) == 0 {
		return nil
	}

	counts := make(map[string]int)
	for _, a := range addresses {
		counts[a.ResourceType]++
	}

	rate := pricingSvc.CalculatePublicIPv4InUseMonthlyCost()
	summaries := make([]model.CategorySummary, 0, len(counts))

	for resourceType, count := range counts {
		summaries = append(summaries, model.CategorySummary{
			Name:  resourceType,
			Count: count,
			Cost:  float64(count) * rate,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})

	return summaries
}

// TotalPublicIPv4 returns the number of addresses and the estimated monthly cost across all groups.
func TotalPublicIPv4(summaries []model.CategorySummary) (int, float64) {
	var (
		count int
		cost  float64
	)

	for _, s := range summaries {
		count += s.Count
		cost += s.Cost
	}

	return count, cost
}

// PresentPublicIPv4Charge returns a ResourceRow for one resource-type group of in-use public IPv4 addresses.
func PresentPublicIPv4Charge(summary model.CategorySummary) ResourceRow {
	metric := fmt.Sprintf("%d addresses", summary.Count)
	if summary.Count == 1 {
		metric = "1 address"
	}

	return ResourceRow{
		Category:      "Public IPv4 (in use)",
		Identifier:    summary.Name,
		EstimatedCost: fmt.Sprintf("$%.2f", summary.Cost),
		Metric:        metric,
		Age:           NAValue,
		Details:       "Informational: not included in estimated waste",
	}
}
