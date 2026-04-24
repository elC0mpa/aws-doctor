package outputshared

import (
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
)

func TestResourceRow_ToSlice(t *testing.T) {
	row := ResourceRow{
		Category:      "Compute",
		Identifier:    "i-123",
		EstimatedCost: "$10.00",
		Metric:        "t3.micro",
		Age:           "30",
		Details:       "Stopped",
	}

	expected := []string{"Compute", "i-123", "$10.00", "t3.micro", "30", "Stopped"}
	assert.Equal(t, expected, row.ToSlice())
}

func TestOrderCostServices(t *testing.T) {
	groups := model.CostGroup{
		"EC2":    {Amount: 100.0, Unit: "USD"},
		"S3":     {Amount: 50.0, Unit: "USD"},
		"Lambda": {Amount: 150.0, Unit: "USD"},
	}

	sorted := OrderCostServices(&groups)
	assert.Len(t, sorted, 3)
	assert.Equal(t, "Lambda", sorted[0].Name)
	assert.Equal(t, "EC2", sorted[1].Name)
	assert.Equal(t, "S3", sorted[2].Name)
}

func TestParseCostString(t *testing.T) {
	amount, unit := ParseCostString("10.50 USD")
	assert.Equal(t, 10.50, amount)
	assert.Equal(t, "USD", unit)

	amount, unit = ParseCostString("10.50")
	assert.Equal(t, 10.50, amount)
	assert.Equal(t, "", unit)
}

func TestFormatCost(t *testing.T) {
	assert.Equal(t, "10.50 USD", FormatCost(10.50, "USD"))
	assert.Equal(t, "10.50", FormatCost(10.50, ""))
}
