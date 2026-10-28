package outputshared

import (
	"testing"

	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
)

func TestSummarizePublicIPv4(t *testing.T) {
	m := new(services.MockPricingService)
	m.On("CalculatePublicIPv4InUseMonthlyCost").Return(3.65)

	addresses := []model.PublicIPv4Address{
		{PublicIP: "3.3.3.1", ResourceType: "EC2 Instance"},
		{PublicIP: "3.3.3.2", ResourceType: "NAT Gateway"},
		{PublicIP: "3.3.3.3", ResourceType: "EC2 Instance"},
		{PublicIP: "3.3.3.4", ResourceType: "Load Balancer"},
		{PublicIP: "3.3.3.5", ResourceType: "EC2 Instance"},
		{PublicIP: "3.3.3.6", ResourceType: "Load Balancer"},
	}

	got := SummarizePublicIPv4(addresses, m)

	assert.Equal(t, []model.CategorySummary{
		{Name: "EC2 Instance", Count: 3, Cost: 3 * 3.65},
		{Name: "Load Balancer", Count: 2, Cost: 2 * 3.65},
		{Name: "NAT Gateway", Count: 1, Cost: 3.65},
	}, got)
}

func TestSummarizePublicIPv4_Empty(t *testing.T) {
	assert.Nil(t, SummarizePublicIPv4(nil, new(services.MockPricingService)))
}

func TestTotalPublicIPv4(t *testing.T) {
	count, cost := TotalPublicIPv4([]model.CategorySummary{
		{Name: "EC2 Instance", Count: 3, Cost: 10.95},
		{Name: "NAT Gateway", Count: 1, Cost: 3.65},
	})

	assert.Equal(t, 4, count)
	assert.InDelta(t, 14.60, cost, 0.0001)
}

func TestPresentPublicIPv4Charge(t *testing.T) {
	p := PresentPublicIPv4Charge(model.CategorySummary{Name: "NAT Gateway", Count: 2, Cost: 7.30})

	assert.Equal(t, ResourceRow{
		Category:      "Public IPv4 (in use)",
		Identifier:    "NAT Gateway",
		EstimatedCost: "$7.30",
		Metric:        "2 addresses",
		Age:           NAValue,
		Details:       "Informational: not included in estimated waste",
	}, p)

	assert.Equal(t, "1 address", PresentPublicIPv4Charge(model.CategorySummary{Name: "NAT Gateway", Count: 1, Cost: 3.65}).Metric)
}
