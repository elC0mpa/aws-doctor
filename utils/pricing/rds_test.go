package pricing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateRDSInstanceMonthlyCost(t *testing.T) {
	// 100 GB single-AZ: 100 * 0.115 = 11.50
	assert.Equal(t, 11.5, CalculateRDSInstanceMonthlyCost(100, false))
	// 100 GB multi-AZ: 100 * 0.115 * 2 = 23.00
	assert.Equal(t, 23.0, CalculateRDSInstanceMonthlyCost(100, true))
}

func TestCalculateRDSSnapshotMonthlyCost(t *testing.T) {
	// 100 GB: 100 * 0.095 = 9.50
	assert.Equal(t, 9.5, CalculateRDSSnapshotMonthlyCost(100))
}

func TestCalculateRDSIdleInstanceMonthlyCost(t *testing.T) {
	// db.t3.micro (12.41 compute) + 20 GB storage (20 * 0.115 = 2.30) = 14.71
	cost := CalculateRDSIdleInstanceMonthlyCost("db.t3.micro", 20, false)
	assert.Equal(t, 14.71, cost)

	// db.r5.large multi-AZ: (175.20 + 100 * 0.115) * 2 = (175.20 + 11.50) * 2 = 373.40
	costMultiAZ := CalculateRDSIdleInstanceMonthlyCost("db.r5.large", 100, true)
	assert.Equal(t, 373.4, costMultiAZ)

	// Unknown instance type: 0 compute + 50 GB storage = 5.75
	costUnknown := CalculateRDSIdleInstanceMonthlyCost("db.unknown.type", 50, false)
	assert.Equal(t, 5.75, costUnknown)
}

func TestRDSInstanceComputeCost(t *testing.T) {
	assert.Equal(t, 12.41, RDSInstanceComputeCost("db.t3.micro"))
	assert.Equal(t, 175.20, RDSInstanceComputeCost("db.r5.large"))
	assert.Equal(t, 0.0, RDSInstanceComputeCost("db.nonexistent.type"))
}
