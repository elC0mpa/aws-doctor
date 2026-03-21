package outputshared

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
)

const (
	// NAValue is the default value for missing information
	NAValue = "-"
)

// ResourceRow represents a common structure for resource-based output (CSV/Table)
type ResourceRow struct {
	Category      string
	Identifier    string
	EstimatedCost string
	Metric        string
	Age           string
	Details       string
}

// ToSlice converts the ResourceRow to a string slice
func (r ResourceRow) ToSlice() []string {
	return []string{
		r.Category,
		r.Identifier,
		r.EstimatedCost,
		r.Metric,
		r.Age,
		r.Details,
	}
}

// OrderCostServices sorts service costs by amount in descending order
func OrderCostServices(costGroups *model.CostGroup) []model.ServiceCost {
	sortedServices := make([]model.ServiceCost, 0, len(*costGroups))
	for key, group := range *costGroups {
		sortedServices = append(sortedServices, model.ServiceCost{
			Name:   key,
			Amount: group.Amount,
			Unit:   group.Unit,
		})
	}

	sort.Slice(sortedServices, func(i, j int) bool {
		return sortedServices[i].Amount > sortedServices[j].Amount
	})

	return sortedServices
}

// ParseCostString parses a cost string like "10.50 USD" into amount and unit
func ParseCostString(costStr string) (float64, string) {
	parts := strings.Split(costStr, " ")
	amount, _ := strconv.ParseFloat(parts[0], 64)
	unit := ""
	if len(parts) > 1 {
		unit = parts[1]
	}
	return amount, unit
}

// FormatCost returns a formatted cost string
func FormatCost(amount float64, unit string) string {
	return fmt.Sprintf("%.2f %s", amount, unit)
}
